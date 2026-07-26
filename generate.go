package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"math"
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
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

// Configuration
var (
	contentDir      = "content"
	postsDir        = filepath.Join(contentDir, "posts")
	journalDir      = filepath.Join(contentDir, "journal")
	libraryDir      = filepath.Join(contentDir, "library")
	photosDir       = filepath.Join(contentDir, "photos")
	imagesDir       = filepath.Join(contentDir, "images")
	outputDir       = "public"
	templatesDir    = "templates"
	staticDir       = "static"
	publicImagesDir = filepath.Join(outputDir, "images")
	basePath        = getBasePath()
)

type LibrarySection struct {
	Key          string
	Title        string
	Verb         string
	Description  string
	CountOne     string
	CountMany    string
	ArchiveLabel string
}

var librarySections = []LibrarySection{
	{
		Key:          "films",
		Title:        "films",
		Verb:         "watched",
		Description:  "stories that stayed after the lights came back on.",
		CountOne:     "film remembered",
		CountMany:    "films remembered",
		ArchiveLabel: "films remembered",
	},
	{
		Key:          "tv",
		Title:        "tv series",
		Verb:         "watched",
		Description:  "longer stories, lived with for a while.",
		CountOne:     "series",
		CountMany:    "series",
		ArchiveLabel: "series remembered",
	},
	{
		Key:          "games",
		Title:        "games",
		Verb:         "played",
		Description:  "worlds explored, finished or left unfinished.",
		CountOne:     "game",
		CountMany:    "games",
		ArchiveLabel: "worlds explored",
	},
	{
		Key:          "books",
		Title:        "books",
		Verb:         "read",
		Description:  "pages that remained after the book was closed.",
		CountOne:     "book",
		CountMany:    "books",
		ArchiveLabel: "books remembered",
	},
	{
		Key:          "music",
		Title:        "music",
		Verb:         "listened",
		Description:  "sounds returned to across different parts of life.",
		CountOne:     "music entry",
		CountMany:    "music entries",
		ArchiveLabel: "sounds remembered",
	},
}

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

// Post represents an essay stored in content/posts.
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
	Title    string
	Category string
	Date     string
	Slug     string
	IsDraft  bool
}

// JournalEntry is a short, title-optional note that lives directly on the timeline.
type JournalEntry struct {
	Title        string
	Content      template.HTML
	Date         time.Time
	DateISO      string
	Day          string
	Month        string
	Weekday      string
	WeekdayShort string
	Year         int
	TimeLabel    string
	Mood         string
	Tags         []string
	Image        string
	ImageAlt     string
	ImageCaption string
	IsDraft      bool
}

// TimelineItem lets short notes and longer essays share one chronology.
type TimelineItem struct {
	Kind         string
	Title        string
	Date         time.Time
	DateISO      string
	Day          string
	Month        string
	Weekday      string
	WeekdayShort string
	Year         int
	TimeLabel    string
	Content      template.HTML
	Mood         string
	Tags         []string
	Image        string
	ImageAlt     string
	ImageCaption string
	ArchiveNo    string
}

type TimelineYear struct {
	Year  int
	Items []TimelineItem
}

type LibraryEntry struct {
	Section      string
	Title        string
	ReleaseYear  int
	Rating       float64
	RatingStars  string
	RatingLabel  string
	Date         time.Time
	DateISO      string
	DateLabel    string
	Day          string
	Month        string
	WeekdayShort string
	Year         int
	HasFullDate  bool
	Content      template.HTML
	ArchiveNo    string
	IsDraft      bool
}

type LibraryYear struct {
	Year    int
	Entries []LibraryEntry
}

type PhotoEntry struct {
	Title       string
	Slug        string
	Date        time.Time
	DateISO     string
	DateLabel   string
	Day         string
	Month       string
	MonthLong   string
	MonthNumber int
	Year        int
	Image       string
	ImageAlt    string
	Caption     string
	ArchiveNo   string
	IsDraft     bool
}

