package main

import (
	"bufio"
	"github.com/russross/blackfriday/v2"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Post struct {
	Metadata PostMetadata
	Body     template.HTML
}

type Index struct {
	PostsMetadata []PostMetadata
}

type PostMetadata struct {
	Title string
	Date  string
	Link  string
}

const mdDir = "../md/"
const numMetadataLines = 4

func split(input []byte) (map[string]string, []byte) {
	frontMatter := make(map[string]string)

	array := strings.Split(string(input), "---")
	r1 := strings.NewReader(array[0])
	r2 := strings.NewReader(array[1])

	scanner := bufio.NewScanner(r1)
	for scanner.Scan() {
		metadataLine := strings.Split(scanner.Text(), ":")
		frontMatter[metadataLine[0]] = metadataLine[1]
	}

	content, _ := ioutil.ReadAll(r2)
	return frontMatter, content
}

func parse(fileName string) (map[string]string, []byte) {
	b, _ := ioutil.ReadFile(fileName)

	frontMatter, content := split(b)
	return frontMatter, content

}

func FormatBlogPostName(blogPostName string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(blogPostName)), " ", "-")

}

func CreateLink(blogPostName string) string {
	blogPostName = FormatBlogPostName(blogPostName)
	return blogPostName
}

func GenerateBlog() {
	files, _ := ioutil.ReadDir(mdDir)
	index := Index{PostsMetadata: []PostMetadata{}}

	for _, file := range files {
		frontMatter, content := parse(mdDir + file.Name())

		body := blackfriday.Run(content)
		blogPostName := FormatBlogPostName(frontMatter["title"])
		link := CreateLink(blogPostName)
		log.Println("Creating link for blog.html at ", link)
		metadata := PostMetadata{Title: frontMatter["title"],
			Date: frontMatter["date"], Link: link}
		post := &Post{Metadata: metadata, Body: template.HTML(body)}

		t := template.Must(template.ParseFiles("templates/post.html"))

		f, err := os.Create("../site/posts/" + blogPostName + ".html")

		defer f.Close()
		if err != nil {
			log.Println("create file: ", err)
			return
		}
		_ = t.Execute(f, post)

		index.PostsMetadata = append(index.PostsMetadata, metadata)

	}
	indexFileName := "../site/index.html"

	t := template.Must(template.ParseFiles("templates/index.html"))
	f, _ := os.Create(indexFileName)
	defer f.Close()
	_ = t.Execute(f, index)
}

func main() {
	GenerateBlog()
}
