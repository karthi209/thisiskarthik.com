package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	reloadClients = make(map[chan bool]bool)
	reloadMutex   sync.Mutex
)

func getLocalIPs() []string {
	addrs := []string{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return addrs
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		ifAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range ifAddrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip != nil {
				addrs = append(addrs, ip.String())
			}
		}
	}
	return addrs
}

func rebuildSite() error {
	cmd := exec.Command("go", "run", "generate.go")
	// Suppress output - only show errors
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	
	// Always notify clients, even on error (they can decide what to do)
	reloadMutex.Lock()
	clientCount := len(reloadClients)
	clients := make([]chan bool, 0, clientCount)
	for client := range reloadClients {
		clients = append(clients, client)
	}
	reloadMutex.Unlock()
	
	// Notify all clients
	for _, client := range clients {
		select {
		case client <- true:
		default:
		}
	}
	
	return err
}

func watchFiles() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Watch directories
	dirs := []string{"templates", "content", "static"}
	for _, dir := range dirs {
		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return watcher.Add(path)
			}
			return nil
		}); err != nil {
			watcher.Close()
			return nil, err
		}
	}

	// Watch Go files in root
	if files, err := filepath.Glob("*.go"); err == nil {
		for _, file := range files {
			watcher.Add(file)
		}
	}

	return watcher, nil
}

func main() {
	outputDir := "public"
	port := "5173"

	// Initial build (silent)
	if err := rebuildSite(); err != nil {
		fmt.Printf("✗ Build failed: %v\n", err)
		os.Exit(1)
	}

	// Check if output directory exists
	info, err := os.Stat(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("The repository is empty. Build failed.")
			os.Exit(1)
		}
		fmt.Printf("I could not access the repository: %v\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Printf("'%s' is not a directory.\n", outputDir)
		os.Exit(1)
	}

	// Check if directory is readable
	if _, err := os.ReadDir(outputDir); err != nil {
		fmt.Printf("I could not read the repository: %v\n", err)
		os.Exit(1)
	}

	// Check if port is available
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Port %s is occupied. Another process may be using it.\n", port)
		os.Exit(1)
	}
	listener.Close()

	// Minimal output (Vite-style)
	fmt.Printf("\n  ➜  Local:   http://localhost:%s\n", port)
	if addrs := getLocalIPs(); len(addrs) > 0 {
		fmt.Printf("  ➜  Network: http://%s:%s\n", addrs[0], port)
	}
	fmt.Println()

	// Configure server for performance
	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Reload endpoint for browser auto-refresh
	http.HandleFunc("/__reload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		clientChan := make(chan bool, 10)
		reloadMutex.Lock()
		reloadClients[clientChan] = true
		reloadMutex.Unlock()

		defer func() {
			reloadMutex.Lock()
			delete(reloadClients, clientChan)
			reloadMutex.Unlock()
			close(clientChan)
		}()

		// Send initial connection message
		fmt.Fprintf(w, "data: connected\n\n")
		flusher.Flush()

		// Keep connection alive and listen for reload signals
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-clientChan:
				// Reload signal received
				fmt.Fprintf(w, "data: reload\n\n")
				flusher.Flush()
				return
			case <-ticker.C:
				// Send keepalive ping
				fmt.Fprintf(w, ": ping\n\n")
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})

	// Serve static files with auto-reload script injection
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Resolve file path
		requestPath := r.URL.Path
		if requestPath == "/" {
			requestPath = "/index.html"
		} else if strings.HasSuffix(requestPath, "/") {
			requestPath = requestPath + "index.html"
		} else if !strings.HasSuffix(requestPath, ".html") {
			// Try adding index.html for directory paths
			dirPath := filepath.Join(outputDir, requestPath)
			if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
				requestPath = requestPath + "/index.html"
			}
		}

		filePath := filepath.Join(outputDir, requestPath)
		
		// Check if it's an HTML file
		if strings.HasSuffix(requestPath, ".html") {
			// Check if file exists
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				// Read the file
				content, err := os.ReadFile(filePath)
				if err != nil {
					http.Error(w, "Error reading file", http.StatusInternalServerError)
					return
				}

				// Inject reload script before </body> or at end of file
				htmlContent := string(content)
				reloadScript := `<script>
(function() {
  if (typeof EventSource !== 'undefined') {
    function connect() {
      var source = new EventSource('/__reload');
      source.onmessage = function(e) {
        if (e.data === 'reload') {
          source.close();
          window.location.reload();
        }
      };
      source.onerror = function() {
        source.close();
        // Reconnect after 1 second
        setTimeout(connect, 1000);
      };
    }
    connect();
  }
})();
</script>`

				// Insert before </body> if it exists, otherwise at the end
				if strings.Contains(htmlContent, "</body>") {
					htmlContent = strings.Replace(htmlContent, "</body>", reloadScript+"</body>", 1)
				} else {
					htmlContent += reloadScript
				}

				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(htmlContent))
				return
			}
		}

		// For non-HTML files, serve normally
		fs := http.FileServer(http.Dir(outputDir))
		fs.ServeHTTP(w, r)
	})

	// Setup file watcher
	watcher, err := watchFiles()
	if err != nil {
		// Silent failure - server will run without auto-rebuild
	} else {
		defer watcher.Close()

		// Watch for file changes and rebuild
		go func() {
			var rebuildTimer *time.Timer
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					// Skip events for files in public directory
					if strings.HasPrefix(event.Name, outputDir) {
						continue
					}
					
					// If a new directory is created, add it to the watcher
					if event.Op&fsnotify.Create == fsnotify.Create {
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							watcher.Add(event.Name)
						}
					}
					
					// Ignore chmod events and only watch for write/create/remove
					if event.Op&fsnotify.Write == fsnotify.Write ||
						event.Op&fsnotify.Create == fsnotify.Create ||
						event.Op&fsnotify.Remove == fsnotify.Remove {
						// Debounce: wait 300ms before rebuilding
						if rebuildTimer != nil {
							rebuildTimer.Stop()
						}
						rebuildTimer = time.AfterFunc(300*time.Millisecond, func() {
							if err := rebuildSite(); err != nil {
								fmt.Printf("  ✗ rebuild failed: %v\n", err)
							} else {
								// Get the filename for the reload message
								fileName := filepath.Base(event.Name)
								fmt.Printf("  ➜  page reload %s\n", fileName)
							}
						})
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					// Silent watcher errors
					_ = err
				}
			}
		}()
	}

	// Handle graceful shutdown on interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("The server encountered difficulty: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
		os.Exit(1)
	}
	// Exit cleanly on successful shutdown
	os.Exit(0)
}
