// https://golang.org/doc/articles/wiki/

// tasks
// - Store templates in tmpl/ and page data in data/.
//    - improve code efficiency
//
// - Implement inter-page linking by converting instances of [PageName] to
// <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this)
//
// - Spruce up the page templates by making them valid HTML and adding some CSS rules.
//
// - Add a home button

package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

var templates = template.Must(template.ParseFiles(
	"tmpl/edit.html",
	"tmpl/view.html",
))

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

var rootTitle = "FrontPage"

// describe a webpage
type Page struct {
	// Body is []byte instead of string because io libraries
	// expect type []byte
	Title string
	Body  []byte
}

// check title is valid URL
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

// save webpage to a text file
func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// load webpage from a text file
func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

// load template
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, "tmpl/"+tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// root page Handler
func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/"+rootTitle, http.StatusFound)
}

// view page handler
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

// edit page handler
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

// save page handler
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

// returns a proper function for http.HandleFunc and runs the handler function
// while passing in the title
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

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}
