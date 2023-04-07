package watchers

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Watchable interface {
	ProvidePlaceholder() client.Object
	GetDigest(client.Object, []string) string
	Update(from, to client.Object)
}

type SecretWatchable struct {
	Watchable
}

func (w *SecretWatchable) ProvidePlaceholder() client.Object {
	return &corev1.Secret{}
}

func (w *SecretWatchable) GetDigest(obj client.Object, keys []string) string {
	secret := obj.(*corev1.Secret)
	// TODO: hash keys content
	return secret.Name
}

func (w *SecretWatchable) Update(from, to client.Object) {
	fromSecret := from.(*corev1.Secret)
	toSecret := to.(*corev1.Secret)
	toSecret.Data = fromSecret.Data
}

type ConfigWatchable struct {
	Watchable
}

func (w *ConfigWatchable) ProvidePlaceholder() client.Object {
	return &corev1.ConfigMap{}
}

func (w *ConfigWatchable) GetDigest(obj client.Object, keys []string) string {
	secret := obj.(*corev1.ConfigMap)
	// TODO: hash keys content
	return secret.Name
}

func (w *ConfigWatchable) Update(from, to client.Object) {
	fromSecret := from.(*corev1.ConfigMap)
	toSecret := to.(*corev1.ConfigMap)
	toSecret.Data = fromSecret.Data
}
