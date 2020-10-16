package web

import (
"net/http"
"html/template"
"log"
)

var debug = true

type Page struct{
  Name string
  Content string
  Path string
  Icon string
  Site *Site
  Tpl *template.Template
  Func func(w http.ResponseWriter, r *http.Request)()
  }

func (p *Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  if debug {
    p.Tpl = template.New(p.Name)
    p.Tpl = template.Must(p.Tpl.ParseFiles("gotpl/index.gohtml","gotpl/menu.gohtml",p.Content))
  }
  err := p.Tpl.ExecuteTemplate(w, "layout", p.Site)
  if err != nil {
      log.Fatalf("Template execution: %s", err)
  }
  p.Func(w,r)
}
