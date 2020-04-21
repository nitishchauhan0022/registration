// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/open-cluster-management/api/client/cluster/clientset/versioned/typed/cluster/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeClusterV1 struct {
	*testing.Fake
}

func (c *FakeClusterV1) SpokeClusters() v1.SpokeClusterInterface {
	return &FakeSpokeClusters{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeClusterV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
