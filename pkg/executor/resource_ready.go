package executor

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/discovery"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
)

var grCRD = schema.GroupResource{
	Group:    "apiextensions.k8s.io",
	Resource: "customresourcedefinitions",
}

func IsResourceExistsAndReady(ctx context.Context, dc dynamic.Interface, mapper discovery.ResourceMapper, gvr schema.GroupVersionResource) (bool, error) {
	exists, err := mapper.ExistsGVR(gvr)
	if gvr.GroupResource() != grCRD {
		return exists, err
	}

	crdGVR, err := mapper.Preferred(gvr)
	if err != nil {
		return false, err
	}
	obj, err := dc.Resource(crdGVR).Get(ctx, gvr.GroupResource().String(), metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	s, err := status.Compute(obj)
	if err != nil {
		return false, err
	}
	if s.Status != status.CurrentStatus {
		// How to surface this back to the Module CR?
		klog.Warningf("gvr %v status %+v", gvr, s)
	}
	return s.Status == status.CurrentStatus, nil
}
