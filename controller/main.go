package main

import (
  "./utils"
  "log"
  "time"
  //ke "./kubeedge"
)



func main(){
  crdClient, _ := utils.NewCRDClient("https://192.168.0.102:6443","/Users/tse/.kube/config")
  //log.Println(crdClient)
  d := utils.CreateDevice("edge-worker-01","default")
  d.AddDesiredJob("Wait")
  d.AddDesiredArg("init")
  _,err := d.PatchStatus(crdClient)
  if err != nil {
    log.Println(err)
  }
  for d.GetStatus() != "ready" {
    log.Println(d.GetStatus())
    d.SyncStatus(crdClient)
    time.Sleep(1*time.Second)
  }
}
