package test

import (
	"context"
	"errors"

	"github.com/stretchr/testify/mock"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientMock struct {
	mock.Mock
	client.Client
	objs map[string]client.Object
}

func key(obj client.Object) string {
	return obj.GetObjectKind().GroupVersionKind().Kind + "/" + obj.GetNamespace() + "/" + obj.GetName()
}

func (o *ClientMock) Get(ctx context.Context, nsname types.NamespacedName, obj client.Object) error {
	args := o.Called(ctx, nsname, obj)
	return args.Error(0)
}

func (o *ClientMock) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	k := key(obj)
	if _, exists := o.objs[k]; exists {
		return errors.New("already exists")
	}
	o.objs[k] = obj
	args := o.Called(ctx, obj, opts)
	return args.Error(0)
}

func (o *ClientMock) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	k := key(obj)
	if _, exists := o.objs[k]; !exists {
		return errors.New("doesn't exist")
	}
	o.objs[k] = obj
	args := o.Called(ctx, obj, opts)
	return args.Error(0)
}

func (o *ClientMock) MockExisting(obj client.Object) {
	if o.objs == nil {
		o.objs = map[string]client.Object{}
	}
	o.objs[key(obj)] = obj
	o.On("Get", mock.Anything, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(client.Object)
		arg.SetName(obj.GetName())
		arg.SetNamespace(obj.GetNamespace())
	}).Return(nil)
}

func (o *ClientMock) MockNonExisting(nsn types.NamespacedName) {
	o.On("Get", mock.Anything, nsn, mock.Anything).Return(kerr.NewNotFound(schema.GroupResource{}, ""))
}

func (o *ClientMock) MockCreateUpdate() {
	o.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	o.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
}
