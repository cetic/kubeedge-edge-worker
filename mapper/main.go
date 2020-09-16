package main

import(
   lib "./lib"
  //"fmt"
)

func main(){
  device := lib.Device{}
  device.GetConfigFromFile("config/config.yaml")
  device.Listen()
}
