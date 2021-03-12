package awsnoderefresher

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

type externalEventWatcher struct {
	channel chan event.GenericEvent
	client  client.Client
}

func newExternalEventWatcher() *externalEventWatcher {
	ch := make(chan event.GenericEvent)

	return &externalEventWatcher{
		channel: ch,
	}
}

func (e *externalEventWatcher) InjectClient(c client.Client) error {
	e.client = c
	return nil
}

func (e *externalEventWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			var refreshers operatorv1alpha1.AWSNodeRefresherList
			err := e.client.List(ctx, &refreshers)
			if err != nil {
				break
			}
			for _, ref := range refreshers.Items {
				e.channel <- event.GenericEvent{
					Object: &ref,
				}
			}
		}
	}
}

func contextFromStopChannel(ch <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		<-ch
	}()
	return ctx
}