type PhotoMonth struct {
	Name   string
	Number int
	Photos []PhotoEntry
}

type PhotoYear struct {
	Year   int
	Months []PhotoMonth
}

// Template data structures
type HomePageData struct {
	PageType      string
	Title         string
	BasePath      string
	Essays        []PostTemplateData
	GroupedEssays []YearGroup
	Timeline      []TimelineYear
	DayCount      int
	FirstYear     int
}

type EssaysPageData struct {
	PageType      string
	Title         string
	BasePath      string
	Essays        []PostTemplateData
	GroupedEssays []YearGroup
}

type YearGroup struct {
	Year  string
	Count int
	Posts []PostTemplateData
}

type PostTemplateData struct {
	Title        string
	Slug         string
	DateLabel    string
	DateISO      string
	Day          string
	Month        string
	Weekday      string
	WeekdayShort string
	Year         int
	Category     string
	Content      template.HTML
	ReadingTime  int
	IsDraft      bool
	CreatedAt    time.Time
	ArchiveNo    string
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

type LibraryIndexPageData struct {
	PageType string
	Title    string
	BasePath string
	Sections []LibraryIndexSection
}

type LibraryIndexSection struct {
	LibrarySection
	HasEntries bool
	CountLabel string
	YearsLabel string
}

type LibraryPageData struct {
	PageType   string
	Title      string
	BasePath   string
	Section    LibrarySection
	Years      []LibraryYear
	ScopeCount string
	YearsLabel string
}

type PhotosPageData struct {
	PageType string
	Title    string
	BasePath string
	Years    []PhotoYear
}

type PhotoPageData struct {
	PageType string
	Title    string
	BasePath string
	Photo    PhotoEntry
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

	// Process all essays
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

	// Convert essays to template data
	postTemplateData := make([]PostTemplateData, 0, len(posts))
	for _, post := range posts {
		if post.IsDraft {
			continue // Skip drafts in static site
		}
		createdAt, _ := time.Parse(time.RFC3339, post.CreatedAt)
		dateLabel := formatDate(post.CreatedAt)
		readingTime := calculateReadingTime(post.Content)
		postTemplateData = append(postTemplateData, PostTemplateData{
			Title:        post.Title,
			Slug:         post.Slug,
			DateLabel:    dateLabel,
			DateISO:      createdAt.Format("2006-01-02"),
			Day:          createdAt.Format("02"),
			Month:        createdAt.Format("Jan"),
			Weekday:      strings.ToLower(createdAt.Format("Monday")),
			WeekdayShort: strings.ToLower(createdAt.Format("Mon")),
			Year:         createdAt.Year(),
			Category:     post.Category,
			Content:      template.HTML(post.Content),
			ReadingTime:  readingTime,
			IsDraft:      post.IsDraft,
			CreatedAt:    createdAt,
		})
	}
	assignPostArchiveNumbers(postTemplateData)

	// Group essays by year
	groupedEssays := groupPostsByYear(postTemplateData)

	journalEntries, err := loadJournalEntries(journalDir)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("▓▓ WARNING: journal entries could not be loaded: %v\n", err)
	}
	timeline, _, dayCount, firstYear := buildTimeline(journalEntries)

	libraryEntries, err := loadLibraryEntries(libraryDir)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("▓▓ WARNING: library entries could not be loaded: %v\n", err)
	}

	photos, err := loadPhotos(photosDir)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("▓▓ WARNING: photos could not be loaded: %v\n", err)
	}

	// Generate pages
	fmt.Println("▓▓ GENERATING PAGES...")
	if err := generateHomePage(templates, postTemplateData, groupedEssays, timeline, dayCount, firstYear); err != nil {
		fmt.Printf("▓▓ ERROR: home page failed: %v\n", err)
	}

	if err := generateEssaysPage(templates, postTemplateData, groupedEssays); err != nil {
		fmt.Printf("▓▓ ERROR: essays page failed: %v\n", err)
	}

	for _, post := range postTemplateData {
		if err := generatePostPage(templates, post); err != nil {
			continue // Skip failed pages silently
		}
	}

	if err := generateAboutPage(templates); err != nil {
		fmt.Printf("▓▓ ERROR: about page failed: %v\n", err)
	}

	if err := generateLibraryIndexPage(templates, libraryEntries); err != nil {
		fmt.Printf("▓▓ ERROR: library index failed: %v\n", err)
	}
	for _, section := range librarySections {
		if err := generateLibraryPage(templates, section, libraryEntries[section.Key]); err != nil {
			fmt.Printf("▓▓ ERROR: library %s page failed: %v\n", section.Key, err)
		}
	}

	if err := generatePhotosPage(templates, photos); err != nil {
		fmt.Printf("▓▓ ERROR: photos page failed: %v\n", err)
	}
	for _, photo := range photos {
		if err := generatePhotoPage(templates, photo); err != nil {
			fmt.Printf("▓▓ ERROR: photo page %s failed: %v\n", photo.Slug, err)
		}
	}

	if templates.Lookup("forai.html") != nil {
		if err := generateForAIPage(templates); err != nil {
			fmt.Printf("▓▓ ERROR: forai page failed: %v\n", err)
		}
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

func generateHomePage(templates *template.Template, posts []PostTemplateData, grouped []YearGroup, timeline []TimelineYear, dayCount, firstYear int) error {
	data := HomePageData{
		PageType:      "home",
		Title:         "Home",
		BasePath:      basePath,
		Essays:        posts,
		GroupedEssays: grouped,
		Timeline:      timeline,
		DayCount:      dayCount,
		FirstYear:     firstYear,
	}

	return writeTemplate(templates, "home.html", filepath.Join(outputDir, "index.html"), data)
}

func generateEssaysPage(templates *template.Template, posts []PostTemplateData, grouped []YearGroup) error {
	data := EssaysPageData{
		PageType:      "essays",
		Title:         "essays",
		BasePath:      basePath,
		Essays:        posts,
		GroupedEssays: grouped,
	}

	return writeTemplate(templates, "essays.html", filepath.Join(outputDir, "essays", "index.html"), data)
}

func generatePostPage(templates *template.Template, post PostTemplateData) error {
	// Create essays/{slug}/index.html structure
	postDir := filepath.Join(outputDir, "essays", post.Slug)
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

func generateLibraryIndexPage(templates *template.Template, entries map[string][]LibraryEntry) error {
	data := LibraryIndexPageData{
		PageType: "library",
		Title:    "library",
		BasePath: basePath,
		Sections: buildLibraryIndexSections(entries),
	}
	return writeTemplate(templates, "library.html", filepath.Join(outputDir, "library", "index.html"), data)
}

func generateLibraryPage(templates *template.Template, section LibrarySection, entries []LibraryEntry) error {
	summary := summarizeLibrarySection(section, entries)
	scopeCount := fmt.Sprintf("%d entries", len(entries))
	if len(entries) == 1 {
		scopeCount = "1 entry"
	}
	data := LibraryPageData{
		PageType:   "library",
		Title:      section.Title,
		BasePath:   basePath,
		Section:    section,
		Years:      groupLibraryEntries(entries),
		ScopeCount: scopeCount,
		YearsLabel: summary.YearsLabel,
	}
	return writeTemplate(
		templates,
		"library-section.html",
		filepath.Join(outputDir, "library", section.Key, "index.html"),
		data,
	)
}

func generatePhotosPage(templates *template.Template, photos []PhotoEntry) error {
	data := PhotosPageData{
		PageType: "photos",
		Title:    "photos",
		BasePath: basePath,
		Years:    groupPhotos(photos),
	}
	return writeTemplate(templates, "photos.html", filepath.Join(outputDir, "photos", "index.html"), data)
}

func generatePhotoPage(templates *template.Template, photo PhotoEntry) error {
	data := PhotoPageData{
		PageType: "photos",
		Title:    photo.Title,
		BasePath: basePath,
		Photo:    photo,
	}
	return writeTemplate(
		templates,
		"photo.html",
		filepath.Join(outputDir, "photos", photo.Slug, "index.html"),
		data,
	)
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

	// Execute template to buffer first
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return fmt.Errorf("template execution failed: %w", err)
	}

	// Format HTML
	formattedHTML, err := formatHTML(buf.Bytes())
	if err != nil {
		// If formatting fails, use original (non-critical)
		formattedHTML = buf.Bytes()
	}

	// Write formatted HTML to file
	if err := os.WriteFile(outputPath, formattedHTML, 0644); err != nil {
		os.Remove(outputPath) // Clean up on error
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}

// formatHTML formats HTML with proper indentation using simple regex-based approach
func formatHTML(input []byte) ([]byte, error) {
	inputStr := string(input)

	// Preserve DOCTYPE
	doctype := ""
	if strings.HasPrefix(strings.TrimSpace(inputStr), "<!DOCTYPE") {
		doctypeIdx := strings.Index(inputStr, "<!DOCTYPE")
		idx := strings.Index(inputStr[doctypeIdx:], ">")
		if idx > 0 {
			doctype = inputStr[doctypeIdx:doctypeIdx+idx+1] + "\n"
			inputStr = strings.TrimSpace(inputStr[doctypeIdx+idx+1:])
		}
	}

	// Simple formatting: add newlines and indentation
	// Replace >< with >\n< (except for inline content)
	inputStr = regexp.MustCompile(`>\s*<`).ReplaceAllString(inputStr, ">\n<")

	// Add indentation
	lines := strings.Split(inputStr, "\n")
	var result []string
	indent := 0
	indentStr := "  "

	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}

	inlineElements := map[string]bool{
		"a": true, "span": true, "strong": true, "em": true, "code": true,
		"b": true, "i": true, "small": true, "sub": true, "sup": true,
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Decrease indent for closing tags
		if strings.HasPrefix(trimmed, "</") {
			indent--
			if indent < 0 {
				indent = 0
			}
		}

		// Add line with indentation
		result = append(result, strings.Repeat(indentStr, indent)+trimmed)

		// Increase indent for opening tags (not void, not self-closing, not inline-only)
		if strings.HasPrefix(trimmed, "<") && !strings.HasPrefix(trimmed, "</") {
			// Extract tag name
			tagName := strings.Fields(strings.TrimPrefix(trimmed, "<"))[0]
			tagName = strings.TrimSuffix(tagName, ">")
			tagName = strings.ToLower(tagName)

			// Check if self-closing
			isSelfClosing := strings.HasSuffix(trimmed, "/>")

			if !voidElements[tagName] && !isSelfClosing && !inlineElements[tagName] {
				indent++
			}
		}
	}

	formatted := strings.Join(result, "\n")
	if doctype != "" {
		formatted = doctype + formatted
	}

	return []byte(formatted), nil
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

func assignPostArchiveNumbers(posts []PostTemplateData) {
	for i := range posts {
		posts[i].ArchiveNo = formatArchiveNumber(len(posts) - i)
	}
}

func loadJournalEntries(dir string) ([]JournalEntry, error) {
	files, err := findMarkdownFiles(dir)
	if err != nil {
		return nil, err
	}

	entries := make([]JournalEntry, 0, len(files))
	for _, path := range files {
		if strings.EqualFold(filepath.Base(path), "README.md") {
			continue
		}
		entry, err := processJournalFile(path)
		if err != nil || entry.IsDraft {
			continue
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})
	return entries, nil
}

func processJournalFile(path string) (JournalEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return JournalEntry{}, err
	}

	meta, body := splitFrontmatter(string(raw))
	if strings.TrimSpace(body) == "" {
		return JournalEntry{}, fmt.Errorf("journal entry has no content: %s", path)
	}

	dateValue := meta["date"]
	if dateValue == "" {
		dateValue = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	entryDate, err := parseJournalDate(dateValue, meta["time"])
	if err != nil {
		return JournalEntry{}, fmt.Errorf("journal entry needs a valid date: %s", path)
	}

	htmlContent, err := renderMarkdown([]byte(strings.TrimSpace(body)))
	if err != nil {
		return JournalEntry{}, err
	}

	var tags []string
	for _, tag := range strings.Split(meta["tags"], ",") {
		tag = strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	timeLabel := ""
	if meta["time"] != "" || strings.Contains(dateValue, "T") {
		timeLabel = entryDate.Format("3:04 pm")
	}

	image := rewriteAssetPath(meta["image"])

	return JournalEntry{
		Title:        meta["title"],
		Content:      template.HTML(htmlContent),
		Date:         entryDate,
		DateISO:      entryDate.Format("2006-01-02"),
		Day:          entryDate.Format("02"),
		Month:        entryDate.Format("Jan"),
		Weekday:      strings.ToLower(entryDate.Format("Monday")),
		WeekdayShort: strings.ToLower(entryDate.Format("Mon")),
		Year:         entryDate.Year(),
		TimeLabel:    timeLabel,
		Mood:         meta["mood"],
		Tags:         tags,
		Image:        image,
		ImageAlt:     meta["image_alt"],
		ImageCaption: meta["image_caption"],
		IsDraft:      strings.EqualFold(meta["draft"], "true"),
	}, nil
}

func splitFrontmatter(raw string) (map[string]string, string) {
	meta := make(map[string]string)
	if !strings.HasPrefix(raw, "---") {
		return meta, raw
	}
	parts := strings.SplitN(raw, "---", 3)
	if len(parts) < 3 {
		return meta, raw
	}
	for _, line := range strings.Split(parts[1], "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		meta[strings.ToLower(strings.TrimSpace(key))] = strings.Trim(strings.TrimSpace(value), "\"'")
	}
	return meta, strings.TrimSpace(parts[2])
}

func parseJournalDate(dateValue, timeValue string) (time.Time, error) {
	dateValue = strings.TrimSpace(dateValue)
	if timeValue != "" && !strings.Contains(dateValue, "T") {
		dateValue += "T" + strings.TrimSpace(timeValue)
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"} {
		if parsed, err := time.Parse(layout, dateValue); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q", dateValue)
}

func rewriteAssetPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || basePath == "/" || !strings.HasPrefix(path, "/") {
		return path
	}
	return basePath + strings.TrimPrefix(path, "/")
}

func renderMarkdown(content []byte) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithHardWraps(),
			goldmarkhtml.WithXHTML(),
		),
	)

	var rendered strings.Builder
	if err := md.Convert(content, &rendered); err != nil {
		return "", fmt.Errorf("could not render markdown: %w", err)
	}

	htmlStr := strings.ReplaceAll(rendered.String(), "<img ", "<img loading=\"lazy\" decoding=\"async\" ")
	if basePath != "/" {
		imgSrcRegex := regexp.MustCompile(`src=["'](/images/[^"']+)["']`)
		htmlStr = imgSrcRegex.ReplaceAllStringFunc(htmlStr, func(match string) string {
			submatches := imgSrcRegex.FindStringSubmatch(match)
			if len(submatches) < 2 {
				return match
			}
			quote := `"`
			if strings.Contains(match, `'`) {
				quote = `'`
			}
			return `src=` + quote + basePath + strings.TrimPrefix(submatches[1], "/") + quote
		})
	}
	return htmlStr, nil
}

