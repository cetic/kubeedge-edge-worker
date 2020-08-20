package main

import(
  lib "./lib"
)

func main(){
  device := lib.Device{}
  device.GetConfigFromFile("config.yaml")
  device.Listen()
}
