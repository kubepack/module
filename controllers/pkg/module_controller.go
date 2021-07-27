/*
Copyright 2021.

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

package pkg

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/discovery"
	clientcmdutil "kmodules.xyz/client-go/tools/clientcmd"
	"log"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	pkg "kubepack.dev/module/apis/pkg/v1alpha1"
)

// ModuleReconciler reconciles a Module object
type ModuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	mgr  ctrl.Manager
	ctrl controller.Controller

	// Module -> Matchers
	ModuleToMatchers map[types.NamespacedName]map[schema.GroupVersionKind][]Matcher

	// Kind -> Matchers -> []Module
	KindToModule map[schema.GroupVersionKind]map[Matcher][]types.NamespacedName
}

type Matcher struct {
	Name      string
	Namespace string
	Selector  *metav1.LabelSelector
}

//+kubebuilder:rbac:groups=pkg.kubepack.com,resources=modules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pkg.kubepack.com,resources=modules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pkg.kubepack.com,resources=modules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Module object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ModuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// your logic here
	var module pkg.Module
	if err := r.Get(ctx, req.NamespacedName, &module); err != nil {
		log.Error(err, "unable to fetch Module")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// FIX(tamal): Observed generation checking is not enough, since the overridden values can change.

	for _, action := range module.Spec.Actions {
		runner := ActionRunner{
			dc:           dc,
			ClientGetter: getter,
			mapper:       discovery.NewResourceMapper(mapper),
			ModuleName:   myflow.Name,
			Namespace:    myflow.Namespace,
			action:       action,
		}
		err := runner.Execute()
		if err != nil {
			klog.Fatalln(err)
		}
	}

	return ctrl.Result{}, nil
}

var (
	masterURL      = ""
	kubeconfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
)

// SetupWithManager sets up the controller with the Manager.
func (r *ModuleReconciler) SetupWithManager(mgr ctrl.Manager) (err error) {
	//mgr.

	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}})
	kubeconfig, err := cc.RawConfig()
	if err != nil {
		klog.Fatal(err)
	}
	getter := clientcmdutil.NewClientGetter(&kubeconfig)

	config, err := cc.ClientConfig() // clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	dc := dynamic.NewForConfigOrDie(config)
	mapper, err := getter.ToRESTMapper()
	if err != nil {
		klog.Fatal(err)
	}

	r.ctrl, err = ctrl.NewControllerManagedBy(mgr).
		For(&pkg.Module{}).
		Build(r)
	return err
}