func buildTimeline(entries []JournalEntry) ([]TimelineYear, int, int, int) {
	items := make([]TimelineItem, 0, len(entries))
	days := make(map[string]bool)

	for _, entry := range entries {
		items = append(items, TimelineItem{
			Kind:         "note",
			Title:        entry.Title,
			Date:         entry.Date,
			DateISO:      entry.DateISO,
			Day:          entry.Day,
			Month:        entry.Month,
			Weekday:      entry.Weekday,
			WeekdayShort: entry.WeekdayShort,
			Year:         entry.Year,
			TimeLabel:    entry.TimeLabel,
			Content:      entry.Content,
			Mood:         entry.Mood,
			Tags:         entry.Tags,
			Image:        entry.Image,
			ImageAlt:     entry.ImageAlt,
			ImageCaption: entry.ImageCaption,
		})
		days[entry.DateISO] = true
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Date.After(items[j].Date)
	})

	for i := range items {
		items[i].ArchiveNo = formatArchiveNumber(len(items) - i)
	}

	var years []TimelineYear
	for _, item := range items {
		if len(years) == 0 || years[len(years)-1].Year != item.Year {
			years = append(years, TimelineYear{Year: item.Year})
		}
		years[len(years)-1].Items = append(years[len(years)-1].Items, item)
	}

	firstYear := time.Now().Year()
	if len(items) > 0 {
		firstYear = items[len(items)-1].Year
	}
	return years, len(items), len(days), firstYear
}

