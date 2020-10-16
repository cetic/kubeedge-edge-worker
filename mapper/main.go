package main

import(
   lib "./lib"
   "os"
)

func main(){
  device := lib.Device{}
  device.GetConfigFromFile(os.Args[1])
  device.Listen()
}
