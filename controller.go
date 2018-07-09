package main

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"sort"

	scalingv1alpha1 "github.com/cstoku/scheduling-scaler/pkg/apis/scaling/v1alpha1"
	clientset "github.com/cstoku/scheduling-scaler/pkg/client/clientset/versioned"
	scalingscheme "github.com/cstoku/scheduling-scaler/pkg/client/clientset/versioned/scheme"
	informers "github.com/cstoku/scheduling-scaler/pkg/client/informers/externalversions/scaling/v1alpha1"
	listers "github.com/cstoku/scheduling-scaler/pkg/client/listers/scaling/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const controllerAgentName = "scheduling-scaler-controller"

const (
	SuccessSynced = "Synced"

	MessageResourceSynced = "Workflow synced successfully"
)

type Controller struct {
	kubeClientset kubernetes.Interface
	scalingClientset clientset.Interface

	mapper          apimeta.RESTMapper
	scaleNamespacer scale.ScalesGetter

	appsLister listers.SchedulingScalerLister
	appsSynced cache.InformerSynced

	recorder record.EventRecorder
}

func NewController(
	kubeClientset kubernetes.Interface,
	scalingClientset clientset.Interface,
	mapper apimeta.RESTMapper,
	scaleNamespacer scale.ScalesGetter,
	appsInformer informers.SchedulingScalerInformer) *Controller {
	scalingscheme.AddToScheme(scheme.Scheme)

	glog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClientset:    kubeClientset,
		scalingClientset: scalingClientset,
		mapper:           mapper,
		scaleNamespacer:  scaleNamespacer,
		appsLister:       appsInformer.Lister(),
		appsSynced:       appsInformer.Informer().HasSynced,
		recorder:         recorder,
	}

	return controller
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()

	glog.Info("Starting Workflow controller")
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.appsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting SchedulingScaler Manager")
	go wait.Until(c.syncAll, 10*time.Second, stopCh)

	glog.Info("Started SchedulingScaler Manager")
	<-stopCh
	glog.Info("Shutting down SchedulingScaler Manager")

	return nil
}

func (c *Controller) syncAll() {
	scalerList, err := c.scalingClientset.ScalingV1alpha1().SchedulingScalers(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		runtime.HandleError(fmt.Errorf("can't list SchedulingScalers: %v", err))
		return
	}
	scalers := scalerList.Items
	glog.V(4).Infof("Found %d SchedulingScalers", len(scalers))

	for _, scaler := range scalers {
		err := c.syncOne(&scaler, metav1.Now())
		if err != nil {
			glog.Warning(fmt.Errorf("SyncOne Error: %v", err))
		}
	}
}

func (c *Controller) syncOne(scaler *scalingv1alpha1.SchedulingScaler, now metav1.Time) error {
	reference := fmt.Sprintf("%s/%s/%s", scaler.Spec.ScaleTargetRef.Kind, scaler.Namespace, scaler.Spec.ScaleTargetRef.Name)
	scalerStatusOriginal := scaler.Status.DeepCopy()

	targetGV, err := schema.ParseGroupVersion(scaler.Spec.ScaleTargetRef.APIVersion)
	if err != nil {
		c.recorder.Event(scaler, corev1.EventTypeWarning, "FailedGetScale", err.Error())
		c.updateStatusIfNeeded(scalerStatusOriginal, scaler)
		return fmt.Errorf("invalid API version in scale target reference: %v", err)
	}

	targetGK := schema.GroupKind{
		Group: targetGV.Group,
		Kind:  scaler.Spec.ScaleTargetRef.Kind,
	}

	mappings, err := c.mapper.RESTMappings(targetGK)
	if err != nil {
		c.recorder.Event(scaler, corev1.EventTypeWarning, "FailedGetScale", err.Error())
		c.updateStatusIfNeeded(scalerStatusOriginal, scaler)
		return fmt.Errorf("unable to determine resource for scale target reference: %v", err)
	}

	scale, targetGR, err := c.scaleForResourceMappings(scaler.Namespace, scaler.Spec.ScaleTargetRef.Name, mappings)
	if err != nil {
		c.recorder.Event(scaler, corev1.EventTypeWarning, "FailedGetScale", err.Error())
		c.updateStatusIfNeeded(scalerStatusOriginal, scaler)
		return fmt.Errorf("failed to query scale subresource for %s: %v", reference, err)
	}
	currentReplicas := scale.Status.Replicas

	desiredReplicas := int32(0)
	rescaleReason := ""
	rescale := true

	if scale.Spec.Replicas == 0 {
		desiredReplicas = 0
		rescale = false
	} else if currentReplicas == 0 {
		rescaleReason = "Current number of replicas must be greater than 0"
		desiredReplicas = 1
	} else {
		scalerSchedules := scaler.Spec.Schedules
		sortSchedules(scalerSchedules)
		targetSchedule := scalerSchedules[len(scalerSchedules)-1]
		for _, s := range scalerSchedules {
			nowTime := convertTime(s.ScheduleTime, now)
			glog.V(4).Infof("Check Schedule Time: %v <-> %v(%v)", s.ScheduleTime, nowTime, nowTime.Local())
			if s.ScheduleTime.Equal(&nowTime) || s.ScheduleTime.After(nowTime.Local()) {
				break
			}
			targetSchedule = s
		}
		glog.V(4).Infof("Target Schedule: %v -> %v", targetSchedule.ScheduleTime, targetSchedule.Replicas)
		rescaleReason = "Scheduling time has changed"
		desiredReplicas = targetSchedule.Replicas
		rescale = desiredReplicas != currentReplicas
	}

	if rescale {
		scale.Spec.Replicas = desiredReplicas
		_, err = c.scaleNamespacer.Scales(scaler.Namespace).Update(targetGR, scale)
		if err != nil {
			c.recorder.Eventf(scaler, corev1.EventTypeWarning, "FailedRescale", "New size: %d; reason: %s; error: %v", desiredReplicas, rescaleReason, err.Error())
			c.setCurrentReplicasInStatus(scaler, currentReplicas)
			if err := c.updateStatusIfNeeded(scalerStatusOriginal, scaler); err != nil {
				runtime.HandleError(err)
			}
			return fmt.Errorf("failed to rescale %s: %v", reference, err)
		}
		c.recorder.Eventf(scaler, corev1.EventTypeNormal, "SuccessfulRescale", "New size: %d; reason: %s", desiredReplicas, rescaleReason)
		glog.Infof("Successful rescale of %s, old size: %d, new size: %d, reason: %s",
			scaler.Name, currentReplicas, desiredReplicas, rescaleReason)
	} else {
		glog.V(4).Infof("decided not to scale %s to %v (last scale time was %s)", reference, desiredReplicas, scaler.Status.LastScaleTime)
		desiredReplicas = currentReplicas
	}

	c.recorder.Event(scaler, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)

	c.setStatus(scaler, currentReplicas, desiredReplicas, &now, rescale)
	return c.updateStatusIfNeeded(scalerStatusOriginal, scaler)
}

