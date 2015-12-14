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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

// ConcertoAPIService is an abstraction for Flexiant Concerto API.
type ConcertoAPIService interface {
	// Retrieves the info related to the instance which name is passed
	GetInstanceByName(name string) (ConcertoInstance, error)
	// Retrieves all instances
	GetInstanceList() ([]ConcertoInstance, error)
	// Creates a LB with the specified name
	CreateLoadBalancer(name string, port int, nodePort int) (*ConcertoLoadBalancer, error)
	// Retrieves a LB with the specified name
	GetLoadBalancerByName(name string) (*ConcertoLoadBalancer, error)
	// Deletes Load Balancer with given Id
	DeleteLoadBalancerById(id string) error
	// Gets the nodes registered with the load balancer
	GetLoadBalancerNodes(loadBalancerId string) ([]ConcertoLoadBalancerNode, error)
	// Gets the IPs of the nodes registered with the load balancer
	GetLoadBalancerNodesAsIPs(loadBalancerId string) ([]string, error)
	// Registers the instances with the load balancer
	RegisterInstancesWithLoadBalancer(loadBalancerId string, nodesIPs []string) error
	// Deregisters the instances from the load balancer
	DeregisterInstancesFromLoadBalancer(loadBalancerId string, nodesIPs []string) error
}

// ConcertoInstance is an abstraction for a Concerto cloud instance
type ConcertoInstance struct {
	Id       string  // Unique identifier for the instance in Concerto
	Name     string  // Hostname for the instance
	PublicIP string  // Public IP for the instance
	CPUs     float64 // Number of cores
	Memory   int64   // Amount of RAM (in MiB)
	Storage  int64   // Amount of disk (in GiB)
}

// Ship is used for deserializing
type Ship struct {
	Id             string
	Fqdn           string
	Name           string
	Public_ip      string
	Server_plan_id string
	Cpus           float64 // Number of cores
	Memory         int64   // Amount of RAM (in MB)
	Storage        int64   // Amount of disk (in GiB)
}

// ConcertoLoadBalancer abstracts a Concerto Load Balancer
type ConcertoLoadBalancer struct {
	Id       string `json:"id"`   // Unique identifier for the LB in Concerto
	Name     string `json:"name"` // Name of the LB in concerto
	FQDN     string `json:"fqdn"` // Fully Qualified domain name
	Port     int    `json:"port"`
	NodePort int    `json:"nodeport"`
	Protocol string `json:"protocol"`
}

// ConcertoLoadBalancer abstracts a Concerto Load Balancer
type ConcertoLoadBalancerNode struct {
	ID string
	IP string `json:"public_ip"`
	// Port int    `json:"port"`
}

// Concerto REST API client implementation
type concertoAPIServiceREST struct {
	// Pre-configured HTTP client
	client *restService
}

// BuildConcertoRESTClient Factory for 'concertoAPIServiceREST' objects
func buildConcertoRESTClient(config ConcertoConfig) (ConcertoAPIService, error) {
	glog.Infoln("buildConcertoRESTClient")
	rs, err := newRestService(config)
	if err != nil {
		glog.Error("Error in buildConcertoRESTClient: ", err)
		return nil, err
	}
	glog.Infoln("buildConcertoRESTClient succeeded")
	return &concertoAPIServiceREST{client: rs}, nil
}

func (c *concertoAPIServiceREST) GetInstanceList() ([]ConcertoInstance, error) {
	glog.Infoln("GetInstanceList")

	var ships []Ship
	var instances []ConcertoInstance

	data, status, err := c.client.Get("/kaas/ships")
	if err != nil {
		glog.Error("Error in GetInstanceList: ", err)
		return nil, err
	}

	if status == 404 {
		return instances, nil
	}

	err = json.Unmarshal(data, &ships)
	if err != nil {
		glog.Error("Error in GetInstanceList: ", err)
		return nil, err
	}

	for _, s := range ships {
		concertoInstance := ConcertoInstance{
			Id:       s.Id,
			Name:     s.Fqdn,
			PublicIP: s.Public_ip,
			CPUs:     s.Cpus,
			Memory:   s.Memory,
			Storage:  s.Storage,
		}
		instances = append(instances, concertoInstance)
	}

	glog.Infof("GetInstanceList got %#v", instances)

	return instances, nil
}

func (c *concertoAPIServiceREST) GetInstanceByName(name string) (ConcertoInstance, error) {
	glog.Infoln("GetInstanceByName", name)

	concertoInstances, err := c.GetInstanceList()
	if err != nil {
		glog.Error("Error in GetInstanceByName: ", err)
		return ConcertoInstance{}, err
	}

	for _, instance := range concertoInstances {
		if instance.Name == name {
			glog.Infof("GetInstanceByName got %#v", instance)
			return instance, nil
		}
	}

	glog.Infof("GetInstanceByName did not find %#v", name)
	return ConcertoInstance{}, cloudprovider.InstanceNotFound
}

