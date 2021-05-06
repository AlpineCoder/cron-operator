/*


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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	planesv1beta1 "github.com/AlpineCoder/cron-operator/api/v1beta1"
)

// TransportReconciler reconciles a Transport object
type TransportReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=planes.chickensushi.ch,resources=transports,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=planes.chickensushi.ch,resources=transports/status,verbs=get;update;patch

func (r *TransportReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("transport", req.NamespacedName)

	// your logic here

	transport := &planesv1beta1.Transport{}

	err := r.Get(ctx, req.NamespacedName, transport)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not existingCronJob, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			klog.Infof("transport %s/%s not found. Ignoring since object must be deleted\n", req.Namespace, req.Name)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		klog.Error(err, "Failed to get transport: %v")
		return ctrl.Result{}, err
	}

	if err := r.generateCronJob(transport); err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TransportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&planesv1beta1.Transport{}).
		Complete(r)
}
