// TODO - Other tasks
//
//Here are some simple tasks you might want to tackle on your own:
//
// [X] Store templates in tmpl/ and page data in data/.
// [X] Add a handler to make the web root redirect to /view/FrontPage.
// [X] Spruce up the page templates by making them valid HTML and adding some CSS rules.
// [ ] Implement inter-page linking by converting instances of [PageName] to
// <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this)

package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

// NOTE - Globals
var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))
var validPath = regexp.MustCompile(`^/(edit|save|view)/([a-zA-Z0-9]+)$`)

// NOTE - Types
type Page struct {
	Title string
	Body  []byte
}

// NOTE - main
func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// NOTE - Handlers
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

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/home", http.StatusMovedPermanently)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", page)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	} else {
		page.Body = []byte(string(page.Body))
	}
	renderTemplate(w, "edit", page)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	err := page.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// NOTE -functions
func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	if p.Body == nil {
		p.Body = []byte("")
	}
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	pageAsString := struct {
		Title string
		Body  string
	}{
		Title: p.Title,
		Body:  string(p.Body),
	}

	err := templates.ExecuteTemplate(w, tmpl+".html", pageAsString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
