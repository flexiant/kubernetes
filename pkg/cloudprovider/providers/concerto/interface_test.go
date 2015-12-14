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
	"testing"
)

func TestProviderName(t *testing.T) {
	concerto := &ConcertoCloud{}
	name := concerto.ProviderName()
	if name != "concerto" {
		t.Errorf("Unexpected provider name: '%v' (expecting 'concerto')", name)
	}
}

func TestInstances(t *testing.T) {
	concerto := &ConcertoCloud{}
	instances, ok := concerto.Instances()
	if !ok {
		t.Errorf("Unexpected error fetching Concerto 'Instances' component")
	}
	if instances != concerto {
		t.Errorf("Unexpected error fetching Concerto 'Instances' component")
	}
}

func TestTCPLoadBalancer(t *testing.T) {
	concerto := &ConcertoCloud{}
	loadbalancers, ok := concerto.TCPLoadBalancer()
	if !ok {
		t.Errorf("Unexpected error fetching Concerto 'TCPLoadBalancer' component")
	}
	if loadbalancers != concerto {
		t.Errorf("Unexpected error fetching Concerto 'TCPLoadBalancer' component")
	}
}

func TestZones(t *testing.T) {
	concerto := &ConcertoCloud{}
	zones, ok := concerto.Zones()
	if !ok {
		t.Errorf("Unexpected error fetching Concerto 'Zones' component")
	}
	if zones != concerto {
		t.Errorf("Unexpected error fetching Concerto 'Zones' component")
	}
}