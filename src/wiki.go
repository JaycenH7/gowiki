// https://golang.org/doc/articles/wiki/
//
// tasks
//
// - Implement inter-page linking by converting instances of [PageName] to
// <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this)
//
// - Spruce up the page templates by making them valid HTML and adding some CSS rules.
//
// - Add a home button
//

package main

import (
    "strings"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
    "log"
    "os"
    "flag"
    "fmt"
)

// unused imports
var _ = fmt.Printf

// define variables
var rootTitle = "FrontPage"
var templates = template.Must(template.ParseGlob("./templates/*.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// define structs
var logger *log.Logger
var logFile *os.File

// command-line options
var logLevel string
var listenPort int

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
    //err := templates.ExecuteTemplate(w, "./tmpl/"+tmpl+".html", p)
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
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

func setLogLevel() {
    logLevel = strings.ToUpper(logLevel)
    formatLogLevel := logLevel + ":"

    logFile, err := os.OpenFile("./log/wiki.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        fmt.Printf("error opening file: %v", err)
        os.Exit(1)
    }
    logger = log.New(logFile, formatLogLevel, log.Ltime)
    // logger = log.New(os.Stdout, formatLogLevel, log.Ltime)

    initLog()
}

func initLog() {
    template_list := templates.DefinedTemplates()
    if logLevel == "DEBUG" {
        logger.Print(template_list)
    }
}

func servePages() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}

func parseArgs() {
    flag.StringVar(&logLevel, "log", "INFO", "logging level")
    flag.IntVar(&listenPort, "port", 8080, "listening port")
    flag.Parse()
}

func main() {
    parseArgs()
    setLogLevel()
    servePages()
}
