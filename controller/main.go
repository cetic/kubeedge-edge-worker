package main

import(
  "net/http"
  //"fmt"
  "log"
  //"reflect"
  "time"
  "./lib"
   "./utils"
   "os"
)


func Facetrack(w http.ResponseWriter, r *http.Request) {
  log.Println(d.GetStatus())
  if d.FSM.Current() == "run" {
    d.AddDesiredJob("Stop")
    d.PatchStatus()
    for d.FSM.Current() != "ready" {
      time.Sleep(500*time.Millisecond)
    }
  }

  filename := "/app/dpu_face_tracking.py"
  url := ""
  go d.Launch(filename,url)
}

func Passthrough(w http.ResponseWriter, r *http.Request) {
  if d.FSM.Current() == "run" {
    d.AddDesiredJob("Stop")
    d.PatchStatus()
    for d.FSM.Current() != "ready" {
      time.Sleep(500*time.Millisecond)
    }
  }

  filename := "/app/passthrough.py"
  url := ""
  go d.Launch(filename,url)
}


func Stop(w http.ResponseWriter, r *http.Request) {
  if d.FSM.Current() == "run" {
    d.AddDesiredJob("Stop")
    d.PatchStatus()
    for d.FSM.Current() != "ready" {
      time.Sleep(500*time.Millisecond)
    }
  }
}

var d = utils.Device{}


func main() {
  crdClient, _ := utils.NewCRDClient(os.Args[1],os.Args[2])
  d.InitDevice(os.Args[3],"default",crdClient)
  s := new(web.Site)
  s.Init()
  s.AddPage("Home","gotpl/welcome.gohtml","/passthrough","home", Passthrough)
  s.AddPage("Home","gotpl/welcome.gohtml","/facetrack","home", Facetrack)
  s.AddPage("Home","gotpl/welcome.gohtml","/stop","home", Stop)
  http.Handle("/", s.Mux)
  http.ListenAndServe(":9090", nil)
}
