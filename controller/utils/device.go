package utils

import(
  ke "../kubeedge"
  "k8s.io/client-go/rest"
  "context"
  "encoding/json"
  "log"
  "github.com/looplab/fsm"
  "time"
)


type Device struct {
  Status ke.DeviceStatus `json:"status"`
  DeviceID string `json:"-"`
  Namespace string `json:"-"`
  crdClient *rest.RESTClient
  FSM *fsm.FSM `json:"-"`
}



func CreateDevice(id,ns string,crdClient *rest.RESTClient) (Device,error) {
  metadata := map[string]string{"timestamp": GetTimeStamp(),"type": "string"}
  twins := []ke.Twin{
    { PropertyName: "job",
      Desired:  ke.TwinProperty{ Value: "unknown", Metadata: metadata},
      Reported: ke.TwinProperty{ Value: "unknown", Metadata: metadata},
    },
    { PropertyName: "arg",
      Desired:  ke.TwinProperty{Value: "unknown", Metadata: metadata},
      Reported: ke.TwinProperty{Value: "unknown", Metadata: metadata},
    },
    { PropertyName: "status",
      Desired:  ke.TwinProperty{Value: "unknown", Metadata: metadata},
    },
  }
  s := Device{Status:ke.DeviceStatus{}}
  s.Status.Twins = twins
  s.DeviceID = id
  s.Namespace = ns
  s.crdClient = crdClient
  s.AddDesiredJob("Wait")
  s.AddDesiredArg("init")
  _,err := s.PatchStatus(s.crdClient)
  if err != nil {
    return s,err
  }
  for s.GetStatus() != "Waiting" {
    s.SyncStatus(crdClient)
    if err != nil {
      return s,err
    }
  }
  log.Println("Device Connected and Ready")
  s.FSM = fsm.NewFSM(
  		"ready",
  		fsm.Events{
  			{Name: "LaunchTask",       Src: []string{"ready"     }, Dst: "run"     },
  			{Name: "TaskCompleted",     Src: []string{"run","Download"}, Dst: "done"    },
        {Name: "FileNotFound",     Src: []string{"run"       }, Dst: "download"},
        {Name: "Finishing",        Src: []string{"done"      }, Dst: "ready"   },
        {Name: "DownloadError",    Src: []string{"download"  }, Dst: "error"   },
        {Name: "TaskError",        Src: []string{"run"       }, Dst: "error"   },
        {Name: "Waiting",          Src: []string{"ready" ,"done"}, Dst: "ready"   },
  		},
  		fsm.Callbacks{
        "TaskCompleted": func(e *fsm.Event) {
          s.AddDesiredJob("Wait")
          s.AddDesiredArg("finished")
          s.PatchStatus(s.crdClient)
          for s.GetStatus() != "Waiting"{
            s.SyncStatus(s.crdClient)
            time.Sleep(1*time.Second)
          }
        },
      },
  	)
  return s,nil
}


func (s* Device) Launch(filename,url string){
  log.Println("Launch app request")
  s.FSM.Event("LaunchTask")
  s.AddDesiredJob("Launch")
  s.AddDesiredArg(filename)
  s.PatchStatus(s.crdClient)
  for s.FSM.Current()!="ready"{
    s.SyncStatus(s.crdClient)
    s.FSM.Event(s.GetStatus())
    if s.GetStatus() == "FileNotFound" {
      s.AddDesiredJob("Download")
      s.AddDesiredArg(url)
      s.PatchStatus(s.crdClient)
      for s.FSM.Current()!="ready"{
        s.SyncStatus(s.crdClient)
        s.FSM.Event(s.GetStatus())
      }  
    }
    time.Sleep(1*time.Second)
  }
  log.Println("app launched")
}

func (s* Device) AddDesiredJob(job string){
    for id,property := range(s.Status.Twins){
      if property.PropertyName == "job" {
        s.Status.Twins[id].Desired.Value = job
      }
    }
}

func (s* Device) AddDesiredArg(arg string){
    for id,property := range(s.Status.Twins){
      if property.PropertyName == "arg" {
        s.Status.Twins[id].Desired.Value = arg
      }
    }
}

func (s* Device) PatchStatus(crdClient *rest.RESTClient)([]byte,error){
  ctx := context.Background()
  body, err := json.Marshal(s)
  if err != nil {
  	return nil,err
  }
  return crdClient.Patch(MergePatchType).Namespace(s.Namespace).Resource(ResourceTypeDevices).Name(s.DeviceID).Body(body).DoRaw(ctx)
}

func (s* Device) SyncStatus(crdClient *rest.RESTClient) error {
  ctx := context.Background()
	raw,err := crdClient.Get().Namespace(s.Namespace).Resource(ResourceTypeDevices).Name(s.DeviceID).DoRaw(ctx)
  _ = json.Unmarshal(raw, &s)
  return err
}

func (s* Device) GetStatus() string {
  for id,property := range(s.Status.Twins){
    if property.PropertyName == "status" {
      return s.Status.Twins[id].Reported.Value
    }
  }
  return "unknown"
}
