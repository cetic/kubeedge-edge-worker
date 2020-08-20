package lib

import(
  mqtt "github.com/eclipse/paho.mqtt.golang"
  "github.com/goombaio/namegenerator"
  "time"
  "log"
)

type MQTT struct {
  BrokerHost string `yaml:"host"`
  BrokerPort string `yaml:"port"`
  Topic string `yaml:"topic"`
  Client mqtt.Client
  ID string
  Action func(mqtt.Message,mqtt.Client,string)
}


func (b *MQTT) Connect() {
  // Generate a unique ID of the device
  if b.ID == "" {
    nameGenerator := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())
    b.ID = nameGenerator.Generate()
  }
  //Create the MQTT client
  opts := mqtt.NewClientOptions().AddBroker(b.BrokerHost+":"+b.BrokerPort).SetClientID(b.ID)
  b.Client = mqtt.NewClient(opts)
  // Connection to the MQTT Broker
  if token := b.Client.Connect(); token.Wait() && token.Error() != nil {
     log.Fatal(token.Error())
  }
  log.Printf("Connected to the broker %s:%s",b.BrokerHost,b.BrokerPort)
}

func (b *MQTT) Subscribe(channel string) {
  if token := b.Client.Subscribe(channel, 0, b.MqttHandlerJSON(channel)); token.Wait() && token.Error() != nil {
     log.Fatal(token.Error())
  }
  log.Printf("Subscribe to %s",channel)
}

// Action made when the device receive a message from the specified Topic
func (b *MQTT) MqttHandlerJSON(channel string) func(client mqtt.Client, message mqtt.Message) {
  var msgRcvd mqtt.MessageHandler = func(client mqtt.Client, message mqtt.Message) {
    b.Action(message,client,channel)
    }
  return msgRcvd
}
