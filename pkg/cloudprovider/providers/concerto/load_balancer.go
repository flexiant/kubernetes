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
	"fmt"
	"net"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
)

// GetTCPLoadBalancer implementation for Flexiant Concerto.
func (c *ConcertoCloud) GetTCPLoadBalancer(name, _region string) (status *api.LoadBalancerStatus, exists bool, err error) {
	glog.Infoln("Concerto GetTCPLoadBalancer", name)

	lb, err := c.service.GetLoadBalancerByName(name)
	if err != nil {
		glog.Error("Error in GetTCPLoadBalancer: ", err)
		return nil, false, err
	}

	if lb == nil {
		return nil, false, nil
	}

	status = toStatus(lb)
	return status, true, nil
}

func toStatus(lb *ConcertoLoadBalancer) *api.LoadBalancerStatus {
	status := &api.LoadBalancerStatus{}

	var ingress api.LoadBalancerIngress
	ingress.Hostname = lb.FQDN
	status.Ingress = []api.LoadBalancerIngress{ingress}

	return status
}

// EnsureTCPLoadBalancer implementation for Flexiant Concerto.
func (c *ConcertoCloud) EnsureTCPLoadBalancer(name, region string, loadBalancerIP net.IP, ports []*api.ServicePort, hosts []string, affinityType api.ServiceAffinity) (*api.LoadBalancerStatus, error) {
	glog.Infof("Concerto EnsureTCPLoadBalancer %s %v", name, hosts)
	for i, p := range ports {
		glog.Infof("Concerto EnsureTCPLoadBalancer port: %v %#v", i, p)
	}

	// Concerto LB does not support session affinity
	if affinityType != api.ServiceAffinityNone {
		return nil, LoadBalancerUnsupportedAffinityError
	}
	// Can not specify a public IP for the LB
	if loadBalancerIP != nil {
		return nil, LoadBalancerUnsupportedExternalIPError
	}
	// Dont support multi-port
	if len(ports) != 1 {
		return nil, LoadBalancerUnsupportedNumberOfPortsError
	}

	// Check previous existence
	lb, err := c.service.GetLoadBalancerByName(name)
	if err != nil {
		glog.Error("Error in EnsureTCPLoadBalancer: ", err)
		return nil, err
	}

	if lb == nil {
		// It does not exist: create it
		lb, err = c.createTCPLoadBalancer(name, ports, hosts)
		if err != nil {
			glog.Error("Error in EnsureTCPLoadBalancer: ", err)
			return nil, err
		}
	} else {
		// It already exists: update it
		err = c.UpdateTCPLoadBalancer(name, region, hosts)
		if err != nil {
			glog.Error("Error in EnsureTCPLoadBalancer: ", err)
			return nil, err
		}
	}

	return toStatus(lb), nil
}

func (c *ConcertoCloud) createTCPLoadBalancer(name string, ports []*api.ServicePort, hosts []string) (*ConcertoLoadBalancer, error) {
	// Create the LB
	port := ports[0].Port // The port that will be exposed on the service.
	// targetPort := ports[0].TargetPort // Optional: The target port on pods selected by this service
	nodePort := ports[0].NodePort // The port on each node on which this service is exposed.
	lb, err := c.service.CreateLoadBalancer(name, port, nodePort)
	if err != nil {
		glog.Error("Error in EnsureTCPLoadBalancer: ", err)
		return nil, err
	}

	// Add the corresponding nodes
	if len(hosts) > 0 {
		ipAddresses, err := c.hostsNamesToIPs(hosts)
		if err != nil {
			glog.Error("Error in EnsureTCPLoadBalancer: ", err)
			return nil, err
		}
		err = c.service.RegisterInstancesWithLoadBalancer(lb.Id, ipAddresses)
		if err != nil {
			glog.Error("Error in EnsureTCPLoadBalancer: ", err)
			return nil, err
		}
	}

	return lb, nil
}

// UpdateTCPLoadBalancer implementation for Flexiant Concerto.
func (c *ConcertoCloud) UpdateTCPLoadBalancer(name, region string, hosts []string) error {
	glog.Infoln("Concerto UpdateTCPLoadBalancer", name, hosts)

	// Get the load balancer
	lb, err := c.service.GetLoadBalancerByName(name)
	if err != nil {
		glog.Error("Error in UpdateTCPLoadBalancer: ", err)
		return err
	}
	// Get the LB nodes
	currentNodes, err := c.service.GetLoadBalancerNodesAsIPs(lb.Id)
	if err != nil {
		glog.Error("Error in UpdateTCPLoadBalancer: ", err)
		return err
	}

	// Calculate nodes to deregister
	wantedNodes, err := c.hostsNamesToIPs(hosts)
	if err != nil {
		glog.Error("Error in UpdateTCPLoadBalancer: ", err)
		return err
	}
	nodesToRemove := subtractStringArrays(currentNodes, wantedNodes)
	// Calculate nodes to be registered
	nodesToAdd := subtractStringArrays(wantedNodes, currentNodes)
	// Lets do it
	glog.Infof("UpdateTCPLoadBalancer will remove %v for %s", nodesToRemove, name)
	glog.Infof("UpdateTCPLoadBalancer will add %v for %s", nodesToAdd, name)
	err = c.service.DeregisterInstancesFromLoadBalancer(lb.Id, nodesToRemove)
	if err != nil {
		glog.Error("Error in UpdateTCPLoadBalancer: ", err)
		return err
	}
	err = c.service.RegisterInstancesWithLoadBalancer(lb.Id, nodesToAdd)
	if err != nil {
		glog.Error("Error in UpdateTCPLoadBalancer: ", err)
		return err
	}
	// Done!
	glog.Infof("UpdateTCPLoadBalancer success: %s", name)
	return nil
}

// EnsureTCPLoadBalancerDeleted implementation for Flexiant Concerto.
func (c *ConcertoCloud) EnsureTCPLoadBalancerDeleted(name, region string) error {
	glog.Infoln("Concerto EnsureTCPLoadBalancerDeleted", name)

	// Get the LB
	lb, err := c.service.GetLoadBalancerByName(name)
	if err != nil {
		glog.Error("Error in EnsureTCPLoadBalancerDeleted: ", err)
		return err
	}
	if lb == nil {
		return nil
	}
	return c.service.DeleteLoadBalancerById(lb.Id)
}

func (c *ConcertoCloud) hostsNamesToIPs(hosts []string) ([]string, error) {
	var ips []string
	glog.Infoln("Looking up following hosts", hosts)
	instances, err := c.service.GetInstanceList()
	if err != nil {
		return nil, fmt.Errorf("Error while converting names to IP addresses: %v", err)
	}
	for _, name := range hosts {
		found := false
		for _, instance := range instances {
			if instance.Name == name {
				ips = append(ips, instance.PublicIP)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("Could not find instance: %s", name)
		}
	}
	return ips, nil
}
