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

import "errors"

var NotYetImplemented = errors.New("Not yet implemented")
var NoConfigFile = errors.New("A Concerto cloud provider config file is required")
var UnsupportedOperation = errors.New("Unsupported operation")
var LoadBalancerCreateError = errors.New("Could not create load balancer")
var LoadBalancerDeleteError = errors.New("Could not delete load balancer")
var LoadBalancerRegisterInstanceError = errors.New("Could not register instance with load balancer")
var LoadBalancerDeregisterInstanceError = errors.New("Could not deregister instance from load balancer")
var LoadBalancerUnsupportedAffinityError = errors.New("Unsupported load balancer affinity")
var LoadBalancerUnsupportedExternalIPError = errors.New("externalIP cannot be specified for Concerto Load Balancer")
var LoadBalancerUnsupportedNumberOfPortsError = errors.New("Concerto Load Balancer only supports one single port")