func (c *concertoAPIServiceREST) GetLoadBalancerList() ([]ConcertoLoadBalancer, error) {
	glog.Infoln("GetLoadBalancerList")

	var lbs []ConcertoLoadBalancer

	data, status, err := c.client.Get("/kaas/load_balancers")
	if err != nil {
		glog.Error("Error in GetLoadBalancerList: ", err)
		return nil, err
	}

	if status >= 400 {
		return nil, fmt.Errorf("HTTP %v when getting '/kaas/load_balancers'", status)
	}

	err = json.Unmarshal(data, &lbs)
	if err != nil {
		glog.Error("Error in GetLoadBalancerList: ", err)
		glog.Error("Received data: ", string(data))
		return nil, err
	}

	glog.Infof("GetLoadBalancerList got %#v", lbs)

	return lbs, nil
}

func (c *concertoAPIServiceREST) GetLoadBalancerByName(name string) (*ConcertoLoadBalancer, error) {
	glog.Infoln("GetLoadBalancerByName", name)

	concertoLBs, err := c.GetLoadBalancerList()
	if err != nil {
		glog.Error("Error in GetLoadBalancerByName: ", err)
		return nil, err
	}

	for _, lb := range concertoLBs {
		if lb.Name == name {
			glog.Infof("GetLoadBalancerByName got %#v", lb)
			return &lb, nil
		}
	}

	glog.Infof("GetLoadBalancerByName did not find %s", name)
	return nil, nil
}

func (c *concertoAPIServiceREST) DeleteLoadBalancerById(id string) error {
	glog.Infoln("DeleteLoadBalancerById", id)

	_, status, err := c.client.Delete("/kaas/load_balancers/" + id)
	if err != nil {
		glog.Error("Error in GetLoadBalancerByName: ", err)
		return err
	}
	if status == 200 || status == 204 {
		glog.Infof("DeleteLoadBalancerById successful: %s", id)
		return nil
	}
	return LoadBalancerDeleteError
}

func (c *concertoAPIServiceREST) RegisterInstancesWithLoadBalancer(loadBalancerId string, ips []string) error {
	glog.Infoln("RegisterInstancesWithLoadBalancer", loadBalancerId, ips)
	for _, ip := range ips {
		err := c.registerInstanceWithLoadBalancer(loadBalancerId, ip)
		if err != nil {
			glog.Error("Error in RegisterInstancesWithLoadBalancer: ", err)
			return err
		}
	}
	glog.Infoln("RegisterInstancesWithLoadBalancer successful")
	return nil
}

func (c *concertoAPIServiceREST) registerInstanceWithLoadBalancer(loadBalancerId string, ip string) error {
	instance, err := c.GetInstanceByIP(ip)
	if err != nil {
		glog.Error("Error in registerInstanceWithLoadBalancer: ", err)
		return err
	}
	jsonNode := instance.toNode().toJson()
	body, status, err := c.client.Post(fmt.Sprintf("/kaas/load_balancers/%s/nodes", loadBalancerId), jsonNode)
	if err != nil {
		glog.Error("Error in registerInstanceWithLoadBalancer: ", err)
		return err
	}
	if status != 201 {
		glog.Errorf("HTTP %s in registerInstanceWithLoadBalancer: %s", status, string(body))
		return LoadBalancerRegisterInstanceError
	}
	glog.Infof("registerInstanceWithLoadBalancer successful: added %s to %s", ip, loadBalancerId)
	return nil
}

func (c *concertoAPIServiceREST) DeregisterInstancesFromLoadBalancer(loadBalancerId string, ips []string) error {
	glog.Infoln("DeregisterInstancesFromLoadBalancer", loadBalancerId, ips)
	for _, ip := range ips {
		err := c.deregisterInstanceFromLoadBalancer(loadBalancerId, ip)
		if err != nil {
			glog.Error("Error in DeregisterInstancesFromLoadBalancer: ", err)
			return err
		}
	}
	glog.Infoln("DeregisterInstancesFromLoadBalancer successful")
	return nil
}

func (c *concertoAPIServiceREST) deregisterInstanceFromLoadBalancer(loadBalancerId string, ip string) error {
	node, err := c.GetNodeByIP(loadBalancerId, ip)
	if err != nil {
		glog.Error("Error in deregisterInstanceFromLoadBalancer: ", err)
		return err
	}
	_, status, err := c.client.Delete(fmt.Sprintf("/kaas/load_balancers/%s/nodes/%s", loadBalancerId, node.ID))
	if err != nil {
		glog.Error("Error in deregisterInstanceFromLoadBalancer: ", err)
		return err
	}
	if status == 200 || status == 204 {
		glog.Infof("deregisterInstanceFromLoadBalancer successful: removed %s from %s", ip, loadBalancerId)
		return nil
	}
	return LoadBalancerDeregisterInstanceError
}

