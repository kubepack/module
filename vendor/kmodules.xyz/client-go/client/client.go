package client

import (
	"context"
	"strings"

	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOrPatch(ctx context.Context, c client.Client, obj client.Object, transform func(client.Object) client.Object, opts metav1.PatchOptions) (client.Object, kutil.VerbType, error) {
	key := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
	err := c.Get(ctx, key, obj)
	if kerr.IsNotFound(err) {
		klog.V(3).Infof("Creating %+v %s/%s.", obj.GetObjectKind().GroupVersionKind(), key.Namespace, key.Name)

		createOpts := []client.CreateOption{
			client.FieldOwner(opts.FieldManager),
		}
		if len(opts.DryRun) > 0 {
			createOpts = append(createOpts, client.DryRunAll)
		}
		obj = transform(obj.DeepCopyObject().(client.Object))
		err := c.Create(ctx, obj, createOpts...)
		return obj, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patchOpts := []client.PatchOption{
		client.FieldOwner(opts.FieldManager),
	}
	if len(opts.DryRun) > 0 {
		patchOpts = append(patchOpts, client.DryRunAll)
	}
	if opts.Force != nil && *opts.Force {
		patchOpts = append(patchOpts, client.ForceOwnership)
	}

	var patch client.Patch
	if isOfficialTypes(obj.GetObjectKind().GroupVersionKind().Group) {
		patch = client.StrategicMergeFrom(obj)
	} else {
		patch = client.MergeFrom(obj)
	}

	obj = transform(obj.DeepCopyObject().(client.Object))
	err = c.Patch(ctx, obj, patch, patchOpts...)
	return obj, kutil.VerbPatched, err
}

func isOfficialTypes(group string) bool {
	return !strings.ContainsRune(group, '.')
}
