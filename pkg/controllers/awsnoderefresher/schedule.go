package awsnoderefresher

import (
	"context"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"

	"github.com/gorhill/cronexpr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (r *AWSNodeRefresherReconciler) scheduleNext(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	next := metav1.NewTime(cronexpr.MustParse(refresher.Spec.Schedule).Next(time.Now()))
	refresher.Status.NextUpdateTime = &next
	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherScheduled
	refresher.Status.Revision += 1

	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf("failed to update AWSNodeRefresher %s/%s: %v", refresher.Namespace, refresher.Name, err)
		return err
	}
	return nil
}
