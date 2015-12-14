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
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
)

type restService struct {
	config ConcertoConfig
	client *http.Client
}

func newRestService(config ConcertoConfig) (*restService, error) {
	glog.Infoln("newRestService")

	client, err := httpClient(config)
	if err != nil {
		return nil, err
	}

	return &restService{config, client}, nil
}

func httpClient(config ConcertoConfig) (*http.Client, error) {
	glog.Infoln("httpClient")

	// Loads Clients Certificates and creates and 509KeyPair
	cert, err := tls.LoadX509KeyPair(config.Connection.Cert, config.Connection.Key)
	if err != nil {
		return nil, err
	}

	// Creates a client with specific transport configurations
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			// InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: transport}

	return client, nil
}

func (r *restService) Post(path string, json []byte) ([]byte, int, error) {
	glog.Infof("Posting %s with %s", path, string(json))
	output := strings.NewReader(string(json))
	response, err := r.client.Post(r.config.Connection.APIEndpoint+path, "application/json", output)
	if err != nil {
		return nil, -1, err
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	glog.Infof("Post response: [%v] '%s'", response.StatusCode, body)

	return body, response.StatusCode, err
}

func (r *restService) Delete(path string) ([]byte, int, error) {
	glog.Infof("Deleting %s", path)

	request, err := http.NewRequest("DELETE", r.config.Connection.APIEndpoint+path, nil)
	if err != nil {
		return nil, -1, err
	}
	response, err := r.client.Do(request)
	if err != nil {
		return nil, -1, err
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	glog.Infof("Delete response: [%v] '%s'", response.StatusCode, body)

	return body, response.StatusCode, nil
}

func (r *restService) Get(path string) ([]byte, int, error) {
	glog.Infof("Getting '%s'", path)
	response, err := r.client.Get(r.config.Connection.APIEndpoint + path)
	if err != nil {
		return nil, -1, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, err
	}

	glog.Infof("Get response: [%v] '%s'", response.StatusCode, body)
	return body, response.StatusCode, nil
}
