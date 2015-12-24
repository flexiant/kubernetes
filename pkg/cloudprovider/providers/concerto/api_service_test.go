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
	"testing"

	"k8s.io/kubernetes/pkg/cloudprovider"
)

func Test_GetInstanceList_Success(t *testing.T) {
	jsonList := "[{\"Id\":\"0001\"},{\"Id\":\"0002\"}]"
	restMock := buildConcertoRESTMockClient(jsonList, 200, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	instances, err := apiService.GetInstanceList()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if instances == nil {
		t.Errorf("Unexpected nil return value")
	} else if len(instances) != 2 {
		t.Errorf("Unexpected slice size: was %v but expected 2", len(instances))
	}
	if len(restMock.receivedCalls) != 1 || restMock.receivedCalls[0] != "GET /kaas/ships" {
		t.Errorf("Expected exactly one 'GET /kaas/ships' but received: %v", restMock.receivedCalls)
	}
}

func Test_GetInstanceList_NoInstances(t *testing.T) {
	restMock := buildConcertoRESTMockClient("", 404, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	instances, err := apiService.GetInstanceList()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if instances == nil {
		t.Errorf("Should return an empty slice but got nil")
	}
	if len(instances) != 0 {
		t.Errorf("Should return an empty slice but got some instances: %v", instances)
	}
}

func Test_GetInstanceByName_Success(t *testing.T) {
	jsonList := "[{\"id\":\"0001\",\"fqdn\":\"myinstance\"},{\"Id\":\"0002\",\"fqdn\":\"anotherinstance\"}]"
	restMock := buildConcertoRESTMockClient(jsonList, 200, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	instance, err := apiService.GetInstanceByName("myinstance")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if instance.Id != "0001" {
		t.Errorf("Incorrect instance: expected Id '0001' but was '%v'", instance.Id)
	}
}

func Test_GetInstanceByName_NotFound(t *testing.T) {
	jsonList := "[{\"id\":\"0003\",\"fqdn\":\"someinstance\"}]"
	restMock := buildConcertoRESTMockClient(jsonList, 200, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	_, err := apiService.GetInstanceByName("anotherinstance")
	if err == nil {
		t.Errorf("Expected to receive an error but didn't")
	} else if err != cloudprovider.InstanceNotFound {
		t.Errorf("Expected error: %v , but was %v", cloudprovider.InstanceNotFound, err)
	}
}

func Test_GetLoadBalancerList_Success(t *testing.T) {
	jsonList := "[{\"Id\":\"0001\"},{\"Id\":\"0002\"}]"
	restMock := buildConcertoRESTMockClient(jsonList, 200, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	lbs, err := apiService.GetLoadBalancerList()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if lbs == nil {
		t.Errorf("Unexpected nil return value")
	} else if len(lbs) != 2 {
		t.Errorf("Unexpected slice size: was %v but expected 2", len(lbs))
	}
	if len(restMock.receivedCalls) != 1 || restMock.receivedCalls[0] != "GET /kaas/load_balancers" {
		t.Errorf("Expected exactly one 'GET /kaas/load_balancers' but received: %v", restMock.receivedCalls)
	}
}

func Test_GetLoadBalancerList_NoInstances(t *testing.T) {
	jsonList := "[]"
	restMock := buildConcertoRESTMockClient(jsonList, 200, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	lbs, err := apiService.GetLoadBalancerList()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if lbs == nil {
		t.Errorf("Unexpected nil return value")
	} else if len(lbs) != 0 {
		t.Errorf("Unexpected slice size: was %v but expected 0", len(lbs))
	}
	if len(restMock.receivedCalls) != 1 || restMock.receivedCalls[0] != "GET /kaas/load_balancers" {
		t.Errorf("Expected exactly one 'GET /kaas/load_balancers' but received: %v", restMock.receivedCalls)
	}
}

func Test_GetLoadBalancerList_UnexpectedHTTPStatus(t *testing.T) {
	jsonList := "[]"
	restMock := buildConcertoRESTMockClient(jsonList, 500, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	_, err := apiService.GetLoadBalancerList()
	if err == nil {
		t.Errorf("Expected error but none was returned")
	}
	if len(restMock.receivedCalls) != 1 || restMock.receivedCalls[0] != "GET /kaas/load_balancers" {
		t.Errorf("Expected exactly one 'GET /kaas/load_balancers' but received: %v", restMock.receivedCalls)
	}
}

func Test_GetLoadBalancerByName_Success(t *testing.T) {
	jsonList := "[{\"id\":\"0001\",\"name\":\"myLB\"},{\"Id\":\"0002\",\"name\":\"anotherLB\"}]"
	restMock := buildConcertoRESTMockClient(jsonList, 200, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	instance, err := apiService.GetLoadBalancerByName("myLB")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if instance.Id != "0001" {
		t.Errorf("Incorrect instance: expected Id '0001' but was '%v'", instance.Id)
	}
}

func Test_GetLoadBalancerByName_NotFound(t *testing.T) {
	jsonList := "[{\"id\":\"0003\",\"name\":\"someLB\"}]"
	restMock := buildConcertoRESTMockClient(jsonList, 200, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	lb, err := apiService.GetLoadBalancerByName("anotherLB")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if lb != nil {
		t.Errorf("Expected nil but got: %v", lb)
	}
}

func Test_DeleteLoadBalancerById_Success_HTTP204(t *testing.T) {
	restMock := buildConcertoRESTMockClient("", 204, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	err := apiService.DeleteLoadBalancerById("0001")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func Test_DeleteLoadBalancerById_Success_HTTP200(t *testing.T) {
	restMock := buildConcertoRESTMockClient("", 200, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	err := apiService.DeleteLoadBalancerById("0001")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func Test_DeleteLoadBalancerById_UnexpectedHTTPStatus(t *testing.T) {
	restMock := buildConcertoRESTMockClient("", 500, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	err := apiService.DeleteLoadBalancerById("0001")
	if err == nil {
		t.Errorf("Expected error but got none")
	}
}

func Test_RegisterInstancesWithLoadBalancer(t *testing.T) {
	jsonList := "[{\"id\":\"1234\",\"public_ip\":\"1.2.3.4\"},{\"id\":\"5678\",\"public_ip\":\"5.6.7.8\"},{\"id\":\"0000\",\"public_ip\":\"0.0.0.0\"}]"
	restMock := buildConcertoRESTMockClient(jsonList, 201, nil)
	apiService := concertoAPIServiceREST{client: restMock}
	err := apiService.RegisterInstancesWithLoadBalancer("someLB", []string{"1.2.3.4", "5.6.7.8"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedCalls := []string{
		"GET /kaas/ships",
		"POST /kaas/load_balancers/someLB/nodes {\"ID\":\"\",\"public_ip\":\"1.2.3.4\"}",
		"GET /kaas/ships",
		"POST /kaas/load_balancers/someLB/nodes {\"ID\":\"\",\"public_ip\":\"5.6.7.8\"}",
	}
	if len(restMock.receivedCalls) != 4 ||
		restMock.receivedCalls[0] != expectedCalls[0] ||
		restMock.receivedCalls[1] != expectedCalls[1] ||
		restMock.receivedCalls[2] != expectedCalls[2] ||
		restMock.receivedCalls[3] != expectedCalls[3] {
		t.Errorf("Received this sequence of calls: '%v' but expected: '%v'", restMock.receivedCalls, expectedCalls)
	}
}

func TestGetLoadBalancerNodes(t *testing.T) {
	t.Skipf("Pending test implementation: GetLoadBalancerNodes")
}

func TestDeregisterInstancesFromLoadBalancer(t *testing.T) {
	t.Skipf("Pending test implementation: DeregisterInstancesFromLoadBalancer")
}

func TestCreateLoadBalancer(t *testing.T) {
	t.Skipf("Pending test implementation: CreateLoadBalancer")
}

func buildConcertoRESTMockClient(body string, status int, err error) *RESTMock {
	return &RESTMock{body: []byte(body), status: status, err: err}
}

type RESTMock struct {
	receivedCalls []string
	body          []byte
	status        int
	err           error
}

func (mock *RESTMock) Get(path string) ([]byte, int, error) {
	mock.receivedCalls = append(mock.receivedCalls, "GET "+path)
	return mock.body, mock.status, mock.err
}

func (mock *RESTMock) Post(path string, body []byte) ([]byte, int, error) {
	mock.receivedCalls = append(mock.receivedCalls, fmt.Sprintf("POST %s %s", path, string(body)))
	return mock.body, mock.status, mock.err
}

func (mock *RESTMock) Delete(path string) ([]byte, int, error) {
	mock.receivedCalls = append(mock.receivedCalls, "DELETE "+path)
	return mock.body, mock.status, mock.err
}
