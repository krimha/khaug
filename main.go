package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

type Post struct {
	Title    string    `yaml:"title"`
	Date     time.Time `yaml:"date"`
	Tags     []string  `yaml:"tags"`
	Slug     string
	Content  template.HTML
	RootPath string
}

type IndexData struct {
	Title    string
	Posts    []Post
	RootPath string
}

func main() {
	serve := flag.Bool("serve", false, "serve the generated site")
	flag.Parse()

	if err := generateSite(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Site generated successfully in 'dist/'")

	if *serve {
		addr := ":8080"
		fmt.Printf("Serving site at http://localhost%s\n", addr)
		log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir("dist"))))
	}
}

func generateSite() error {
	// 1. Clean/Create dist
	os.RemoveAll("dist")
	if err := os.MkdirAll("dist/posts", 0755); err != nil {
		return err
	}

	// 2. Parse Posts
	posts, err := parsePosts("content")
	if err != nil {
		return err
	}

	// Sort posts by date (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	// 3. Render Individual Posts
	postTmpl, err := template.ParseFiles("templates/layout.html", "templates/post.html")
	if err != nil {
		return err
	}

	for i := range posts {
		posts[i].RootPath = "../"
		f, err := os.Create(filepath.Join("dist/posts", posts[i].Slug+".html"))
		if err != nil {
			return err
		}
		defer f.Close()

		if err := postTmpl.ExecuteTemplate(f, "layout.html", posts[i]); err != nil {
			return err
		}
	}

	// 4. Render Index Page
	indexTmpl, err := template.ParseFiles("templates/layout.html", "templates/index.html")
	if err != nil {
		return err
	}

	indexFile, err := os.Create("dist/index.html")
	if err != nil {
		return err
	}
	defer indexFile.Close()

	if err := indexTmpl.ExecuteTemplate(indexFile, "layout.html", IndexData{
		Title:    "Home",
		Posts:    posts,
		RootPath: "./",
	}); err != nil {
		return err
	}

	// 5. Copy Static Assets
	return copyStatic("static", "dist")
}

func parsePosts(dir string) ([]Post, error) {
	var posts []Post
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}

		post, err := parsePost(path)
		if err != nil {
			return fmt.Errorf("error parsing %s: %w", path, err)
		}
		posts = append(posts, post)
		return nil
	})
	return posts, err
}

func parsePost(path string) (Post, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Post{}, err
	}

	// Split Front Matter
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		return Post{}, fmt.Errorf("invalid front matter in %s", path)
	}

	var post Post
	if err := yaml.Unmarshal([]byte(parts[1]), &post); err != nil {
		return Post{}, err
	}

	// Markdown to HTML
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(parts[2]), &buf); err != nil {
		return Post{}, err
	}

	post.Content = template.HTML(buf.String())
	post.Slug = strings.TrimSuffix(filepath.Base(path), ".md")

	return post, nil
}

func copyStatic(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(target)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
