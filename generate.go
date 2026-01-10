package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

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
	PageType        string
	Title           string
	BasePath        string
	Writings        []PostTemplateData
	GroupedWritings []YearGroup
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
	Title           string
	Slug            string
	DateLabel       string
	DateLabelFormal string
	Year            int
	Category        string
	CategoryUpper   string
	Content         template.HTML
	ReadingTime     int
	IsDraft         bool
	CreatedAt       time.Time
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

type MetaPageData struct {
	PageType  string
	Title     string
	BasePath  string
	BuildYear int
	BuildTime string
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
	fmt.Println("▓▓ SITE GENERATOR V1.0")
	fmt.Println("▓▓ INITIALIZING...")
	fmt.Println()

	// Validate directories exist
	if err := validateDirectories(); err != nil {
		fmt.Printf("▓▓ ERROR: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists and is writable
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("▓▓ ERROR: cannot prepare workspace: %v\n", err)
		os.Exit(1)
	}

	// Load templates
	templates, err := loadTemplates()
	if err != nil {
		fmt.Printf("▓▓ ERROR: template load failed: %v\n", err)
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
		fmt.Printf("▓▓ LOADING %d POST%s...\n", len(postFiles), strings.ToUpper(plural(len(postFiles))))
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
		fmt.Printf("▓▓ PROCESSED %d POST%s\n", len(posts), strings.ToUpper(plural(len(posts))))
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
			Title:           post.Title,
			Slug:            post.Slug,
			DateLabel:       dateLabel,
			DateLabelFormal: dateLabelFormal,
			Year:            createdAt.Year(),
			Category:        post.Category,
			CategoryUpper:   strings.ToUpper(post.Category),
			Content:         template.HTML(post.Content),
			ReadingTime:     readingTime,
			IsDraft:         post.IsDraft,
			CreatedAt:       createdAt,
		})
	}

	// Group posts by year
	groupedWritings := groupPostsByYear(postTemplateData)

	// Generate pages
	fmt.Println("▓▓ GENERATING PAGES...")
	if err := generateHomePage(templates, postTemplateData, groupedWritings); err != nil {
		fmt.Printf("▓▓ ERROR: home page failed: %v\n", err)
	}

	if err := generateWritingsPage(templates, postTemplateData, groupedWritings); err != nil {
		fmt.Printf("▓▓ ERROR: writings page failed: %v\n", err)
	}

	for _, post := range postTemplateData {
		if err := generatePostPage(templates, post); err != nil {
			continue // Skip failed pages silently
		}
	}

	if err := generateAboutPage(templates); err != nil {
		fmt.Printf("▓▓ ERROR: about page failed: %v\n", err)
	}

	if err := generateForAIPage(templates); err != nil {
		fmt.Printf("▓▓ ERROR: forai page failed: %v\n", err)
	}

	buildDuration := time.Since(buildStart)
	if err := generateMetaPage(templates, buildDuration); err != nil {
		fmt.Printf("▓▓ ERROR: meta page failed: %v\n", err)
	}

	if err := generateRSSFeed(postTemplateData); err != nil {
		fmt.Printf("▓▓ ERROR: RSS feed failed: %v\n", err)
	}

	// Copy static files
	fmt.Println("▓▓ COPYING ASSETS...")
	if err := copyStaticFiles(); err != nil {
		fmt.Printf("▓▓ WARNING: asset copy failed: %v\n", err)
	}

	// Copy images (non-critical, continue on error)
	_ = copyImages()

	// Completion message
	fmt.Println()
	if len(postTemplateData) > 0 {
		fmt.Printf("▓▓ BUILD COMPLETE: %d POST%s → %s/\n", len(postTemplateData), strings.ToUpper(plural(len(postTemplateData))), outputDir)
	} else {
		fmt.Printf("▓▓ BUILD COMPLETE → %s/\n", outputDir)
	}
	fmt.Printf("▓▓ TIME: %dms\n", buildDuration.Milliseconds())
	fmt.Println()
}

