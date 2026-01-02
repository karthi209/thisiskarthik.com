package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"

	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

// Configuration
var (
	contentDir      = "content"
	postsDir        = filepath.Join(contentDir, "posts")
	imagesDir       = filepath.Join(contentDir, "images")
	outputDir       = "public"
	templatesDir    = "templates"
	staticDir       = "static"
	publicImagesDir = filepath.Join(outputDir, "images")
	basePath        = getBasePath()
)

// getBasePath returns the base path for assets and links
// Reads from BASE_PATH environment variable, defaults to "/"
// For GitHub Pages project sites, set BASE_PATH="/repo-name/"
func getBasePath() string {
	path := os.Getenv("BASE_PATH")
	if path == "" {
		return "/"
	}
	// Ensure it starts and ends with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return path
}

// Post represents a writing
type Post struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Category  string `json:"category"`
	Slug      string `json:"slug"`
	IsDraft   bool   `json:"is_draft"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Frontmatter represents the YAML frontmatter in markdown files
type Frontmatter struct {
	Title    string `yaml:"title"`
	Category string `yaml:"category"`
	Date     string `yaml:"date"`
	Slug     string `yaml:"slug"`
	IsDraft  bool   `yaml:"is_draft"`
}

// Template data structures
type HomePageData struct {
	PageType string
	Title    string
	BasePath string
}

type WritingsPageData struct {
	PageType        string
	Title           string
	BasePath        string
	Writings        []PostTemplateData
	GroupedWritings []YearGroup
}

type YearGroup struct {
	Year  string
	Count int
	Posts []PostTemplateData
}

type PostTemplateData struct {
	Title         string
	Slug          string
	DateLabel     string
	DateLabelFormal string
	Year          int
	Category      string
	CategoryUpper string
	Content       template.HTML
	ReadingTime   int
	IsDraft       bool
	CreatedAt     time.Time
}

type PostPageData struct {
	PageType string
	Title    string
	BasePath string
	Post     PostTemplateData
}

type AboutPageData struct {
	PageType string
	Title    string
	BasePath string
}

type ColophonPageData struct {
	PageType    string
	Title       string
	BasePath    string
	LinesOfCode int
	PageCount   int
	ImageCount  int
	BuildYear   int
	BuildTime   string
}

type IndexPageData struct {
	PageType        string
	Title           string
	BasePath        string
	Writings        []PostTemplateData
	GroupedWritings []YearGroup
	WritingCount    int
}

// plural returns "s" if count is not 1, empty string otherwise
func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// validateDirectories checks that required directories exist
func validateDirectories() error {
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return fmt.Errorf("templates directory not found: %s", templatesDir)
	}
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		return fmt.Errorf("static directory not found: %s", staticDir)
	}
	return nil
}

func main() {
	buildStart := time.Now()
	fmt.Println("Commencing the typesetting of pages...")

	// Validate directories exist
	if err := validateDirectories(); err != nil {
		fmt.Printf("I could not proceed: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists and is writable
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("I could not prepare the workspace: %v\n", err)
		os.Exit(1)
	}

	// Load templates
	templates, err := loadTemplates()
	if err != nil {
		fmt.Printf("I encountered difficulty in loading the templates: %v\n", err)
		os.Exit(1)
	}

	// Find all markdown files
	markdownFiles, err := findMarkdownFiles(postsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("A note: difficulty in scanning the repository: %v\n", err)
		}
		markdownFiles = []string{}
	}

	// Filter out README files and invalid paths
	var postFiles []string
	for _, file := range markdownFiles {
		if strings.Contains(file, "README.md") || strings.Contains(filepath.Base(file), "README") {
			continue
		}
		if _, err := os.Stat(file); err == nil {
			postFiles = append(postFiles, file)
		}
	}

	if len(postFiles) > 0 {
		fmt.Printf("   Discovered %d manuscript%s\n", len(postFiles), plural(len(postFiles)))
	}

	// Process all posts
	var posts []Post
	for _, filePath := range postFiles {
		post, err := processPostFile(filePath)
		if err != nil {
			continue // Silently skip invalid files
		}
		if post != nil && post.Title != "" {
			posts = append(posts, *post)
		}
	}

	// Sort by created_at descending (newest first)
	sort.Slice(posts, func(i, j int) bool {
		dateI, _ := time.Parse(time.RFC3339, posts[i].CreatedAt)
		dateJ, _ := time.Parse(time.RFC3339, posts[j].CreatedAt)
		return dateI.After(dateJ)
	})

	if len(posts) > 0 {
		fmt.Printf("   Committed %d entr%s to print\n", len(posts), plural(len(posts)))
	}

	// Convert posts to template data
	postTemplateData := make([]PostTemplateData, 0, len(posts))
	for _, post := range posts {
		if post.IsDraft {
			continue // Skip drafts in static site
		}
		createdAt, _ := time.Parse(time.RFC3339, post.CreatedAt)
		dateLabel := formatDate(post.CreatedAt)
		dateLabelFormal := formatDateFormal(post.CreatedAt)
		readingTime := calculateReadingTime(post.Content)
		postTemplateData = append(postTemplateData, PostTemplateData{
			Title:         post.Title,
			Slug:          post.Slug,
			DateLabel:     dateLabel,
			DateLabelFormal: dateLabelFormal,
			Year:          createdAt.Year(),
			Category:      post.Category,
			CategoryUpper: strings.ToUpper(post.Category),
			Content:       template.HTML(post.Content),
			ReadingTime:   readingTime,
			IsDraft:       post.IsDraft,
			CreatedAt:     createdAt,
		})
	}

	// Group posts by year
	groupedWritings := groupPostsByYear(postTemplateData)

	// Generate pages
	if err := generateHomePage(templates, postTemplateData); err != nil {
		fmt.Printf("I encountered difficulty in composing the home page: %v\n", err)
	}

	if err := generateWritingsPage(templates, postTemplateData, groupedWritings); err != nil {
		fmt.Printf("I encountered difficulty in composing the writings page: %v\n", err)
	}

	for _, post := range postTemplateData {
		if err := generatePostPage(templates, post); err != nil {
			continue // Skip failed pages silently
		}
	}

	if err := generateAboutPage(templates); err != nil {
		fmt.Printf("I encountered difficulty in composing the about page: %v\n", err)
	}

	buildDuration := time.Since(buildStart)
	if err := generateColophonPage(templates, buildDuration); err != nil {
		fmt.Printf("I encountered difficulty in composing the colophon page: %v\n", err)
	}

	if err := generateIndexPage(templates, postTemplateData, groupedWritings); err != nil {
		fmt.Printf("I encountered difficulty in composing the index page: %v\n", err)
	}

	// Copy static files
	if err := copyStaticFiles(); err != nil {
		fmt.Printf("A note: difficulty in gathering materials: %v\n", err)
	}

	// Copy images (non-critical, continue on error)
	_ = copyImages()

	// Minimal completion message
	if len(postTemplateData) > 0 {
		fmt.Printf("\nThe volume is complete. %d entr%s rendered to %s.\n", len(postTemplateData), plural(len(postTemplateData)), outputDir)
	} else {
		fmt.Printf("\nThe volume is complete. Rendered to %s.\n", outputDir)
	}
}

func loadTemplates() (*template.Template, error) {
	tmpl := template.New("base")

	// Load base template first
	basePath := filepath.Join(templatesDir, "base.html")
	baseContent, err := os.ReadFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("difficulty in reading the base template: %w", err)
	}

	if len(baseContent) == 0 {
		return nil, fmt.Errorf("base template is empty")
	}

	tmpl, err = tmpl.Parse(string(baseContent))
	if err != nil {
		return nil, fmt.Errorf("difficulty in parsing the base template: %w", err)
	}

	// Load all other template files
	files, err := filepath.Glob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("difficulty in finding templates: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no template files found in %s", templatesDir)
	}

	// Parse all templates
	for _, file := range files {
		if filepath.Base(file) == "base.html" {
			continue // Already parsed
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("difficulty in reading template %s: %w", file, err)
		}
		if len(content) == 0 {
			continue // Skip empty templates
		}
		_, err = tmpl.New(filepath.Base(file)).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("difficulty in parsing template %s: %w", file, err)
		}
	}

	return tmpl, nil
}

func generateHomePage(templates *template.Template, posts []PostTemplateData) error {
	data := HomePageData{
		PageType: "home",
		Title:    "Home",
		BasePath: basePath,
	}

	return writeTemplate(templates, "home.html", filepath.Join(outputDir, "index.html"), data)
}

func generateWritingsPage(templates *template.Template, posts []PostTemplateData, grouped []YearGroup) error {
	data := WritingsPageData{
		PageType:        "writings",
		Title:           "Writings",
		BasePath:        basePath,
		Writings:        posts,
		GroupedWritings: grouped,
	}

	return writeTemplate(templates, "writings.html", filepath.Join(outputDir, "writings", "index.html"), data)
}

func generatePostPage(templates *template.Template, post PostTemplateData) error {
	// Create writings/{slug}/index.html structure
	postDir := filepath.Join(outputDir, "writings", post.Slug)
	if err := os.MkdirAll(postDir, 0755); err != nil {
		return err
	}

	data := PostPageData{
		PageType: "post",
		Title:    post.Title,
		BasePath: basePath,
		Post:     post,
	}

	return writeTemplate(templates, "post.html", filepath.Join(postDir, "index.html"), data)
}

func generateAboutPage(templates *template.Template) error {
	data := AboutPageData{
		PageType: "about",
		Title:    "About",
		BasePath: basePath,
	}

	aboutDir := filepath.Join(outputDir, "about")
	if err := os.MkdirAll(aboutDir, 0755); err != nil {
		return err
	}

	return writeTemplate(templates, "about.html", filepath.Join(aboutDir, "index.html"), data)
}

func generateColophonPage(templates *template.Template, buildDuration time.Duration) error {
	buildYear := time.Now().Year()
	buildTimeMs := buildDuration.Milliseconds()
	buildTimeStr := fmt.Sprintf("%d", buildTimeMs)

	data := ColophonPageData{
		PageType:    "colophon",
		Title:       "Colophon",
		BasePath:    basePath,
		LinesOfCode: 0,
		PageCount:   0,
		ImageCount:  0,
		BuildYear:   buildYear,
		BuildTime:   buildTimeStr,
	}

	colophonDir := filepath.Join(outputDir, "colophon")
	if err := os.MkdirAll(colophonDir, 0755); err != nil {
		return err
	}

	return writeTemplate(templates, "colophon.html", filepath.Join(colophonDir, "index.html"), data)
}

// countLinesOfCode counts lines in core site code only:
// - generate.go (site generator)
// - CSS files (styling)
// - Template HTML files (structure)
// Excludes: serve.go (dev tool), scripts (build tools), content (posts)
func countLinesOfCode() int {
	total := 0

	// Count generate.go only (exclude serve.go - it's a dev tool)
	if content, err := os.ReadFile("generate.go"); err == nil {
		total += strings.Count(string(content), "\n")
	}

	// Count CSS files in static/css
	err := filepath.WalkDir("static/css", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".css") {
			if content, err := os.ReadFile(path); err == nil {
				total += strings.Count(string(content), "\n")
			}
		}
		return nil
	})
	if err != nil {
		// Continue if CSS directory doesn't exist
	}

	// Count template HTML files
	err = filepath.WalkDir("templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".html") {
			if content, err := os.ReadFile(path); err == nil {
				total += strings.Count(string(content), "\n")
			}
		}
		return nil
	})
	if err != nil {
		// Continue if templates directory doesn't exist
	}

	return total
}

// countGeneratedPages counts HTML pages in the output directory
// Note: This is called before colophon, notes, library, and index pages are generated
// So we count existing pages + pages that will be generated
func countGeneratedPages() int {
	count := 0

	// Count pages already generated (home, writings, posts, about)
	err := filepath.WalkDir(outputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() && strings.HasSuffix(path, "index.html") {
			count++
		}

		return nil
	})

	if err != nil {
		// If outputDir doesn't exist yet, start with 0
		count = 0
	}

	// Add pages that will be generated after colophon:
	// - Colophon (1)
	// - Index (1)
	count += 2 // Colophon + Index

	return count
}

// countImages counts image files in static/images
func countImages() int {
	count := 0
	imageDir := filepath.Join(staticDir, "images")

	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"}

	err := filepath.WalkDir(imageDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			lowerPath := strings.ToLower(path)
			for _, ext := range imageExts {
				if strings.HasSuffix(lowerPath, ext) {
					count++
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return 0
	}

	return count
}

func generateIndexPage(templates *template.Template, posts []PostTemplateData, grouped []YearGroup) error {
	data := IndexPageData{
		PageType:        "index",
		Title:           "Index",
		BasePath:        basePath,
		Writings:        posts,
		GroupedWritings: grouped,
		WritingCount:    len(posts),
	}

	indexDir := filepath.Join(outputDir, "index")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return err
	}

	return writeTemplate(templates, "index.html", filepath.Join(indexDir, "index.html"), data)
}

func writeTemplate(templates *template.Template, templateName, outputPath string, data interface{}) error {
	// Validate template exists
	if templates.Lookup(templateName) == nil {
		return fmt.Errorf("template %s not found", templateName)
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != outputDir {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("could not create directory: %w", err)
		}
	}

	// Create file with buffer for performance
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := templates.ExecuteTemplate(file, templateName, data); err != nil {
		os.Remove(outputPath) // Clean up on error
		return fmt.Errorf("template execution failed: %w", err)
	}

	return nil
}

func groupPostsByYear(posts []PostTemplateData) []YearGroup {
	groups := make(map[int][]PostTemplateData)
	for _, post := range posts {
		year := post.Year
		groups[year] = append(groups[year], post)
	}

	var result []YearGroup
	for year, posts := range groups {
		result = append(result, YearGroup{
			Year:  fmt.Sprintf("%d", year),
			Count: len(posts),
			Posts: posts,
		})
	}

	// Sort by year descending
	sort.Slice(result, func(i, j int) bool {
		yearI, _ := strconv.Atoi(result[i].Year)
		yearJ, _ := strconv.Atoi(result[j].Year)
		return yearI > yearJ
	})

	return result
}

func formatDate(dateStr string) string {
	if dateStr == "" {
		return "—"
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "—"
	}

	return t.Format("Jan 2")
}

func formatDateFormal(dateStr string) string {
	if dateStr == "" {
		return "—"
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "—"
	}

	// Format in formal vintage book style: "the 15th of May, 2025"
	day := t.Day()
	month := t.Format("January")
	year := t.Year()
	
	// Add ordinal suffix for day (1st, 2nd, 3rd, 4th, etc.)
	var suffix string
	switch day {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	default:
		suffix = "th"
	}
	
	return fmt.Sprintf("the %d%s of %s, %d", day, suffix, month, year)
}

func calculateReadingTime(content string) int {
	// Remove HTML tags
	text := strings.ReplaceAll(content, "<", " ")
	text = strings.ReplaceAll(text, ">", " ")
	words := strings.Fields(text)
	wordCount := len(words)
	readingTime := wordCount / 200
	if readingTime < 1 {
		return 1
	}
	return readingTime
}

// Copy functions from original generate-static.go
func findMarkdownFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func processPostFile(filePath string) (*Post, error) {
	// Validate file exists and is readable
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not accessible: %w", err)
	}
	if info.Size() == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("difficulty in reading the manuscript: %w", err)
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("file contains no content")
	}

	// Parse frontmatter
	var frontmatter Frontmatter
	rest := content

	// Check for frontmatter delimiter
	if strings.HasPrefix(string(content), "---") {
		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) >= 3 {
			if len(parts[1]) == 0 {
				return nil, fmt.Errorf("frontmatter is empty")
			}
			if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err != nil {
				return nil, fmt.Errorf("difficulty in parsing the frontmatter: %w", err)
			}
			rest = []byte(strings.TrimSpace(parts[2]))
			if len(rest) == 0 {
				return nil, fmt.Errorf("manuscript has no content after frontmatter")
			}
		}
	}

	// Required fields
	if frontmatter.Title == "" {
		return nil, fmt.Errorf("the manuscript lacks a title; it shall be omitted")
	}

	// Validate title is not just whitespace
	if strings.TrimSpace(frontmatter.Title) == "" {
		return nil, fmt.Errorf("title is empty")
	}

	// Generate slug if not provided
	slug := frontmatter.Slug
	if slug == "" {
		slug = generateSlug(frontmatter.Title)
	}
	if slug == "" {
		return nil, fmt.Errorf("could not generate a valid slug from title")
	}

	// Parse date
	var createdAt time.Time
	if frontmatter.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", frontmatter.Date)
		if err != nil {
			parsedDate, err = time.Parse(time.RFC3339, frontmatter.Date)
			if err != nil {
				info, _ := os.Stat(filePath)
				if info != nil {
					createdAt = info.ModTime()
				} else {
					createdAt = time.Now()
				}
			} else {
				createdAt = parsedDate
			}
		} else {
			createdAt = parsedDate
		}
	} else {
		info, err := os.Stat(filePath)
		if err == nil {
			createdAt = info.ModTime()
		} else {
			createdAt = time.Now()
		}
	}

	// Process markdown content to HTML
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var htmlContent strings.Builder
	if err := md.Convert(rest, &htmlContent); err != nil {
		return nil, fmt.Errorf("difficulty in converting the manuscript to print: %w", err)
	}

	// Post-process HTML to add lazy loading to images
	htmlStr := htmlContent.String()
	htmlStr = strings.ReplaceAll(htmlStr, "<img ", "<img loading=\"lazy\" decoding=\"async\" ")

	// Build post object
	post := Post{
		ID:        slug,
		Title:     frontmatter.Title,
		Content:   htmlStr,
		Category:  frontmatter.Category,
		Slug:      slug,
		IsDraft:   frontmatter.IsDraft,
		CreatedAt: createdAt.Format(time.RFC3339),
		UpdatedAt: createdAt.Format(time.RFC3339),
	}

	if post.Category == "" {
		post.Category = "life"
	}

	return &post, nil
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.TrimSpace(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	return slug
}

func copyStaticFiles() error {
	// Validate static directory exists
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		return fmt.Errorf("static directory not found: %s", staticDir)
	}

	// Copy all static files (including styles)
	err := filepath.Walk(staticDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(staticDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(outputDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		srcData, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, srcData, 0644)
	})

	return err
}

func copyDir(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // Source doesn't exist, skip silently
	}

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// Skip if source is empty
		if info.Size() == 0 {
			return nil
		}

		srcData, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Ensure destination directory exists
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		return os.WriteFile(dstPath, srcData, 0644)
	})
}

func copyImages() error {
	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		return fmt.Errorf("no repository of illustrations found; proceeding without")
	}

	if err := os.MkdirAll(publicImagesDir, 0755); err != nil {
		return fmt.Errorf("difficulty in preparing the illustration repository: %w", err)
	}

	var copied int
	err := filepath.WalkDir(imagesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
		isImage := false
		for _, imgExt := range imageExts {
			if ext == imgExt {
				isImage = true
				break
			}
		}

		if !isImage {
			return nil
		}

		relPath, err := filepath.Rel(imagesDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(publicImagesDir, relPath)

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		srcData, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(destPath, srcData, 0644); err != nil {
			return err
		}

		copied++
		return nil
	})

	if err != nil {
		return fmt.Errorf("difficulty in gathering illustrations: %w", err)
	}

	fmt.Printf("Gathered %d illustration%s\n", copied, plural(copied))
	return nil
}
