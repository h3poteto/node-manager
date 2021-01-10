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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

// AWSNodeReplenisherReconciler reconciles a AWSNodeReplenisher object
type AWSNodeReplenisherReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodereplenishers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodereplenishers/status,verbs=get;update;patch

func (r *AWSNodeReplenisherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("awsnodereplenisher", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *AWSNodeReplenisherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AWSNodeReplenisher{}).
		Watches(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
