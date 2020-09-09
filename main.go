package main

import(
   lib "./lib"
  "fmt"
)

func main(){
  device := lib.Device{}
  device.GetConfigFromFile("/app/config/config.yaml")
  device.Listen()
  fmt.Println("hello")
}
