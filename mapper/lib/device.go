package lib

import(
  "log"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "os/signal"
  "syscall"
  "os"
  "os/exec"
  "bufio"
  "io"
  mqtt "github.com/eclipse/paho.mqtt.golang"
  "encoding/json"
  "strings"
  "github.com/kubeedge/kubeedge/edge/pkg/devicetwin/dttype"
)

// Device Type
type Device struct {
  Type string `yaml:"type"`
  Model string `yaml:"model"`
  Launcher string `yaml:"launchers"`
  MQTT MQTT `yaml:"broker"`
  Path string `yaml:"sourcepath"`
  DeviceID string `yaml:"DeviceID"`
  // Channel used to kill a running app
  Stop chan string
  // Twin Parameters
  status string
  job string
  arg string
}


const (
	DeviceETPrefix            = "$hw/events/device/"
	TwinETUpdateSuffix        = "/twin/update"
	TwinETUpdateDeltaSuffix   = "/twin/update/delta"
)


// Function used to parse a yaml file to a Device configuration
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


//Function that initialise the Mapper and keep it open.
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


// Function called by the mqtt handler.
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
        log.Println("Waiting request")
        d.sendTwinActualValue("Waiting")
      case "Stop":
        log.Println("Stop request")
        d.Stop <- "Stop"
        d.sendTwinActualValue("TaskCompleted")
      default:
        d.sendTwinActualValue("notavailable")
      }
    }
  }
}


// Function used to download a file
func (d *Device) download() {
    log.Printf("Download of file request : %s",d.arg)
    //try to download the file with wget tools
    _, err := exec.Command("wget","-P",d.Path,d.arg).Output()
  	if err != nil {
      d.sendTwinActualValue("DownloadError")
  		log.Println(err)
  	} else {
        d.sendTwinActualValue("LaunchTask")
    }
}


// Function used to grep the log of the application launched
func copyOutput(r io.Reader) {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        log.Printf("App : %s",scanner.Text())
    }
}

func (d *Device) run() {
  log.Printf("Run application request")
  // Create the Command that will be launched
  args := strings.Split(d.arg," ")
  cmd:= exec.Command(d.Launcher,args...)
  stdout, err := cmd.StdoutPipe()
  if err != nil {
      panic(err)
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
      panic(err)
  }
  err = cmd.Start()
  if err != nil {
      panic(err)
  }

  //Threads that send the output of the Application to the standard output
  go copyOutput(stdout)

  //Threads that send the error of the Application to the standard error
  go copyOutput(stderr)

  done := make(chan error, 1)
  d.Stop = make(chan string)

  // Thread that triggers the end of the application
  go func() {
      done <- cmd.Wait()
  }()

  // Thread that triggers the end of the launched application
  // Either because there is a external request
  // Either because the process is correctly terminated
  go func(){
    select {
    case msg := <-d.Stop:
        log.Println(msg)
        if err := cmd.Process.Kill(); err != nil {
            log.Fatal("failed to kill process: ", err)
        }
        log.Println("process manually terminated")
        d.sendTwinActualValue("TaskCompleted")
    case err := <-done:
        if err != nil {
            log.Println("Process Terminated")
        }
        log.Print("process finished successfully")
        d.sendTwinActualValue("TaskCompleted")
    }
  }()
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
