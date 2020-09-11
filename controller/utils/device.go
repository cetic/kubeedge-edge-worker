package utils

import(
  ke "../kubeedge"
  "k8s.io/client-go/rest"
  "context"
  "encoding/json"
  //"log"
)


type Device struct {
  Status ke.DeviceStatus `json:"status"`
  DeviceID string `json:"-"`
  Namespace string `json:"-"`
}



func CreateDevice(id,ns string) Device {
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
  return s
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