func loadLibraryEntries(dir string) (map[string][]LibraryEntry, error) {
	result := make(map[string][]LibraryEntry, len(librarySections))
	if _, err := os.Stat(dir); err != nil {
		return result, err
	}

	for _, section := range librarySections {
		sectionDir := filepath.Join(dir, section.Key)
		files, err := findMarkdownFiles(sectionDir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return result, err
		}

		for _, path := range files {
			if strings.EqualFold(filepath.Base(path), "README.md") {
				continue
			}
			entry, err := processLibraryFile(path, section.Key)
			if err != nil || entry.IsDraft {
				continue
			}
			result[section.Key] = append(result[section.Key], entry)
		}

		sort.SliceStable(result[section.Key], func(i, j int) bool {
			return result[section.Key][i].Date.After(result[section.Key][j].Date)
		})
		for i := range result[section.Key] {
			result[section.Key][i].ArchiveNo = formatArchiveNumber(len(result[section.Key]) - i)
		}
	}
	return result, nil
}

func processLibraryFile(path, section string) (LibraryEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return LibraryEntry{}, err
	}
	meta, body := splitFrontmatter(string(raw))
	if strings.TrimSpace(meta["title"]) == "" {
		return LibraryEntry{}, fmt.Errorf("library entry needs a title: %s", path)
	}
	if strings.TrimSpace(body) == "" {
		return LibraryEntry{}, fmt.Errorf("library entry needs a note: %s", path)
	}

	entryDate, dateISO, dateLabel, hasFullDate, err := parseLibraryDate(meta["date"])
	if err != nil {
		return LibraryEntry{}, fmt.Errorf("library entry needs a valid date: %s", path)
	}
	rating, err := strconv.ParseFloat(meta["rating"], 64)
	if err != nil || rating < 0.5 || rating > 5 || math.Mod(rating*2, 1) != 0 {
		return LibraryEntry{}, fmt.Errorf("library entry rating must be between 0.5 and 5 in half-star steps: %s", path)
	}

	releaseYear := 0
	if meta["release_year"] != "" {
		releaseYear, err = strconv.Atoi(meta["release_year"])
		if err != nil || releaseYear < 1 {
			return LibraryEntry{}, fmt.Errorf("library entry has an invalid release year: %s", path)
		}
	}

	fullStars := int(rating)
	ratingStars := strings.Repeat("★", fullStars)
	remainingStars := 5 - fullStars
	if rating-float64(fullStars) == 0.5 {
		ratingStars += "½"
		remainingStars--
	}
	ratingStars += strings.Repeat("☆", remainingStars)

	htmlContent, err := renderMarkdown([]byte(strings.TrimSpace(body)))
	if err != nil {
		return LibraryEntry{}, err
	}
	entry := LibraryEntry{
		Section:     section,
		Title:       meta["title"],
		ReleaseYear: releaseYear,
		Rating:      rating,
		RatingStars: ratingStars,
		RatingLabel: fmt.Sprintf("%g out of 5", rating),
		Date:        entryDate,
		DateISO:     dateISO,
		DateLabel:   dateLabel,
		Year:        entryDate.Year(),
		HasFullDate: hasFullDate,
		Content:     template.HTML(htmlContent),
		IsDraft:     strings.EqualFold(meta["draft"], "true"),
	}
	if hasFullDate {
		entry.Day = entryDate.Format("02")
		entry.Month = entryDate.Format("Jan")
		entry.WeekdayShort = strings.ToLower(entryDate.Format("Mon"))
	}
	return entry, nil
}

