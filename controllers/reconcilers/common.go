package reconcilers

import (
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/discover"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
)

type Common struct {
	helper.ClientHelper
	Watcher           *watchers.Watcher
	Namespace         string
	PreviousNamespace string
	UseOpenShiftSCC   bool
	AvailableAPIs     *discover.AvailableAPIs
}

func (c *Common) PrivilegedNamespace() string {
	return c.Namespace + constants.EBPFPrivilegedNSSuffix
}

func (c *Common) PreviousPrivilegedNamespace() string {
	return c.PreviousNamespace + constants.EBPFPrivilegedNSSuffix
}

type Instance struct {
	*Common
	Managed *NamespacedObjectManager
	Image   string
}

func (c *Common) NewInstance(image string) *Instance {
	managed := NewNamespacedObjectManager(c)
	return &Instance{
		Common:  c,
		Managed: managed,
		Image:   image,
	}
}
