package externalevent

import (
	"context"
	"time"

	"github.com/h3poteto/node-manager/pkg/util/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type externalEventWatcher[T client.Object] struct {
	Channel  chan event.TypedGenericEvent[T]
	client   client.Client
	interval time.Duration
	fn       ListFunc[T]
}

type ListFunc[T client.Object] func(ctx context.Context, c client.Client) ([]T, error)

func NewExternalEventWatcher[T client.Object](interval time.Duration, listFn ListFunc[T]) *externalEventWatcher[T] {
	ch := make(chan event.TypedGenericEvent[T])

	return &externalEventWatcher[T]{
		Channel:  ch,
		interval: interval,
		fn:       listFn,
	}
}

func (e *externalEventWatcher[T]) InjectClient(c client.Client) error {
	e.client = c
	return nil
}

func (e *externalEventWatcher[T]) Start(ctx context.Context) error {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			list, err := e.fn(ctx, e.client)
			if err != nil {
				break
			}
			for _, ref := range list {
				klog.Infof(ctx, "Force syncing %s/%s", ref.GetNamespace(), ref.GetName())
				e.Channel <- event.TypedGenericEvent[T]{
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
