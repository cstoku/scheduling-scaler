package main

import (
	"flag"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
	"github.com/cstoku/scheduling-scaler/pkg/signals"

	kubeinformers "k8s.io/client-go/informers"
	clientset "github.com/cstoku/scheduling-scaler/pkg/client/clientset/versioned"
	informers "github.com/cstoku/scheduling-scaler/pkg/client/informers/externalversions"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	scscClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	scscInformerFactory := informers.NewSharedInformerFactory(scscClient, time.Second*30)

	controller := NewController(kubeClient, scscClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		scscInformerFactory.Scheduling().V1alpha1().SchedulingScalers())

	go kubeInformerFactory.Start(stopCh)
	go scscInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}