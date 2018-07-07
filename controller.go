package main

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	corev1 "k8s.io/api/core/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	appsv1alpha1 "github.com/cstoku/scheduling-scaler/pkg/apis/apps/v1alpha1"
	clientset "github.com/cstoku/scheduling-scaler/pkg/client/clientset/versioned"
	appsscheme "github.com/cstoku/scheduling-scaler/pkg/client/clientset/versioned/scheme"
	informers "github.com/cstoku/scheduling-scaler/pkg/client/informers/externalversions/apps/v1alpha1"
	listers "github.com/cstoku/scheduling-scaler/pkg/client/listers/apps/v1alpha1"
)

const controllerAgentName = "scheduling-scaler-controller"

const (
	SuccessSynced = "Synced"

	MessageResourceSynced = "Workflow synced successfully"
)

type Controller struct {
	kubeClientset kubernetes.Interface
	appsClientset clientset.Interface

	scaleNamespacer scale.ScalesGetter

	appsLister listers.SchedulingScalerLister
	appsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

func NewController(
	kubeClientset kubernetes.Interface,
	appsClientset clientset.Interface,
	scaleNamespacer scale.ScalesGetter,
	appsInformer informers.SchedulingScalerInformer) *Controller {
	kubeClientset.AutoscalingV1().RESTClient()
	appsscheme.AddToScheme(scheme.Scheme)
	glog.Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClientset:   kubeClientset,
		appsClientset:   appsClientset,
		scaleNamespacer: scaleNamespacer,
		appsLister:      appsInformer.Lister(),
		appsSynced:      appsInformer.Informer().HasSynced,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "appss"),
		recorder:        recorder,
	}
	scaleNamespacer.Scales("")
	glog.Info("Setting up event handlers")
	appsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueSchedulingScaler,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueSchedulingScaler(new)
		},
	})
	return controller
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	glog.Info("Starting Workflow controller")

	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.appsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		c.workqueue.AddRateLimited(obj)
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	wf, err := c.appsLister.SchedulingScalers(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("workflow '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	wfName := wf.Spec.Name
	if wfName == "" {
		runtime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}

	err = c.updateWorkflowStatus(wf)
	if err != nil {
		return err
	}

	c.recorder.Event(wf, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateWorkflowStatus(wf *appsv1alpha1.SchedulingScaler) error {
	wfCopy := wf.DeepCopy()
	wfCopy.Status.Name = wf.Spec.Name
	_, err := c.appsClientset.AppsV1alpha1().SchedulingScalers(wf.Namespace).Update(wfCopy)
	return err
}

func (c *Controller) enqueueSchedulingScaler(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}
