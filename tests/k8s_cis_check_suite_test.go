// Copyright 2019 Oath, Inc.
// Licensed under the terms of the Apache Version 2.0 License. See LICENSE file for terms.

package tests

import (
	"log"
	"testing"

	"github.com/yahoo/k8s-cis-check/client"
	"github.com/yahoo/k8s-cis-check/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	defaultStyle = "\x1b[0m"
	redColor     = "\x1b[91m"
	//greenColor   = "\x1b[32m"
)

var err error

var _ = BeforeSuite(func() {

	client.KubernetesClient, client.RestConfig, err = client.GetClients()
	if err != nil {
		GinkgoT().
			Error("Failed in before setup " + err.Error())
	}
	log.Println("Target Namespace: " + util.TargetNamespace)
	log.Println("Target Service Account: " + util.TargetServiceAccount)
})

var _ = AfterSuite(func() {
	log.Println("Done running K8s Cluster Check Tests")
})

func TestK8sCISCheckTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s Cluster CIS Check")
}
