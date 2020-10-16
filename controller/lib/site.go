package web

import (
"net/http"
"html/template"
"log"
"github.com/gorilla/mux"
"github.com/gorilla/websocket"
)

type Site struct {
  Pages []*Page
  Mux *mux.Router
  LogChan chan string
  logChanWS []chan string
  logUpgraderWS websocket.Upgrader
  Data map[string]interface{}
}

func (s *Site) Init() {
  s.Data = make(map[string]interface{})
  s.Data["Test"] = "test"
  s.LogChan = make(chan string)
  s.logChanWS = []chan string {}
  s.logUpgraderWS = websocket.Upgrader{}
  go func() {
    for {
      msg := <-s.LogChan
      for _,v := range s.logChanWS {
        v <- msg
      }
    }
  }()
  s.Mux = mux.NewRouter()
  s.Mux.Handle("/", s)
  s.Mux.HandleFunc("/exec",s.wsHandlerLogPage)
  http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
  //s.AddPage("Log","gotpl/log.gohtml","/log","bug", s.logPage)
}

func (s *Site) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  tpl := template.New("Welcome")
  tpl = template.Must(tpl.ParseFiles("gotpl/index.gohtml","gotpl/menu.gohtml","gotpl/welcome.gohtml"))
  err := tpl.ExecuteTemplate(w, "layout", s)
  if err != nil {
      log.Fatalf("Template execution: %s", err)
  }
}

func (s *Site) AddPage(name, tpl, path, icon string, fn func(w http.ResponseWriter, r *http.Request)) {
  p := &Page{name,tpl,path,icon,s,template.New(name),fn}
  p.Tpl = template.Must(p.Tpl.ParseFiles("gotpl/index.gohtml","gotpl/menu.gohtml",tpl))
  s.Mux.Handle(p.Path, p)
  s.Pages = append([]*Page{p},s.Pages...)
}


func (s *Site) logPage(w http.ResponseWriter, r *http.Request) {
  log.Println("On Log Page")
}


func (s *Site) wsHandlerLogPage(w http.ResponseWriter, r *http.Request) {
	conn, err := s.logUpgraderWS.Upgrade(w, r, nil)
  ch := make(chan string)
  s.logChanWS = append(s.logChanWS,ch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  // Goroutine that close the websocket
  // if there is a error or a close request
	go func(conn *websocket.Conn) {
		for {
			mType, _, err := conn.ReadMessage()
			if err != nil || mType == websocket.CloseMessage {
				conn.Close()
				return
			}
		}
	}(conn)

	go func(conn *websocket.Conn) {
		for  {
			err := conn.WriteMessage(websocket.TextMessage, []byte(<-ch))
      if err != nil {
        conn.Close()
      }
		}

	}(conn)
}
