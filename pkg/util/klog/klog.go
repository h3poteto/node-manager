package klog

import (
	"context"

	"k8s.io/klog/v2"

	pkgctx "github.com/h3poteto/node-manager/pkg/util/context"
)

func Info(ctx context.Context, args ...interface{}) {
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		klog.Info(args...)
		return
	}
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.Infof("[%s] %v", requestID, args)
		return
	}
	klog.Infof("[%s] {\"controller\": %q} %v", requestID, controller, args)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		klog.Infof(format, args...)
		return
	}
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.Infof("[%s] "+format, requestID, args)
		return
	}
	klog.Infof("[%s] {\"controller\": %q} "+format, requestID, controller, args)
}

func Warning(ctx context.Context, args ...interface{}) {
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		klog.Warning(args...)
		return
	}
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.Warningf("[%s] %v", requestID, args)
		return
	}
	klog.Warningf("[%s] {\"controller\": %q} %v", requestID, controller, args)
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		klog.Warningf(format, args...)
		return
	}
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.Warningf("[%s] "+format, requestID, args)
		return
	}
	klog.Warningf("[%s] {\"controller\": %q} "+format, requestID, controller, args)
}

func Error(ctx context.Context, args ...interface{}) {
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		klog.Error(args...)
		return
	}
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.Errorf("[%s] %v", requestID, args)
		return
	}
	klog.Errorf("[%s] {\"controller\": %q} %v", requestID, controller, args)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	rv := ctx.Value(pkgctx.RequestIDKey)
	requestID, ok := rv.(string)
	if !ok {
		klog.Errorf(format, args...)
		return
	}
	cv := ctx.Value(pkgctx.ControllerKey)
	controller, ok := cv.(string)
	if !ok {
		klog.Errorf("[%s] "+format, requestID, args)
		return
	}
	klog.Errorf("[%s] {\"controller\": %q} "+format, requestID, controller, args)
}
