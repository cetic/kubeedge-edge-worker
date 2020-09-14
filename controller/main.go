package main

import (
  "./utils"
  //ke "./kubeedge"
  //"github.com/looplab/fsm"
)

func main(){
  crdClient, _ := utils.NewCRDClient("https://192.168.0.102:6443","/Users/tse/.kube/config")
  //log.Println(crdClient)
  d,_ := utils.CreateDevice("edge-worker-01","default",crdClient)
  d.Launch("main.py","")
}