func parseLibraryDate(value string) (time.Time, string, string, bool, error) {
	value = strings.TrimSpace(value)
	if len(value) == 4 {
		year, err := strconv.Atoi(value)
		if err == nil && year > 0 {
			return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), value, value, false, nil
		}
	}
	entryDate, err := parseJournalDate(value, "")
	if err != nil {
		return time.Time{}, "", "", false, err
	}
	return entryDate, entryDate.Format("2006-01-02"), entryDate.Format("2 Jan 2006"), true, nil
}

func groupLibraryEntries(entries []LibraryEntry) []LibraryYear {
	var years []LibraryYear
	for _, entry := range entries {
		if len(years) == 0 || years[len(years)-1].Year != entry.Year {
			years = append(years, LibraryYear{Year: entry.Year})
		}
		years[len(years)-1].Entries = append(years[len(years)-1].Entries, entry)
	}
	return years
}

func buildLibraryIndexSections(entries map[string][]LibraryEntry) []LibraryIndexSection {
	sections := make([]LibraryIndexSection, 0, len(librarySections))
	for _, section := range librarySections {
		sections = append(sections, summarizeLibrarySection(section, entries[section.Key]))
	}
	return sections
}

func summarizeLibrarySection(section LibrarySection, entries []LibraryEntry) LibraryIndexSection {
	summary := LibraryIndexSection{LibrarySection: section}
	if len(entries) == 0 {
		return summary
	}

	summary.HasEntries = true
	countText := section.CountMany
	if len(entries) == 1 {
		countText = section.CountOne
	}
	summary.CountLabel = fmt.Sprintf("%d %s", len(entries), countText)

	earliestYear := entries[0].Year
	latestYear := entries[0].Year
	for _, entry := range entries[1:] {
		if entry.Year < earliestYear {
			earliestYear = entry.Year
		}
		if entry.Year > latestYear {
			latestYear = entry.Year
		}
	}
	if earliestYear == latestYear {
		summary.YearsLabel = fmt.Sprintf("%d", latestYear)
	} else {
		summary.YearsLabel = fmt.Sprintf("%d — %d", earliestYear, latestYear)
	}
	return summary
}

