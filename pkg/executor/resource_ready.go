package executor

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/discovery"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ridCRD = kmapi.ResourceID{
	Group:   "apiextensions.k8s.io",
	Name:    "customresourcedefinitions",
	Version: "v1",
	Kind:    "CustomResourceDefinition",
}

func IsResourceExistsAndReady(ctx context.Context, dc client.Client, mapper discovery.ResourceMapper, gvr schema.GroupVersionResource) (bool, error) {
	exists, err := mapper.ExistsGVR(gvr)
	if gvr.GroupResource() != ridCRD.GroupResource() {
		return exists, err // not a CRD
	}

	var obj unstructured.Unstructured
	obj.SetGroupVersionKind(ridCRD.GroupVersionKind())
	err = dc.Get(ctx, client.ObjectKey{Name: gvr.GroupResource().String()}, &obj)
	if err != nil {
		return false, err
	}
	s, err := status.Compute(&obj)
	if err != nil {
		return false, err
	}
	if s.Status != status.CurrentStatus {
		// How to surface this back to the Module CR?
		klog.Warningf("gvr %v status %+v", gvr, s)
	}
	return s.Status == status.CurrentStatus, nil
}
