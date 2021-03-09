package klog

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	pkgctx "github.com/h3poteto/node-manager/pkg/util/context"
)

func Info(ctx context.Context, args ...interface{}) {
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.InfoDepth(1, args...)
		return
	}
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		ctr := []interface{}{
			fmt.Sprintf("{\"controller\": %q} ", controller),
		}
		args = append(ctr, args...)
		klog.InfoDepth(1, args...)
		return
	}
	prefix := []interface{}{
		fmt.Sprintf("[%s] ", requestID),
		fmt.Sprintf("{\"controller\": %q} ", controller),
	}
	args = append(prefix, args...)
	klog.InfoDepth(1, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.InfoDepth(1, fmt.Sprintf(format, args...))
		return
	}
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		args = append([]interface{}{controller}, args...)
		klog.InfoDepth(1, fmt.Sprintf("{\"controller\": %q} "+format, args...))
		return
	}
	prefix := []interface{}{
		requestID,
		controller,
	}
	args = append(prefix, args...)
	klog.InfoDepth(1, fmt.Sprintf("[%s] {\"controller\": %q} "+format, args...))
}

func Warning(ctx context.Context, args ...interface{}) {
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.WarningDepth(1, args...)
		return
	}
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		ctr := []interface{}{
			fmt.Sprintf("{\"controller\": %q} ", controller),
		}
		args = append(ctr, args...)
		klog.WarningDepth(1, args...)
		return
	}
	prefix := []interface{}{
		fmt.Sprintf("[%s] ", requestID),
		fmt.Sprintf("{\"controller\": %q} ", controller),
	}
	args = append(prefix, args...)
	klog.WarningDepth(1, args...)
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.WarningDepth(1, fmt.Sprintf(format, args...))
		return
	}
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		args = append([]interface{}{controller}, args...)
		klog.WarningDepth(1, fmt.Sprintf("{\"controller\": %q} "+format, args...))
		return
	}
	prefix := []interface{}{
		requestID,
		controller,
	}
	args = append(prefix, args...)
	klog.WarningDepth(1, fmt.Sprintf("[%s] {\"controller\": %q} "+format, args...))
}

func Error(ctx context.Context, args ...interface{}) {
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.ErrorDepth(1, args...)
		return
	}
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		ctr := []interface{}{
			fmt.Sprintf("{\"controller\": %q} ", controller),
		}
		args = append(ctr, args...)
		klog.ErrorDepth(1, args...)
		return
	}
	prefix := []interface{}{
		fmt.Sprintf("[%s] ", requestID),
		fmt.Sprintf("{\"controller\": %q} ", controller),
	}
	args = append(prefix, args...)
	klog.ErrorDepth(1, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.ErrorDepth(1, fmt.Sprintf(format, args...))
		return
	}
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		args = append([]interface{}{controller}, args...)
		klog.ErrorDepth(1, fmt.Sprintf("{\"controller\": %q} "+format, args...))
		return
	}
	prefix := []interface{}{
		requestID,
		controller,
	}
	args = append(prefix, args...)
	klog.ErrorDepth(1, fmt.Sprintf("[%s] {\"controller\": %q} "+format, args...))
}
