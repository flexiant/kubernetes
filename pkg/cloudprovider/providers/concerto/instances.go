/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package concerto_cloud

import (
	"net"
	"regexp"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
)

// NodeAddresses returns the addresses of the specified instance.
func (concerto *ConcertoCloud) NodeAddresses(name string) ([]api.NodeAddress, error) {
	glog.Infoln("Concerto NodeAddresses", name)
	ci, err := concerto.service.GetInstanceByName(name)
	if err != nil {
		return nil, err
	}
	publicAddress := api.NodeAddress{
		Type: api.NodeExternalIP, Address: net.ParseIP(ci.PublicIP).String()}
	return []api.NodeAddress{publicAddress}, nil
}

// ExternalID returns the cloud provider ID of the specified instance (deprecated).
func (concerto *ConcertoCloud) ExternalID(name string) (string, error) {
	glog.Infoln("Concerto ExternalID", name)
	return concerto.InstanceID(name)
}

// InstanceID returns the cloud provider ID of the specified instance.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (concerto *ConcertoCloud) InstanceID(name string) (string, error) {
	glog.Infoln("Concerto InstanceID", name)
	ci, err := concerto.service.GetInstanceByName(name)
	if err != nil {
		return "", err
	}
	return ci.Id, nil
}

// List lists instances that match 'filter' which is a regular expression which must match the entire instance name (fqdn)
func (concerto *ConcertoCloud) List(filter string) ([]string, error) {
	glog.Infoln("Concerto List", filter)
	regexp, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}
	instances, err := concerto.service.GetInstanceList()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, instance := range instances {
		if regexp.MatchString(instance.Name) {
			names = append(names, instance.Name)
		}
	}
	return names, nil
}

// GetNodeResources gets the resources for a particular node
func (concerto *ConcertoCloud) GetNodeResources(name string) (*api.NodeResources, error) {
	glog.Infoln("Concerto GetNodeResources", name)
	ci, err := concerto.service.GetInstanceByName(name)
	if err != nil {
		return nil, err
	}
	return makeNodeResources(ci.CPUs, ci.Memory), nil
}

// Returns the name of the node we are currently running on
func (concerto *ConcertoCloud) CurrentNodeName(hostname string) (string, error) {
	glog.Infoln("Concerto CurrentNodeName", hostname)
	return hostname, nil
}

// NOT SUPPORTED in Concerto Cloud
func (concerto *ConcertoCloud) AddSSHKeyToAllInstances(user string, keyData []byte) error {
	glog.Infoln("Concerto AddSSHKeyToAllInstances!")
	return UnsupportedOperation
}

// Builds an api.NodeResources
// cpu is in cores, memory is in MiB
func makeNodeResources(cpu float64, memory int64) *api.NodeResources {
	return &api.NodeResources{
		Capacity: api.ResourceList{
			api.ResourceCPU:    *resource.NewMilliQuantity(int64(cpu*1000), resource.DecimalSI),
			api.ResourceMemory: *resource.NewQuantity(int64(memory*1024*1024), resource.BinarySI),
		},
	}
}
