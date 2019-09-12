// Copyright 2019 Oath, Inc.
// Licensed under the terms of the Apache Version 2.0 License. See LICENSE file for terms.

// Package client helps to setup Kubernetes client and config
// All kubernetes connection related utilities go here.
package client

import (
	"fmt"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// KubernetesClient : Kubernetes client
	KubernetesClient kubernetes.Interface
	// RestConfig : Kubernetes rest config
	RestConfig *rest.Config
)

// GetClients retrieve the Kubernetes cluster client and restConfig
// based on KUBECONFIG file or inClusterConfig
// KUBECONFIG is an absolute file path, if not set then fallback on inClusterConfig
func GetClients() (kubernetes.Interface, *rest.Config, error) {

	var inClusterConfig bool
	var kubeConfig string

	// if KUBECONFIG is set then use it. use cluster config otherwise
	kubeConfig = os.Getenv("KUBECONFIG")
	if kubeConfig == "" {
		inClusterConfig = true
	} else {
		log.Println("Using KUBECONFIG: " + kubeConfig)
	}

	// if inCluserConfig is set to true
	if inClusterConfig {
		emptystr := ""
		kubeConfig = emptystr
	}

	// Build kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.Println(err.Error())
		return nil, nil, err
	}

	// generate the client based off of the config
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create k8s client from config. Error: %v", err)
	}

	log.Println("Successfully constructed k8s client")
	return client, config, nil
}
