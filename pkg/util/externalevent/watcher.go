package externalevent

import (
	"context"
	"time"

	"github.com/h3poteto/node-manager/pkg/util/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type externalEventWatcher struct {
	Channel  chan event.GenericEvent
	client   client.Client
	interval time.Duration
	fn       ListFunc
}

type ListFunc = func(ctx context.Context, c client.Client) ([]client.Object, error)

func NewExternalEventWatcher(interval time.Duration, listFn ListFunc) *externalEventWatcher {
	ch := make(chan event.GenericEvent)

	return &externalEventWatcher{
		Channel:  ch,
		interval: interval,
		fn:       listFn,
	}
}

func (e *externalEventWatcher) InjectClient(c client.Client) error {
	e.client = c
	return nil
}

func (e *externalEventWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			klog.Info(ctx, "Force syncing")
			list, err := e.fn(ctx, e.client)
			if err != nil {
				break
			}
			for _, ref := range list {
				e.Channel <- event.GenericEvent{
					Object: ref,
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
