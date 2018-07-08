package main

import (
	"flag"
	"time"

	"github.com/cstoku/scheduling-scaler/pkg/signals"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/cstoku/scheduling-scaler/pkg/apis/scaling"
	clientset "github.com/cstoku/scheduling-scaler/pkg/client/clientset/versioned"
	informers "github.com/cstoku/scheduling-scaler/pkg/client/informers/externalversions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
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

	scalingClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building %s clientset: %s", scaling.GroupName, err.Error())
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building discovery client: %s", err.Error())
	}

	resolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	ar, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		glog.Fatalf("Error get apigroup resources: %s", err.Error())
	}
	mapper := restmapper.NewDiscoveryRESTMapper(ar)
	scaleNamespacer, err := scale.NewForConfig(cfg, mapper, dynamic.LegacyAPIPathResolverFunc, resolver)
	if err != nil {
		glog.Fatalf("Error binding scale client: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	appsInformerFactory := informers.NewSharedInformerFactory(scalingClient, time.Second*30)

	controller := NewController(kubeClient, scalingClient, mapper, scaleNamespacer,
		appsInformerFactory.Scaling().V1alpha1().SchedulingScalers())

	go kubeInformerFactory.Start(stopCh)
	go appsInformerFactory.Start(stopCh)

	if err = controller.Run(stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}
