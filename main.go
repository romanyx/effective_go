package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/russross/blackfriday"
)

var (
	addr    = flag.String("hostport", "localhost:8080", "server address")
	logPath = flag.String("log", "", "Log file path, default is output")
)

const (
	mdDir       = "md"
	readTimeout = 9 * time.Second
)

func main() {
	flag.Parse()

	if err := setLogOut(*logPath); err != nil {
		log.Fatal(err)
	}

	html, err := generateHTML(mdDir)
	if err != nil {
		log.Fatal(err)
	}

	s := NewServer(*addr, ReadTimeout(readTimeout), GET("/", indexHandler(html)), GzipOn)

	log.Fatal(s.ListenAndServe())
}

func indexHandler(html []byte) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := r.Context()

		select {
		case <-ctx.Done():
			err := ctx.Err()
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			w.Write(html)
		}
	}
}

func sortFiles(files []os.FileInfo) []string {
	s := make([]string, len(files), len(files))

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		index, err := strconv.Atoi(strings.Split(f.Name(), "_")[0])
		if err != nil {
			log.Println(err)
		}

		s[index-1] = fmt.Sprintf("%s/%s", mdDir, f.Name())
	}

	return s
}

func generateHTML(dir string) ([]byte, error) {
	buf := new(bytes.Buffer)
	files, _ := ioutil.ReadDir(dir)

	paths := sortFiles(files)

	for _, path := range paths {
		file, err := os.Open(path)

		if err != nil {
			return []byte{}, err
		}

		io.Copy(buf, file)
		if err := file.Close(); err != nil {
			return []byte{}, errors.Wrapf(err, "on file close %s", path)
		}
	}

	content := template.HTML(blackfriday.MarkdownCommon(buf.Bytes()))

	t, err := template.ParseFiles("templates/layout.html")
	if err != nil {
		return []byte{}, err
	}

	view := struct{ Content template.HTML }{content}
	buf = new(bytes.Buffer)

	if err := t.ExecuteTemplate(buf, "layout.html", view); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func setLogOut(path string) error {
	if path == "" {
		return nil
	}

	logOut, err := os.Open(path)
	if err != nil {
		return err
	}

	log.SetOutput(logOut)

	return nil
}
