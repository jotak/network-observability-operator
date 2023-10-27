package cachefilter

import (
	"context"
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	client.Client
	cache *PerObjectCache
}

func NewClient(cl client.Client, cache *PerObjectCache) *Client {
	return &Client{
		Client: cl,
		cache:  cache,
	}
}

func (c *Client) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	kind := reflect.TypeOf(obj).Elem().Name()
	cached := c.cache.Get(kind, key.Name, key.Namespace)
	if cached != nil {
		// Use our cache in priority

		// Copy the value of the item in the cache to the returned value
		// TODO(directxman12): this is a terrible hack, pls fix (we should have deepcopyinto)
		objVal := reflect.ValueOf(obj)
		cachedVal := reflect.ValueOf(cached)
		if !cachedVal.Type().AssignableTo(objVal.Type()) {
			return fmt.Errorf("cache had type %s, but %s was asked for", cachedVal.Type(), objVal.Type())
		}
		reflect.Indirect(objVal).Set(reflect.Indirect(cachedVal))
		return nil
	} else {
		// TODO: 2 scenarios here: either we're on restricted cache content (CM, Secret) in which case we need to run a live client call (no cache)
		// Or we using plain cache and we call c.Client.Get
		return c.Client.Get(ctx, key, obj, opts...)
	}
	return o
}
