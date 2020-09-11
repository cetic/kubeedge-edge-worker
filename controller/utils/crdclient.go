package utils

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	ke "../kubeedge"
)

// NewCRDClient is used to create a restClient for crd
func NewCRDClient(MasterAdress,KubeConfigPath string) (*rest.RESTClient, error) {
	kubeconf := NewK8SConfig(MasterAdress,KubeConfigPath)
	cfg, err := kubeconf.Generate()
	if err != nil {
		return nil, err
	}
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(AddDeviceCrds)

	err = schemeBuilder.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	config := *cfg
	config.APIPath = "/apis"
	config.GroupVersion = &ke.SchemeGroupVersion
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.NewCodecFactory(scheme).WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		klog.Errorf("Failed to create REST Client due to error %v", err)
		return nil, err
	}

	return client, nil
}

func AddDeviceCrds(scheme *runtime.Scheme) error {
	// Add Device
	scheme.AddKnownTypes(ke.SchemeGroupVersion, &ke.Device{}, &ke.DeviceList{})
	v1.AddToGroupVersion(scheme, ke.SchemeGroupVersion)
	// Add DeviceModel
	scheme.AddKnownTypes(ke.SchemeGroupVersion, &ke.DeviceModel{}, &ke.DeviceModelList{})
	v1.AddToGroupVersion(scheme, ke.SchemeGroupVersion)

	return nil
}