func (c *concertoAPIServiceREST) GetLoadBalancerNodes(loadBalancerId string) ([]ConcertoLoadBalancerNode, error) {
	glog.Infoln("GetLoadBalancerNodes", loadBalancerId)

	var nodes []ConcertoLoadBalancerNode

	data, status, err := c.client.Get(fmt.Sprintf("/kaas/load_balancers/%s/nodes", loadBalancerId))
	if err != nil {
		glog.Error("Error in GetLoadBalancerNodes: ", err)
		return nil, err
	}

	if status == 404 {
		return nil, errors.New("Load balancer not found " + loadBalancerId)
	}

	err = json.Unmarshal(data, &nodes)
	if err != nil {
		glog.Error("Error in GetLoadBalancerNodes: ", err)
		return nil, err
	}

	glog.Infof("GetLoadBalancerNodes received json : %s", string(data))

	return nodes, nil
}

func (c *concertoAPIServiceREST) GetLoadBalancerNodesAsIPs(loadBalancerId string) (nodeips []string, e error) {
	glog.Infoln("GetLoadBalancerNodes", loadBalancerId)

	nodes, err := c.GetLoadBalancerNodes(loadBalancerId)
	if err != nil {
		glog.Error("Error in GetLoadBalancerNodesAsIPs: ", err)
		return nil, err
	}

	for _, node := range nodes {
		nodeips = append(nodeips, node.IP)
	}

	glog.Infof("GetLoadBalancerNodesAsIPs got %v nodes : %#v", len(nodeips), nodeips)
	return
}

func (c *concertoAPIServiceREST) CreateLoadBalancer(name string, port int, nodePort int) (*ConcertoLoadBalancer, error) {
	glog.Infoln("CreateLoadBalancer", name, port)

	lb := ConcertoLoadBalancer{
		Name:     name,
		FQDN:     name,
		Port:     port,
		NodePort: nodePort,
		Protocol: "tcp",
	}
	data, status, err := c.client.Post("/kaas/load_balancers", lb.toJson())
	if err != nil {
		glog.Error("Error in CreateLoadBalancer: ", err)
		return nil, err
	}
	if status != 201 {
		return nil, fmt.Errorf("HTTP %v when creating load balancer %s", status, name)
	}

	err = json.Unmarshal(data, &lb) // So that we get the Id
	if err != nil {
		glog.Error("Error in CreateLoadBalancer: ", err)
		return nil, err
	}

	glog.Infof("CreateLoadBalancer successful: %v", lb)
	return &lb, nil
}

func (c *concertoAPIServiceREST) GetInstanceByIP(ip string) (ConcertoInstance, error) {
	glog.Infoln("GetInstanceByIP", ip)

	concertoInstances, err := c.GetInstanceList()
	if err != nil {
		glog.Error("Error in GetInstanceByIP: ", err)
		return ConcertoInstance{}, err
	}

	for _, instance := range concertoInstances {
		if instance.PublicIP == ip {
			glog.Infof("GetInstanceByIP got %#v", instance)
			return instance, err
		}
	}

	glog.Infof("GetInstanceByIP did not find %#v", ip)
	return ConcertoInstance{}, cloudprovider.InstanceNotFound
}

func (c *concertoAPIServiceREST) GetNodeByIP(loadBalancerId, ip string) (ConcertoLoadBalancerNode, error) {
	glog.Infoln("GetNodeByIP", ip)

	lbNodes, err := c.GetLoadBalancerNodes(loadBalancerId)
	if err != nil {
		glog.Error("Error in GetNodeByIP: ", err)
		return ConcertoLoadBalancerNode{}, err
	}

	for _, node := range lbNodes {
		if node.IP == ip {
			glog.Infof("GetNodeByIP got %#v", node)
			return node, err
		}
	}

	glog.Infof("GetNodeByIP did not find %#v", ip)
	return ConcertoLoadBalancerNode{}, fmt.Errorf("Node %s not found in load balancer %s", ip, loadBalancerId)
}

func (ci ConcertoInstance) toNode() ConcertoLoadBalancerNode {
	var node ConcertoLoadBalancerNode
	node.IP = ci.PublicIP
	return node
}

func (cn ConcertoLoadBalancerNode) toJson() []byte {
	b, err := json.Marshal(cn)
	if err != nil {
		return nil
	}
	return b
}

func (lb ConcertoLoadBalancer) toJson() []byte {
	b, err := json.Marshal(lb)
	if err != nil {
		return nil
	}
	return b
}
