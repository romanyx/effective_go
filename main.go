package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/russross/blackfriday"
)

const (
	mdDir = "md"
)

func main() {
	buf := new(bytes.Buffer)
	files, _ := ioutil.ReadDir(mdDir)

	sortedFilePaths := sortFiles(files)

	for _, path := range sortedFilePaths {
		file, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		io.Copy(buf, file)
	}

	content := template.HTML(blackfriday.MarkdownCommon(buf.Bytes()))

	t, err := template.ParseFiles("templates/layout.html")
	if err != nil {
		log.Fatal(err)
	}

	view := struct{ Content template.HTML }{content}
	buf = new(bytes.Buffer)

	if err := t.ExecuteTemplate(buf, "layout.html", view); err != nil {
		log.Fatal(err)
	}

	html := buf.Bytes()
	buf.Reset()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(html)
	})

	http.ListenAndServe(":8080", nil)
}

func sortFiles(files []os.FileInfo) []string {
	s := make([]string, len(files), len(files))

	for _, f := range files {
		if !f.IsDir() {
			index, err := strconv.Atoi(strings.Split(f.Name(), "_")[0])

			if err != nil {
				log.Println(err)
			}

			s[index-1] = fmt.Sprintf("%s/%s", mdDir, f.Name())
		}
	}

	return s
}