func loadTemplates() (*template.Template, error) {
	tmpl := template.New("base")

	// Load all template files
	files, err := filepath.Glob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("difficulty in finding templates: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no template files found in %s", templatesDir)
	}

	// Parse all templates
	for _, file := range files {
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

func generateHomePage(templates *template.Template, posts []PostTemplateData, grouped []YearGroup) error {
	data := HomePageData{
		PageType:        "home",
		Title:           "Home",
		BasePath:        basePath,
		Writings:        posts,
		GroupedWritings: grouped,
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

func generateForAIPage(templates *template.Template) error {
	data := AboutPageData{
		PageType: "forai",
		Title:    "For AI",
		BasePath: basePath,
	}

	foraiDir := filepath.Join(outputDir, "forai")
	if err := os.MkdirAll(foraiDir, 0755); err != nil {
		return err
	}

	return writeTemplate(templates, "forai.html", filepath.Join(foraiDir, "index.html"), data)
}

func generateMetaPage(templates *template.Template, buildDuration time.Duration) error {
	buildYear := time.Now().Year()
	buildTimeMs := buildDuration.Milliseconds()
	buildTimeStr := fmt.Sprintf("%d", buildTimeMs)

	data := MetaPageData{
		PageType:  "meta",
		Title:     "Meta",
		BasePath:  basePath,
		BuildYear: buildYear,
		BuildTime: buildTimeStr,
	}

	metaDir := filepath.Join(outputDir, "meta")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return err
	}

	return writeTemplate(templates, "meta.html", filepath.Join(metaDir, "index.html"), data)
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

	// Post-process HTML to add lazy loading to images and fix image paths
	htmlStr := htmlContent.String()
	htmlStr = strings.ReplaceAll(htmlStr, "<img ", "<img loading=\"lazy\" decoding=\"async\" ")

	// Fix image src paths to include base path (for GitHub Pages compatibility)
	// Match src="/images/..." or src='/images/...' and prepend basePath
	if basePath != "/" {
		// Match src="/images/..." or src='/images/...'
		imgSrcRegex := regexp.MustCompile(`src=["'](/images/[^"']+)["']`)
		htmlStr = imgSrcRegex.ReplaceAllStringFunc(htmlStr, func(match string) string {
			// Extract the path
			pathMatch := regexp.MustCompile(`src=["'](/images/[^"']+)["']`)
			submatches := pathMatch.FindStringSubmatch(match)
			if len(submatches) > 1 {
				// Prepend basePath (which already ends with /)
				newPath := basePath + strings.TrimPrefix(submatches[1], "/")
				// Preserve the original quote style
				quote := ""
				if strings.Contains(match, `"`) {
					quote = `"`
				} else {
					quote = `'`
				}
				return `src=` + quote + newPath + quote
			}
			return match
		})
	}

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

	if copied > 0 {
		fmt.Printf("▓▓ COPIED %d IMAGE%s\n", copied, strings.ToUpper(plural(copied)))
	}
	return nil
}

func generateRSSFeed(posts []PostTemplateData) error {
	if len(posts) == 0 {
		return nil // No posts, skip RSS generation
	}

	rssPath := filepath.Join(outputDir, "rss.xml")
	file, err := os.Create(rssPath)
	if err != nil {
		return fmt.Errorf("could not create RSS file: %w", err)
	}
	defer file.Close()

	// Get site URL from environment variable or use default
	siteURL := os.Getenv("SITE_URL")
	if siteURL == "" {
		siteURL = "https://thisiskarthik.com"
		if basePath != "/" {
			// For GitHub Pages project sites, construct URL from basePath
			path := strings.Trim(basePath, "/")
			if path != "" {
				// Try to extract username from path or use default
				username := os.Getenv("GITHUB_USERNAME")
				if username == "" {
					username = "karthi209"
				}
				siteURL = fmt.Sprintf("https://%s.github.io/%s", username, path)
			}
		}
	}
	// Ensure siteURL doesn't end with /
	siteURL = strings.TrimSuffix(siteURL, "/")

	// Get current time for feed date
	now := time.Now().UTC().Format(time.RFC1123Z)

	// Write RSS header
	rssLink := fmt.Sprintf("%s%s", siteURL, basePath)
	rssLink = strings.TrimSuffix(rssLink, "/")
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
<title>The Book of Odds and Ends</title>
<link>%s</link>
<description>Writings and observations by Karthik</description>
<language>en-us</language>
<lastBuildDate>%s</lastBuildDate>
<atom:link href="%srss.xml" rel="self" type="application/rss+xml"/>
`, rssLink, now, rssLink)

	// Write RSS items (limit to 20 most recent)
	maxItems := 20
	if len(posts) < maxItems {
		maxItems = len(posts)
	}

	for i := 0; i < maxItems; i++ {
		post := posts[i]
		postURL := fmt.Sprintf("%s%swritings/%s", siteURL, basePath, post.Slug)

		// Use CreatedAt time directly
		pubDate := post.CreatedAt.UTC().Format(time.RFC1123Z)

		// Clean HTML content for description (strip tags, limit length)
		// Simple HTML tag removal
		description := string(post.Content)
		// Remove HTML tags
		for {
			start := strings.Index(description, "<")
			if start == -1 {
				break
			}
			end := strings.Index(description[start:], ">")
			if end == -1 {
				break
			}
			description = description[:start] + " " + description[start+end+1:]
		}
		// Clean up whitespace
		description = strings.TrimSpace(description)
		// Replace HTML entities
		description = strings.ReplaceAll(description, "&nbsp;", " ")
		description = strings.ReplaceAll(description, "&amp;", "&")
		description = strings.ReplaceAll(description, "&lt;", "<")
		description = strings.ReplaceAll(description, "&gt;", ">")
		// Limit length
		if len(description) > 500 {
			description = description[:500] + "..."
		}

		fmt.Fprintf(file, `<item>
<title><![CDATA[%s]]></title>
<link>%s</link>
<guid isPermaLink="true">%s</guid>
<pubDate>%s</pubDate>
<description><![CDATA[%s]]></description>
</item>
`, post.Title, postURL, postURL, pubDate, description)
	}

	// Write RSS footer
	fmt.Fprintf(file, `</channel>
</rss>`)

	return nil
}
