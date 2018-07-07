package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/cstoku/scheduling-scaler/pkg/apis/scsc"
	"k8s.io/apimachinery/pkg/runtime"
)

var SchemeGroupVersion = schema.GroupVersion{Group: scsc.GroupName, Version: "v1alpha"}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SchedulingScaler{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