func (c *Controller) scaleForResourceMappings(namespace, name string, mappings []*apimeta.RESTMapping) (*autoscalingv1.Scale, schema.GroupResource, error) {
	var firstErr error
	for i, mapping := range mappings {
		targetGR := mapping.Resource.GroupResource()
		scale, err := c.scaleNamespacer.Scales(namespace).Get(targetGR, name)
		if err == nil {
			return scale, targetGR, nil
		}

		if i == 0 {
			firstErr = err
		}
	}

	if firstErr == nil {
		firstErr = fmt.Errorf("unrecognized resource")
	}

	return nil, schema.GroupResource{}, firstErr
}

func (c *Controller) setCurrentReplicasInStatus(scaler *scalingv1alpha1.SchedulingScaler, currentReplicas int32) {
	c.setStatus(scaler, currentReplicas, scaler.Status.DesiredReplicas, nil, false)
}

func (c *Controller) setStatus(scaler *scalingv1alpha1.SchedulingScaler, currentReplicas, desiredReplias int32, now *metav1.Time, rescale bool) {
	scaler.Status = scalingv1alpha1.SchedulingScalerStatus{
		CurrentReplicas: currentReplicas,
		DesiredReplicas: desiredReplias,
		LastScaleTime:   scaler.Status.LastScaleTime,
	}

	if rescale && now != nil {
		scaler.Status.LastScaleTime = *now
	}
}

func (c *Controller) updateStatusIfNeeded(oldStatus *scalingv1alpha1.SchedulingScalerStatus, newScaler *scalingv1alpha1.SchedulingScaler) error {
	if apiequality.Semantic.DeepEqual(oldStatus, &newScaler.Status) {
		return nil
	}
	return c.updateStatus(newScaler)
}

func (c *Controller) updateStatus(scaler *scalingv1alpha1.SchedulingScaler) error {
	_, err := c.scalingClientset.ScalingV1alpha1().SchedulingScalers(scaler.Namespace).Update(scaler)
	if err != nil {
		c.recorder.Event(scaler, corev1.EventTypeWarning, "FailedUpdateStatus", err.Error())
		return fmt.Errorf("failed to update status for %s: %v", scaler.Name, err)
	}
	glog.V(2).Infof("Successfully updated status for %s", scaler.Name)
	return nil
}

func sortSchedules(st []scalingv1alpha1.SchedulingScalerSchedule) {
	sort.Slice(st, func(i, j int) bool {
		return st[i].ScheduleTime.Before(&metav1.Time{st[j].ScheduleTime.Local()})
	})
}

func convertTime(st scalingv1alpha1.SchedulingTime, now metav1.Time) metav1.Time {
	y, m, d := st.Date()
	h, min, s := now.Clock()
	return metav1.Date(y, m, d, h, min, s, now.Nanosecond(), st.Location())
}
