/*
Copyright 2018 Pressinfra SRL.

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

package wordpresspolicy

import (
	"context"
	"fmt"
	"log"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new WordpressPolicy Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this wordpress.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWordpressPolicy{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("wordpress-policy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to WordpressPolicy
	err = c.Watch(&source.Kind{Type: &wordpressv1alpha1.WordpressPolicy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileWordpressPolicy{}

// ReconcileWordpressPolicy reconciles a WordpressPolicy object
type ReconcileWordpressPolicy struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a WordpressPolicy object and makes changes based on the state read
// and what is in the WordpressPolicy.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wordpress.presslabs.org,resources=wordpresspolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wordpress.presslabs.org,resources=wordpress,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileWordpressPolicy) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the WordpressPolicy instance
	policy := &wordpressv1alpha1.WordpressPolicy{}
	err := r.Get(context.TODO(), request.NamespacedName, policy)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// TODO(calind): Use finalizers for removing policies from Wordpress
			// objects
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	wps := &wordpressv1alpha1.WordpressList{}
	opts := &client.ListOptions{}

	if policy.Spec.Selector != nil {
		selector, err := metav1.LabelSelectorAsSelector(policy.Spec.Selector)
		if err != nil {
			return reconcile.Result{}, err
		}
		opts.LabelSelector = selector
	}

	err = r.List(context.TODO(), opts, wps)
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, wp := range wps.Items {
		if len(wp.Annotations) == 0 {
			wp.Annotations = make(map[string]string)
		}
		ann := fmt.Sprintf("%s%s", wordpressv1alpha1.WordpressPolicyAnnotationPrefix, policy.Name)
		log.Printf("Setting %s/%s policy %s version to %s", wp.Namespace, wp.Name, policy.Name, policy.ResourceVersion)
		wp.Annotations[ann] = policy.ResourceVersion

		err = r.Update(context.TODO(), &wp)
		if err != nil {
			return reconcile.Result{}, nil
		}
	}

	return reconcile.Result{}, nil
}
