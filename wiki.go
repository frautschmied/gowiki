// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"io/ioutil"
	"io"
	"log"
	"net/http"
	"regexp"
	"bytes"
	"os"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func homeHandler(w http.ResponseWriter, r *http.Request, title string) {
	viewHandler(w, r, "FrontPage")
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

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

func makePageLink(x []byte) []byte {
	y := string(x[1: len(x)-1])
	ret := "<a href=\"/view/" + y + "\">" + y + "</a>"
	return []byte(ret)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	m := new(bytes.Buffer)
	err := templates.ExecuteTemplate(m, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	re := regexp.MustCompile("\\[[a-zA-Z0-9]+\\]")
	linked := re.ReplaceAllFunc(m.Bytes(), makePageLink)
	_, err = io.WriteString(w, string(linked))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			fn(w, r, "")
			return
		}
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}



func main() {
	os.Mkdir("data", 0755)

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/"     , makeHandler(homeHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
