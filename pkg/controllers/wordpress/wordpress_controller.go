/*
Copyright 2019 Pressinfra SRL.

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

package wordpress

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/kubernetes/client-go/kubernetes/scheme"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/client-go/tools/record"

	"github.com/presslabs/controller-util/syncer"
	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/internal/wordpress"
)

// WordpressReconciler reconciles a Wordpress object
type WordpressReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	scheme   *scheme.Scheme
}

// +kubebuilder:rbac:groups=wordpress.presslabs.org,resources=wordpresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wordpress.presslabs.org,resources=wordpresses/status,verbs=get;update;patch

func (r *WordpressReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("wordpress", req.NamespacedName)

	// Fetch the Wordpress instance
	wp := wordpress.New(&wordpressv1alpha1.Wordpress{})
	err := r.Get(context.TODO(), req.NamespacedName, wp.Unwrap())
	if ignoreNotFound(err) != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	r.scheme.Default(wp.Unwrap())
	wp.SetDefaults()

	secretSyncer := r.newSecretSyncer(wp)
	deploySyncer := r.newDeploymentSyncer(wp, secretSyncer.GetObject().(*corev1.Secret))
	syncers := []syncer.Interface{
		secretSyncer,
		deploySyncer,
		r.newServiceSyncer(wp),
		r.newIngressSyncer(wp),
		r.newWPCronSyncer(wp),
		// r.newDBUpgradeJobSyncer(wp),
	}

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, r.newCodePVCSyncer(wp))
	}

	if wp.Spec.MediaVolumeSpec != nil && wp.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, r.newMediaPVCSyncer(wp))
	}

	if err = r.sync(syncers); err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WordpressReconciler) sync(syncers []syncer.Interface) error {
	for _, s := range syncers {
		if err := syncer.Sync(context.TODO(), s, r.Recorder); err != nil {
			return err
		}
	}
	return nil
}

func (r *WordpressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("wordpress-controller")
	r.scheme = mgr.GetScheme()
	return ctrl.NewControllerManagedBy(mgr).
		For(&wordpressv1alpha1.Wordpress{}).
		Complete(r)
}
