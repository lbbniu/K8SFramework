/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	"fmt"

	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
	crdv1beta1 "k8s.tars.io/client-go/clientset/versioned/typed/crd/v1beta1"
	crdv1beta2 "k8s.tars.io/client-go/clientset/versioned/typed/crd/v1beta2"
	crdv1beta3 "k8s.tars.io/client-go/clientset/versioned/typed/crd/v1beta3"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	CrdV1beta1() crdv1beta1.CrdV1beta1Interface
	CrdV1beta2() crdv1beta2.CrdV1beta2Interface
	CrdV1beta3() crdv1beta3.CrdV1beta3Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	crdV1beta1 *crdv1beta1.CrdV1beta1Client
	crdV1beta2 *crdv1beta2.CrdV1beta2Client
	crdV1beta3 *crdv1beta3.CrdV1beta3Client
}

// CrdV1beta1 retrieves the CrdV1beta1Client
func (c *Clientset) CrdV1beta1() crdv1beta1.CrdV1beta1Interface {
	return c.crdV1beta1
}

// CrdV1beta2 retrieves the CrdV1beta2Client
func (c *Clientset) CrdV1beta2() crdv1beta2.CrdV1beta2Interface {
	return c.crdV1beta2
}

// CrdV1beta3 retrieves the CrdV1beta3Client
func (c *Clientset) CrdV1beta3() crdv1beta3.CrdV1beta3Interface {
	return c.crdV1beta3
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.crdV1beta1, err = crdv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.crdV1beta2, err = crdv1beta2.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.crdV1beta3, err = crdv1beta3.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.crdV1beta1 = crdv1beta1.NewForConfigOrDie(c)
	cs.crdV1beta2 = crdv1beta2.NewForConfigOrDie(c)
	cs.crdV1beta3 = crdv1beta3.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.crdV1beta1 = crdv1beta1.New(c)
	cs.crdV1beta2 = crdv1beta2.New(c)
	cs.crdV1beta3 = crdv1beta3.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
