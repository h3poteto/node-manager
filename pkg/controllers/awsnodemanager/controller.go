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

package awsnodemanager

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

// AWSNodeManagerReconciler reconciles a AWSNodeManager object
type AWSNodeManagerReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Session  *session.Session
}

// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodemanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodemanagers/status,verbs=get;update;patch

func (r *AWSNodeManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("awsnodemanager", req.NamespacedName)

	klog.Info("fetching AWSNodeManager resources")
	awsNodeManager := operatorv1alpha1.AWSNodeManager{}
	if err := r.Client.Get(ctx, req.NamespacedName, &awsNodeManager); err != nil {
		klog.Infof("failed to get AWSNodeManager resources: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Generate aws client
	r.Session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if err := r.syncAWSNodeManager(ctx, &awsNodeManager); err != nil {
		klog.Errorf("failed to sync AWSNodeManager: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AWSNodeManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AWSNodeManager{}).
		Owns(&operatorv1alpha1.AWSNodeReplenisher{}).
		Owns(&operatorv1alpha1.AWSNodeRefresher{}).
		Complete(r)
}

func (r *AWSNodeManagerReconciler) syncAWSNodeManager(ctx context.Context, awsNodeManager *operatorv1alpha1.AWSNodeManager) error {
	klog.Info("syncing nodes and aws instances")
	updated, err := r.syncAWSNodes(ctx, awsNodeManager)
	if err != nil {
		return err
	}
	if updated {
		return nil
	}
	refresher, err := r.syncAWSNodeRefresher(ctx, awsNodeManager)
	if err != nil {
		return err
	}
	replenisher, err := r.syncAWSNodeReplenisher(ctx, awsNodeManager)
	if err != nil {
		return err
	}
	if refresher != nil {
		switch refresher.Status.Phase {
		case operatorv1alpha1.AWSNodeRefresherUpdateIncreasing,
			operatorv1alpha1.AWSNodeRefresherUpdateReplacing,
			operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
			operatorv1alpha1.AWSNodeRefresherUpdateDecreasing:
			awsNodeManager.Status.Phase = operatorv1alpha1.AWSNodeManagerRefreshing
		default:
			awsNodeManager.Status.Phase = operatorv1alpha1.AWSNodeManagerSynced
		}
		awsNodeManager.Status.NodeRefresher = &operatorv1alpha1.AWSNodeRefresherRef{
			Namespace: refresher.Namespace,
			Name:      refresher.Name,
		}
	}
	if replenisher != nil {
		if replenisher.Status.Phase == operatorv1alpha1.AWSNodeReplenisherAWSUpdating {
			awsNodeManager.Status.Phase = operatorv1alpha1.AWSNodeManagerReplenishing
		} else {
			awsNodeManager.Status.Phase = operatorv1alpha1.AWSNodeManagerSynced
		}
		awsNodeManager.Status.NodeReplenisher = &operatorv1alpha1.AWSNodeReplenisherRef{
			Namespace: replenisher.Namespace,
			Name:      replenisher.Name,
		}
	}
	currentAWSNodeManager := operatorv1alpha1.AWSNodeManager{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: awsNodeManager.Namespace, Name: awsNodeManager.Name}, &currentAWSNodeManager); err != nil {
		klog.Errorf("failed to get AWSNodeManager %s/%s: %v", awsNodeManager.Namespace, awsNodeManager.Name, err)
		return err
	}
	if reflect.DeepEqual(awsNodeManager.Status, currentAWSNodeManager.Status) {
		klog.Infof("AWSNodeManager %s/%s is already synced", awsNodeManager.Namespace, awsNodeManager.Name)
		return nil
	}
	currentAWSNodeManager.Status = awsNodeManager.Status
	if err := r.Client.Update(ctx, &currentAWSNodeManager); err != nil {
		klog.Errorf("failed to update AWSNodemanager %s/%s: %v", currentAWSNodeManager.Namespace, currentAWSNodeManager.Name, err)
		return err
	}
	klog.Infof("updated AWSNodeManager status %s/%s", currentAWSNodeManager.Namespace, currentAWSNodeManager.Name)
	r.Recorder.Eventf(&currentAWSNodeManager, corev1.EventTypeNormal, "Updated", "Updated AWSNodeManager %s/%s", currentAWSNodeManager.Namespace, currentAWSNodeManager.Name)
	return nil
}
