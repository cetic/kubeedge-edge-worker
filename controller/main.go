package main

import (
  "./utils"
  "time"
)

func main(){
  //init a CRDClient
  //First Parameter is the ip of the Kubernetes Master
  //Second paramater is the path to the kubeconfig Path (Let empty if the controller is inside the cluster)
  crdClient, _ := utils.NewCRDClient("","")1
  //Create the device.
  //First parmater is the Id of device, it must be the same as kubeedge use.
  //Second parameter is the Namespace where is the device.
  d := utils.Device{}.InitDevice("edge-dev-01","default",crdClient)

  filename := "/app/hello-loop.py"
  url := "https://raw.githubusercontent.com/Sellto/kubeedge-edge-worker/master/examples/hello-loop.py"
  //Launch function is used to launche application on the Edged device.
  //First paramater is the filename path of the application
  //Second parameter is the url where the file can be download (if not present on the device)
  d.Launch(filename,url)
  //This thread is used to kill the running app after 20 Seconds
  go func(){
    time.Sleep(20*time.Second)
    d.AddDesiredJob("Stop")
    d.PatchStatus()
  }()
}
