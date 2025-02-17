// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Clients holds the various API clients required by Contour.
type Clients struct {
	meta.RESTMapper

	core    *kubernetes.Clientset
	dynamic dynamic.Interface
}

// NewClients returns a new set of the various API clients required
// by Contour using the supplied kubeconfig path, or the cluster
// environment variables if inCluster is true.
func NewClients(kubeconfig string, inCluster bool) (*Clients, error) {
	config, err := newRestConfig(kubeconfig, inCluster)
	if err != nil {
		return nil, err
	}

	var clients Clients
	clients.core, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	clients.dynamic, err = dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	clients.RESTMapper, err = apiutil.NewDiscoveryRESTMapper(config)
	if err != nil {
		return nil, err
	}

	return &clients, nil
}

func newRestConfig(kubeconfig string, inCluster bool) (*rest.Config, error) {
	if kubeconfig != "" && !inCluster {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

// ClientSet returns the Kubernetes Core v1 ClientSet.
func (c *Clients) ClientSet() *kubernetes.Clientset {
	return c.core
}

// DynamicClient returns the dynamic client.
func (c *Clients) DynamicClient() dynamic.Interface {
	return c.dynamic
}

// ResourcesExist returns true if all of the GroupVersionResources
// passed exists in the cluster.
func (c *Clients) ResourcesExist(gvr ...schema.GroupVersionResource) bool {
	for _, r := range gvr {
		_, err := c.KindFor(r)
		if meta.IsNoMatchError(err) {
			return false
		}
	}

	return true
}
