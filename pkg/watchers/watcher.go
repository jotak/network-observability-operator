package watchers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	rec "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

type Watcher struct {
	watched          map[string]interface{}
	defaultNamespace string
	secrets          SecretWatchable
	configs          ConfigWatchable
}

func NewWatcher() Watcher {
	return Watcher{
		watched: make(map[string]interface{}),
	}
}

func RegisterWatcher(builder *builder.Builder) *Watcher {
	w := NewWatcher()
	w.registerWatches(builder, &w.secrets, flowslatest.RefTypeSecret)
	w.registerWatches(builder, &w.configs, flowslatest.RefTypeConfigMap)
	return &w
}

func (w *Watcher) registerWatches(builder *builder.Builder, watchable Watchable, kind flowslatest.MountableType) {
	builder.Watches(
		&source.Kind{Type: watchable.ProvidePlaceholder()},
		handler.EnqueueRequestsFromMapFunc(func(o client.Object) []rec.Request {
			if w.isWatched(kind, o.GetName(), o.GetNamespace()) {
				// Trigger FlowCollector reconcile
				return []rec.Request{{NamespacedName: constants.FlowCollectorName}}
			}
			return []rec.Request{}
		}),
	)
}

func (w *Watcher) Reset(namespace string) {
	w.defaultNamespace = namespace
	w.watched = make(map[string]interface{})
}

func key(kind flowslatest.MountableType, name, namespace string) string {
	return string(kind) + "/" + namespace + "/" + name
}

func (w *Watcher) watch(kind flowslatest.MountableType, name, namespace string) {
	w.watched[key(kind, name, namespace)] = true
}

func (w *Watcher) isWatched(kind flowslatest.MountableType, name, namespace string) bool {
	if _, ok := w.watched[key(kind, name, namespace)]; ok {
		return true
	}
	return false
}

func (w *Watcher) ProcessMTLSCerts(ctx context.Context, cl helper.ClientHelper, tls *flowslatest.ClientTLS, targetNamespace string) (caDigest string, userDigest string, err error) {
	if tls.CACert.Name != "" {
		caRef := w.refFromCert(&tls.CACert)
		caDigest, err = w.reconcile(ctx, cl, caRef, targetNamespace)
		if err != nil {
			return "", "", err
		}
	}
	if tls.UserCert.Name != "" {
		userRef := w.refFromCert(&tls.UserCert)
		userDigest, err = w.reconcile(ctx, cl, userRef, targetNamespace)
		if err != nil {
			return "", "", err
		}
	}
	return caDigest, userDigest, nil
}

func (w *Watcher) Process(ctx context.Context, cl helper.ClientHelper, cos *flowslatest.ConfigOrSecret, targetNamespace string) (string, error) {
	return w.reconcile(ctx, cl, w.refFromConfigOrSecret(cos), targetNamespace)
}

func (w *Watcher) reconcile(ctx context.Context, cl helper.ClientHelper, ref objectRef, destNamespace string) (string, error) {
	rlog := log.FromContext(ctx, "Name", ref.name, "Source namespace", ref.namespace, "Target namespace", destNamespace)

	w.watch(ref.kind, ref.name, ref.namespace)
	var watchable Watchable
	if ref.kind == flowslatest.RefTypeConfigMap {
		watchable = &w.configs
	} else {
		watchable = &w.secrets
	}

	obj := watchable.ProvidePlaceholder()
	err := cl.Get(ctx, types.NamespacedName{Name: ref.name, Namespace: ref.namespace}, obj)
	if err != nil {
		return "", err
	}
	digest := watchable.GetDigest(obj, ref.keys)
	if ref.namespace != destNamespace {
		// copy to namespace
		target := watchable.ProvidePlaceholder()
		err := cl.Get(ctx, types.NamespacedName{Name: ref.name, Namespace: destNamespace}, target)
		if err != nil {
			if !errors.IsNotFound(err) {
				return "", err
			}
			rlog.Info(fmt.Sprintf("creating %s %s in namespace %s", ref.kind, ref.name, destNamespace))
			obj.SetName(ref.name)
			obj.SetNamespace(destNamespace)
			obj.SetAnnotations(map[string]string{
				constants.NamespaceCopyAnnotation: ref.namespace + "/" + ref.name,
			})
			obj.SetName(ref.name)
			if err := cl.CreateOwned(ctx, obj); err != nil {
				return "", err
			}
		} else {
			// Update existing
			rlog.Info(fmt.Sprintf("updating %s %s in namespace %s", ref.kind, ref.name, destNamespace))
			watchable.Update(obj, target)
			if err := cl.UpdateOwned(ctx, target, target); err != nil {
				return "", err
			}
		}
	}
	return digest, nil
}

func Annotation(key string) string {
	return constants.PodWatchedSuffix + key
}
