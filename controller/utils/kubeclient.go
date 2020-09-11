package utils

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8SConfig struct {
	MasterAdress string
	KubeConfigPath string
	QPS float32
	Burst int
	ContentType string
}

func NewK8SConfig(MasterAdress,KubeConfigPath string) K8SConfig {
	config := K8SConfig{}
	config.Burst = 10
	config.QPS = 5.000000
	config.ContentType = "application/vnd.kubernetes.protobuf"
	config.MasterAdress = MasterAdress
	config.KubeConfigPath = KubeConfigPath
	return config
}

func (k *K8SConfig) Generate() (conf *rest.Config, err error){
	kubeConfig, err := clientcmd.BuildConfigFromFlags(k.MasterAdress, k.KubeConfigPath)
	if err != nil {
		return nil, err
	}
	kubeConfig.QPS = k.QPS
	kubeConfig.Burst = k.Burst
	kubeConfig.ContentType = k.ContentType
	return kubeConfig, err
}
