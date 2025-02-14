package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/fsnotify/fsnotify"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/pelletier/go-toml/v2"
)

// Blog post read from a markdown file
type Post struct {
	Title   string    `toml:"title"`
	Date    time.Time `toml:"date"`
	Slug    string    `toml:"slug"`
	Content string    `toml:"content"`
}

// Main app function
func main() {
	// Check if dev flag is present
	isDev := false
	if len(os.Args) > 1 && os.Args[1] == "--dev" {
		isDev = true
	}

	// Build the site
	if err := build(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Site built successfully in docs/")

	// If in dev mode, start the server and file watcher
	if isDev {
		log.Printf("Starting development server")
		go watchChanges()

		log.Printf("Server running at http://localhost:8080")
		http.Handle("/", http.FileServer(http.Dir("docs")))
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

// Build the blog
// create subdirectories, parse all markdown files, generate html files
func build() error {
	// Clear output directory if it exists
	if err := os.RemoveAll("docs"); err != nil {
		return fmt.Errorf("failed to clear output directory: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll("docs", 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Copy static assets from theme/static
	if err := copyDir("theme/static", "docs"); err != nil {
		log.Printf("Warning: failed to copy static assets: %v", err)
	}

	// Copy all assets to /docs/assets
	assetsDestDir := filepath.Join("docs", "assets")
	if err := os.MkdirAll(assetsDestDir, 0755); err != nil {
		return fmt.Errorf("failed to create assets directory: %w", err)
	}
	if err := copyDir("assets", assetsDestDir); err != nil {
		log.Printf("Warning: failed to copy assets: %v", err)
	}

	// Parse and generate posts
	posts, err := parsePosts()
	if err != nil {
		return fmt.Errorf("failed to parse posts: %w", err)
	}

	// Generate pages
	if err := generateSite(posts); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	// Generate syntax highlighting CSS
	style := styles.Get("base16-snazzy")
	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.ClassPrefix("ch-"),
	)
	cssFile := filepath.Join("docs", "syntax.css")
	f, err := os.Create(cssFile)
	if err != nil {
		return fmt.Errorf("failed to create syntax.css: %w", err)
	}
	defer f.Close()
	if err := formatter.WriteCSS(f, style); err != nil {
		return fmt.Errorf("failed to write syntax.css: %w", err)
	}

	return nil
}

// Parse all markdown files into []Post
func parsePosts() ([]Post, error) {
	files, err := os.ReadDir("posts")
	if err != nil {
		return nil, fmt.Errorf("failed to read posts directory: %w", err)
	}

	var posts []Post
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			post, err := parsePost(filepath.Join("posts", file.Name()))
			if err != nil {
				log.Printf("Warning: failed to parse post %s: %v", file.Name(), err)
				continue
			}
			posts = append(posts, post)
		}
	}

	return posts, nil
}

// Parse a single markdown file into a Post
func parsePost(filename string) (Post, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return Post{}, err
	}

	// Split frontmatter and content
	parts := strings.SplitN(string(content), "+++", 3)
	if len(parts) != 3 {
		return Post{}, fmt.Errorf("invalid frontmatter format")
	}

	// Parse frontmatter
	var post Post
	if err := toml.Unmarshal([]byte(parts[1]), &post); err != nil {
		return Post{}, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Configure markdown parser with fenced code extension
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.FencedCode
	p := parser.NewWithExtensions(extensions)

	// Parse markdown
	md := markdown.NormalizeNewlines([]byte(parts[2]))
	doc := p.Parse(md)

	// Create HTML renderer with syntax highlighting
	htmlFlags := mdhtml.CommonFlags | mdhtml.HrefTargetBlank
	opts := mdhtml.RendererOptions{Flags: htmlFlags}
	baseRenderer := mdhtml.NewRenderer(opts)
	renderer := &customRenderer{baseRenderer}

	html := markdown.Render(doc, renderer)
	post.Content = string(html)

	return post, nil
}

// Watch for changes in the source directory and rebuild the site
func watchChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Watch content and static directories
	dirsToWatch := []string{
		"posts",
		"assets",
		"theme/static",
		"theme/templates",
	}

	for _, dir := range dirsToWatch {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("Error accessing path %s: %v", path, err)
				return nil
			}
			if info.IsDir() {
				return watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			log.Printf("Error watching directory %s: %v", dir, err)
		}
	}

	// Watch for changes
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				log.Printf("Detected change in %s", event.Name)
				// Rebuild static files
				if err := build(); err != nil {
					log.Printf("Error rebuilding site: %v", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// Copy a directory and all its contents
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}

// generateSite creates HTML pages for all posts and the index page
func generateSite(posts []Post) error {
	// Load templates
	templatesDir := "theme/templates"

	// Load post template
	postTmpl, err := template.New("post").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"now": time.Now,
	}).ParseFiles(
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "post.html"),
	)
	if err != nil {
		return fmt.Errorf("failed to parse post template: %w", err)
	}

	// Load index template
	indexTmpl, err := template.New("index").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"now": time.Now,
	}).ParseFiles(
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "index.html"),
	)
	if err != nil {
		return fmt.Errorf("failed to parse index template: %w", err)
	}

	// Create posts directory
	postsDir := filepath.Join("docs", "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		return fmt.Errorf("failed to create posts directory: %w", err)
	}

	// Generate individual post pages
	for _, post := range posts {
		outPath := filepath.Join(postsDir, post.Slug+".html")
		file, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create post file %s: %w", post.Slug, err)
		}
		defer file.Close()

		data := struct {
			Post    Post
			BaseURL string
			SiteURL string
		}{
			Post:    post,
			BaseURL: "/",
			SiteURL: "",
		}

		if err := postTmpl.ExecuteTemplate(file, "post.html", data); err != nil {
			return fmt.Errorf("failed to execute post template for %s: %w", post.Slug, err)
		}
	}

	// Generate index page
	indexPath := filepath.Join("docs", "index.html")
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}
	defer indexFile.Close()

	// Sort posts by date (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	indexData := struct {
		Posts   []Post
		BaseURL string
		SiteURL string
	}{
		Posts:   posts,
		BaseURL: "/",
		SiteURL: "",
	}

	if err := indexTmpl.ExecuteTemplate(indexFile, "index.html", indexData); err != nil {
		return fmt.Errorf("failed to execute index template: %w", err)
	}

	return nil
}

// Apply syntax highlighting to code blocks
func highlightCode(code, language string) string {
	l := lexers.Get(language)
	if l == nil {
		l = lexers.Fallback
	}

	iterator, err := l.Tokenise(nil, code)
	if err != nil {
		return code
	}

	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.ClassPrefix("ch-"),
	)
	style := styles.Get("base16-snazzy")

	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code
	}

	return buf.String()
}

// Create custom html node renderer
type customRenderer struct {
	*mdhtml.Renderer
}

func (r *customRenderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	if code, ok := node.(*ast.CodeBlock); ok && entering {
		language := string(code.Info)
		highlighted := highlightCode(string(code.Literal), language)
		fmt.Fprintf(w, "<pre><code class=\"language-%s\">%s</code></pre>\n", language, highlighted)
		return ast.SkipChildren
	}
	return r.Renderer.RenderNode(w, node, entering)
}