func loadPhotos(dir string) ([]PhotoEntry, error) {
	files, err := findMarkdownFiles(dir)
	if err != nil {
		return nil, err
	}
	photos := make([]PhotoEntry, 0, len(files))
	seenSlugs := make(map[string]bool)
	for _, path := range files {
		if strings.EqualFold(filepath.Base(path), "README.md") {
			continue
		}
		photo, err := processPhotoFile(path)
		if err != nil || photo.IsDraft || seenSlugs[photo.Slug] {
			continue
		}
		seenSlugs[photo.Slug] = true
		photos = append(photos, photo)
	}
	sort.SliceStable(photos, func(i, j int) bool {
		return photos[i].Date.After(photos[j].Date)
	})
	for i := range photos {
		photos[i].ArchiveNo = formatArchiveNumber(len(photos) - i)
	}
	return photos, nil
}

func processPhotoFile(path string) (PhotoEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return PhotoEntry{}, err
	}
	meta, _ := splitFrontmatter(string(raw))
	title := strings.TrimSpace(meta["title"])
	if title == "" {
		return PhotoEntry{}, fmt.Errorf("photo needs a title: %s", path)
	}
	photoDate, err := parseJournalDate(meta["date"], "")
	if err != nil {
		return PhotoEntry{}, fmt.Errorf("photo needs a valid date: %s", path)
	}
	image := rewriteAssetPath(meta["image"])
	if image == "" {
		return PhotoEntry{}, fmt.Errorf("photo needs an image: %s", path)
	}
	if strings.TrimSpace(meta["image_alt"]) == "" {
		return PhotoEntry{}, fmt.Errorf("photo needs image alt text: %s", path)
	}
	slug := meta["slug"]
	if slug == "" {
		slug = generateSlug(title)
	} else if generateSlug(slug) != slug {
		return PhotoEntry{}, fmt.Errorf("photo slug must contain only lowercase letters, numbers, and hyphens: %s", path)
	}
	if slug == "" {
		return PhotoEntry{}, fmt.Errorf("photo needs a valid slug: %s", path)
	}

	return PhotoEntry{
		Title:       title,
		Slug:        slug,
		Date:        photoDate,
		DateISO:     photoDate.Format("2006-01-02"),
		DateLabel:   photoDate.Format("2 Jan 2006"),
		Day:         photoDate.Format("02"),
		Month:       photoDate.Format("Jan"),
		MonthLong:   photoDate.Format("January"),
		MonthNumber: int(photoDate.Month()),
		Year:        photoDate.Year(),
		Image:       image,
		ImageAlt:    meta["image_alt"],
		Caption:     meta["caption"],
		IsDraft:     strings.EqualFold(meta["draft"], "true"),
	}, nil
}

