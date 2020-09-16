package utils

import(
  ke "../kubeedge"
  "k8s.io/client-go/rest"
  "context"
  "encoding/json"
  "log"
  "github.com/looplab/fsm"
  //"time"
)


type Device struct {
  Status ke.DeviceStatus `json:"status"`
  DeviceID string `json:"-"`
  Namespace string `json:"-"`
  crdClient *rest.RESTClient
  FSM *fsm.FSM `json:"-"`
  Filename string
  Url string
}



func (s* Device) InitDevice(id,ns string,crdClient *rest.RESTClient) error {
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
  s.Status = ke.DeviceStatus{}
  s.Status.Twins = twins
  s.DeviceID = id
  s.Namespace = ns
  s.crdClient = crdClient
  s.AddDesiredJob("Wait")
  s.AddDesiredArg("init")
  _,err := s.PatchStatus()
  if err != nil {
    return err
  }
  for s.GetStatus() != "Waiting" {
    s.SyncStatus()
    if err != nil {
      return err
    }
  }
  log.Println("Device Connected and Ready")
  s.FSM = fsm.NewFSM(
  		"ready",
  		fsm.Events{
  			{Name: "LaunchTask",       Src: []string{"ready","download"     }, Dst: "run"     },
  			{Name: "TaskCompleted",    Src: []string{"run"                  }, Dst: "done"},
        {Name: "FileNotFound",     Src: []string{"run"                  }, Dst: "download"},
        {Name: "DownloadError",    Src: []string{"download"             }, Dst: "error"   },
        {Name: "TaskError",        Src: []string{"run"                  }, Dst: "error"   },
        {Name: "Waiting",          Src: []string{"ready" ,"done"        }, Dst: "ready"   },
  		},
  		fsm.Callbacks{
        "LaunchTask": func(e *fsm.Event) {
          log.Printf("Launch app %s request",s.Filename)
          s.AddDesiredJob("Launch")
          s.AddDesiredArg(s.Filename)
          s.PatchStatus()
        },
        "TaskCompleted": func(e *fsm.Event) {
          s.AddDesiredJob("Wait")
          s.AddDesiredArg("Finished")
          s.PatchStatus()
        },
        "FileNotFound": func(e *fsm.Event) {
          log.Printf("File %s not found on Device %s",s.Filename,s.DeviceID)
          log.Printf("Download from ",s.Url)
          s.AddDesiredJob("Download")
          s.AddDesiredArg(s.Url)
          s.PatchStatus()
        },
      },
  	)
  return nil
}


func (s* Device) Launch(filename,url string){
  s.Filename = filename
  s.Url = url
  s.FSM.Event("LaunchTask")
  for s.FSM.Current() != "ready"{
    s.SyncStatus()
    s.FSM.Event(s.GetStatus())
    //time.Sleep(500*time.Millisecond)
  }
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

func (s* Device) PatchStatus()([]byte,error){
  ctx := context.Background()
  body, err := json.Marshal(s)
  if err != nil {
  	return nil,err
  }
  return s.crdClient.Patch(MergePatchType).Namespace(s.Namespace).Resource(ResourceTypeDevices).Name(s.DeviceID).Body(body).DoRaw(ctx)
}

func (s* Device) SyncStatus() error {
  ctx := context.Background()
	raw,err := s.crdClient.Get().Namespace(s.Namespace).Resource(ResourceTypeDevices).Name(s.DeviceID).DoRaw(ctx)
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
