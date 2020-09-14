package lib

import(
  "log"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "os/signal"
  "syscall"
  "os"
  "os/exec"
  mqtt "github.com/eclipse/paho.mqtt.golang"
  "encoding/json"
  "strings"
  //"time"
  "github.com/kubeedge/kubeedge/edge/pkg/devicetwin/dttype"
  //"github.com/kubeedge/kubeedge/cloud/pkg/devicecontroller/types"
)

type Device struct {
  Type string `yaml:"type"`
  Model string `yaml:"model"`
  Launcher string `yaml:"launchers"`
  MQTT MQTT `yaml:"broker"`
  Path string `yaml:"sourcepath"`
  DeviceID string `yaml:"DeviceID"`
  status string
  job string
  arg string
}


const (
	DeviceETPrefix            = "$hw/events/device/"
	TwinETUpdateSuffix        = "/twin/update"
	TwinETUpdateDeltaSuffix   = "/twin/update/delta"
)


func (d *Device) GetConfigFromFile(filename string){
  // Read file
  yamlFile, err := ioutil.ReadFile(filename)
  if err != nil {
    log.Println(err)
  }
  // Parse file
  err = yaml.Unmarshal(yamlFile, &d)
  if err != nil {
     log.Fatal(err)
  }
}


func (d *Device) Listen() {
	c := make(chan os.Signal)
  d.MQTT.Action = d.action
  d.MQTT.Connect()
  //Subscribe to the update/delta topic
  d.MQTT.Subscribe(DeviceETPrefix+d.DeviceID+TwinETUpdateDeltaSuffix)
  //d.sendStatus()
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
  // Goroutine that handle the Keyboard Interrupt
	go func() {
		<-c
		log.Println("\r- Ctrl+C pressed in Terminal")
    // Disconnect to the MQTT Broker
    d.MQTT.Client.Disconnect(250)
    // Close Application
		os.Exit(0)
	}()
  // Still App Open

  for{}
}


func (d *Device) action(m mqtt.Message, client mqtt.Client,channel string) {
  if channel == DeviceETPrefix+d.DeviceID+TwinETUpdateDeltaSuffix {
    msg := &dttype.DeviceTwinUpdate{}
    //Parse the incoming message to a Message struct
    err := json.Unmarshal(m.Payload(), msg)
    if err != nil {
      log.Printf("Bad Format message : %s",err)
    }
    // Get the expected value from cloud
    expectedJob := *(msg.Twin["job"].Expected.Value)
    expectedArg := *(msg.Twin["arg"].Expected.Value)
    if ((expectedArg != d.arg)||(expectedJob != d.job)) {
      d.job = expectedJob
      d.arg = expectedArg
      switch d.job {
      case "Launch":
        d.run()
      case "Download":
        d.download()
      case "Wait":
        d.sendTwinActualValue("Waiting")
      default:
        d.sendTwinActualValue("notavailable")
      }
    }
  }
}



func (d *Device) download() {
    log.Printf("Download of file request : %s",d.arg)
    //try to download the file with wget tools
    _, err := exec.Command("wget","-P",d.Path,d.arg).Output()
  	if err != nil {
      d.sendTwinActualValue("DownloadError")
  		log.Println(err)
  	} else {
        d.sendTwinActualValue("DownloadComplete")
    }
}


func (d *Device) run() {
  log.Printf("Run application request")
  args := strings.Split(d.arg," ")
  out, err := exec.Command(d.Launcher,args...).Output()
  if err != nil {
      d.sendTwinActualValue("FileNotFound")
      log.Println(err)
  } else {
     d.sendTwinActualValue("TaskCompleted")
     log.Printf("Execution output: %s",out)
  }
}



func (d *Device) sendTwinActualValue(status string) {
	var deviceTwinUpdateMessage dttype.DeviceTwinUpdate
  // Create Twin actual value
	actualMap := map[string]*dttype.MsgTwin{
    "status": {Actual: &dttype.TwinValue{Value: &status},   Metadata: &dttype.TypeMetadata{Type: "Updated"}},
    "job":    {Actual: &dttype.TwinValue{Value: &d.job},    Metadata: &dttype.TypeMetadata{Type: "Updated"}},
    "arg":    {Actual: &dttype.TwinValue{Value: &d.arg},    Metadata: &dttype.TypeMetadata{Type: "Updated"}},
  }
	deviceTwinUpdateMessage.Twin = actualMap
  // Convert message to Json
  twinUpdateBody, err := json.Marshal(deviceTwinUpdateMessage)
  if err != nil {
    log.Printf("Can't parse message to JSON : %s",err)
  }
  // Publish message to MQTT
  if token := d.MQTT.Client.Publish(DeviceETPrefix+d.DeviceID+TwinETUpdateSuffix, 0, false, twinUpdateBody); token.Wait() && token.Error() != nil {
           log.Fatal(token.Error())
  }
}