func groupPhotos(photos []PhotoEntry) []PhotoYear {
	var years []PhotoYear
	for _, photo := range photos {
		if len(years) == 0 || years[len(years)-1].Year != photo.Year {
			years = append(years, PhotoYear{Year: photo.Year})
		}
		year := &years[len(years)-1]
		if len(year.Months) == 0 || year.Months[len(year.Months)-1].Number != photo.MonthNumber {
			year.Months = append(year.Months, PhotoMonth{
				Name:   photo.MonthLong,
				Number: photo.MonthNumber,
			})
		}
		month := &year.Months[len(year.Months)-1]
		month.Photos = append(month.Photos, photo)
	}
	return years
}

func formatArchiveNumber(number int) string {
	return fmt.Sprintf("%03d", number)
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

// parseFrontmatter parses simple YAML frontmatter (only handles key: value pairs)
func parseFrontmatter(yamlContent string, fm *Frontmatter) error {
	lines := strings.Split(strings.TrimSpace(yamlContent), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Simple key: value parser
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		switch strings.ToLower(key) {
		case "title":
			fm.Title = value
		case "category":
			fm.Category = value
		case "date":
			fm.Date = value
		case "slug":
			fm.Slug = value
		case "draft":
			fm.IsDraft = strings.ToLower(value) == "true"
		}
	}

	return nil
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
			// Parse frontmatter manually (simple YAML parser for our needs)
			if err := parseFrontmatter(parts[1], &frontmatter); err != nil {
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
			goldmarkhtml.WithHardWraps(),
			goldmarkhtml.WithXHTML(),
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
		return fmt.Errorf("no image directory found; proceeding without")
	}

	if err := os.MkdirAll(publicImagesDir, 0755); err != nil {
		return fmt.Errorf("difficulty in preparing the image directory: %w", err)
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
		return fmt.Errorf("difficulty in gathering images: %w", err)
	}

	if copied > 0 {
		fmt.Printf("▓▓ COPIED %d IMAGE%s\n", copied, strings.ToUpper(plural(copied)))
	}
	return nil
}

func generateRSSFeed(posts []PostTemplateData) error {
	if len(posts) == 0 {
		return nil // No essays, skip RSS generation
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
<title>for later, when i forget</title>
<link>%s</link>
<description>essays and observations by karthik</description>
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
		postURL := fmt.Sprintf("%s%sessays/%s", siteURL, basePath, post.Slug)

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
