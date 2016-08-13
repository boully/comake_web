// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

var root, _ = os.Getwd()
var fold = path.Join(root, "COMAKE_PATH")

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := path.Join(fold, p.Title+".txt")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := path.Join(fold, title+".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func comakeHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/text")
	w.Write(p.Body)
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var indexTempl = template.Must(template.ParseFiles("index.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view|comake)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(fold)
	if err != nil {
		log.Fatal(err)
	}
	fileList := []string{}
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, strings.TrimSuffix(file.Name(), ".txt"))
		}
	}
	err = indexTempl.ExecuteTemplate(w, "index.html", fileList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("comake")
	fmt.Println(title)
	if len(title) != 0 {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	http.Error(w, "missing comake file name", http.StatusInternalServerError)
}

func init() {
	if _, err := os.Stat(fold); os.IsNotExist(err) {
		os.Mkdir(fold, 0700)
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/comake/", makeHandler(comakeHandler))
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/index", indexHandler)

	http.ListenAndServe(":8080", nil)
}
