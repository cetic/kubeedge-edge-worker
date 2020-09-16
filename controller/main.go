package main

import (
  "./utils"
  //ke "./kubeedge"
  //"github.com/looplab/fsm"
)

func main(){
  crdClient, _ := utils.NewCRDClient("https://10.129.1.26:6443","/Users/tse/.kube/config")
  //log.Println(crdClient)
  var d utils.Device
  d.InitDevice("edge-worker-01","default",crdClient)
  filename := "hello.py"
  url := "https://raw.githubusercontent.com/Sellto/kubeedge-edge-worker/master/examples/hello.py"
  d.Launch(filename,url)
}
