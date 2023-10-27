package cachefilter

import (
	"sync"

	toolscache "k8s.io/client-go/tools/cache"
	kcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PerObjectCache struct {
	Options kcache.Options
	wmut    sync.Mutex
	watches map[string]client.Object
}

func NewPerObject(gvks ...client.Object) *PerObjectCache {
	npo := PerObjectCache{
		watches: make(map[string]client.Object),
	}
	npo.Options = kcache.Options{
		ByObject: map[client.Object]kcache.ByObject{},
	}
	for _, obj := range gvks {
		npo.Options.ByObject[obj] = kcache.ByObject{Transform: npo.filter(obj)}
	}
	return &npo
}

func (c *PerObjectCache) Get(kind, name, namespace string) client.Object {
	k := key(kind, name, namespace)
	c.wmut.Lock()
	obj, ok := c.watches[k]
	if !ok {
		c.watches[k] = nil
	}
	c.wmut.Unlock()
	return obj
}

func (c *PerObjectCache) GetObject(obj client.Object) client.Object {
	return c.Get(
		obj.GetObjectKind().GroupVersionKind().Kind,
		obj.GetName(),
		obj.GetNamespace(),
	)
}

func key(kind, name, namespace string) string {
	return kind + "/" + namespace + "/" + name
}

func (c *PerObjectCache) filter(placeholder client.Object) toolscache.TransformFunc {
	return func(p any) (any, error) {
		obj, ok := p.(client.Object)
		if !ok {
			return p, nil
		}

		k := key(
			obj.GetObjectKind().GroupVersionKind().Kind,
			obj.GetName(),
			obj.GetNamespace(),
		)
		c.wmut.Lock()
		_, isWatched := c.watches[k]
		if isWatched {
			c.watches[k] = obj
		}
		c.wmut.Unlock()

		// Set a lighter version in cache
		lightObj := placeholder.DeepCopyObject().(client.Object)
		lightObj.SetName(obj.GetName())
		lightObj.SetNamespace(obj.GetNamespace())
		lightObj.SetUID(obj.GetUID())
		lightObj.SetOwnerReferences(obj.GetOwnerReferences())
		lightObj.SetGeneration(obj.GetGeneration())
		lightObj.SetResourceVersion(obj.GetResourceVersion())
		return lightObj, nil
	}
}
