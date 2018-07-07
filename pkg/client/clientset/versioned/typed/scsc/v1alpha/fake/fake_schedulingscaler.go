/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha "github.com/cstoku/scheduling-scaler/pkg/apis/scsc/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSchedulingScalers implements SchedulingScalerInterface
type FakeSchedulingScalers struct {
	Fake *FakeSchedulingV1alpha
	ns   string
}

var schedulingscalersResource = schema.GroupVersionResource{Group: "scheduling.scaler.cstoku.me", Version: "v1alpha1", Resource: "schedulingscalers"}

var schedulingscalersKind = schema.GroupVersionKind{Group: "scheduling.scaler.cstoku.me", Version: "v1alpha1", Kind: "SchedulingScaler"}

// Get takes name of the schedulingScaler, and returns the corresponding schedulingScaler object, and an error if there is any.
func (c *FakeSchedulingScalers) Get(name string, options v1.GetOptions) (result *v1alpha.SchedulingScaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(schedulingscalersResource, c.ns, name), &v1alpha.SchedulingScaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.SchedulingScaler), err
}

// List takes label and field selectors, and returns the list of SchedulingScalers that match those selectors.
func (c *FakeSchedulingScalers) List(opts v1.ListOptions) (result *v1alpha.SchedulingScalerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(schedulingscalersResource, schedulingscalersKind, c.ns, opts), &v1alpha.SchedulingScalerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha.SchedulingScalerList{ListMeta: obj.(*v1alpha.SchedulingScalerList).ListMeta}
	for _, item := range obj.(*v1alpha.SchedulingScalerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested schedulingScalers.
func (c *FakeSchedulingScalers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(schedulingscalersResource, c.ns, opts))

}

// Create takes the representation of a schedulingScaler and creates it.  Returns the server's representation of the schedulingScaler, and an error, if there is any.
func (c *FakeSchedulingScalers) Create(schedulingScaler *v1alpha.SchedulingScaler) (result *v1alpha.SchedulingScaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(schedulingscalersResource, c.ns, schedulingScaler), &v1alpha.SchedulingScaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.SchedulingScaler), err
}

// Update takes the representation of a schedulingScaler and updates it. Returns the server's representation of the schedulingScaler, and an error, if there is any.
func (c *FakeSchedulingScalers) Update(schedulingScaler *v1alpha.SchedulingScaler) (result *v1alpha.SchedulingScaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(schedulingscalersResource, c.ns, schedulingScaler), &v1alpha.SchedulingScaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.SchedulingScaler), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSchedulingScalers) UpdateStatus(schedulingScaler *v1alpha.SchedulingScaler) (*v1alpha.SchedulingScaler, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(schedulingscalersResource, "status", c.ns, schedulingScaler), &v1alpha.SchedulingScaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.SchedulingScaler), err
}

// Delete takes name of the schedulingScaler and deletes it. Returns an error if one occurs.
func (c *FakeSchedulingScalers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(schedulingscalersResource, c.ns, name), &v1alpha.SchedulingScaler{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSchedulingScalers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(schedulingscalersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha.SchedulingScalerList{})
	return err
}

// Patch applies the patch and returns the patched schedulingScaler.
func (c *FakeSchedulingScalers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha.SchedulingScaler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(schedulingscalersResource, c.ns, name, data, subresources...), &v1alpha.SchedulingScaler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.SchedulingScaler), err
}
