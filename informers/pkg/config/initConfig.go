package config

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func InitConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeConfig := flag.String("kubeconfig", "/Users/xin/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
		flag.Parse()
		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func InitClient() *kubernetes.Clientset {
	config, err := InitConfig()
	if err != nil {
		panic(err.Error())
	}
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return c
}
