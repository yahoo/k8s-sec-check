// Copyright 2019 Oath, Inc.
// Licensed under the terms of the Apache Version 2.0 License. See LICENSE file for terms.
package tests

import (
	"log"

	"github.com/yahoo/k8s-sec-check/client"
	"github.com/yahoo/k8s-sec-check/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//Test case:
//  User impersonation

//	Create a deployment with impersonated user and it should return an error and failed
//  to create the deployment. More information on impersonation on kubernetes is available
//  on the following page.

// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation

// Following is a table test where different users can be used to run the same test case.
// Test runs as a cd user and a random user and checks if impersonation operation can be
// successful or not.

var _ = Describe("creating a deployment", func() {

	var deploymentName = "nginx-priv-impersonation"

	Context("as Impersonated user or group with Privileged container enabled", func() {

		DescribeTable("impersonate as user to create prvileged deployment",
			func(ImpersonationUser string) {
				client.RestConfig.Impersonate = rest.ImpersonationConfig{
					// UserName is the username to impersonate on each request.
					UserName: ImpersonationUser,
					// Groups are the groups to impersonate on each request.
					Groups: []string{"system:masters"},
					// Extra is a free-form field which can be used to link some authentication information
					// to authorization information.  This field allows you to impersonate it.
					//Extra: ImpersonateUserExtra,
				}
				log.Println(CurrentGinkgoTestDescription().FullTestText +
					": Creating deployment as user: " +
					client.RestConfig.Impersonate.UserName)

				kc, err := kubernetes.NewForConfig(client.RestConfig)
				if err != nil {
					Fail(CurrentGinkgoTestDescription().FullTestText +
						": Failed to get kubernetes client set" + err.Error())
				}

				// create deployment with privilege true and replicacount set to 1
				deployment := GetNginxDeploymentSpec(util.TargetNamespace, deploymentName, 1, true)
				deployment.Spec.Template.Spec.HostNetwork = true
				deployment.Spec.Template.Spec.HostPID = true
				deployment.Spec.Template.Spec.HostIPC = true

				// create deployment with whitelisted service account name matching with the namespace
				deployment.Spec.Template.Spec.ServiceAccountName = util.TargetServiceAccount

				err = util.CreateDeployment(kc, deployment, util.TargetNamespace)

				// it should return an error as operation is forbidden.
				// NOTE: if you're a cluster admin and running this test, if will fail
				// as cluster admin user as a permission to impersonate.
				Ω(err).ShouldNot(BeNil())
				// if error happens, make sure it matches with forbidden message
				Expect(err.Error()).To(MatchRegexp("Failed to create deployment: users \".*\" is forbidden: " +
					"User \".*\" cannot impersonate resource \"users\" in API group \"\" at the cluster scope"))
			},
			// as an random non-existent user
			Entry("Impersonate as wronguser", "wronguser"),
		)
	})

	AfterEach(func() {
		// delete the deployment once the test is complete
		err := util.DeleteDeployment(client.KubernetesClient, deploymentName, util.TargetNamespace)
		if err != nil {
			GinkgoT().Logf("%s Failed in teardown. %v:%d err: %v %s",
				redColor,
				CurrentGinkgoTestDescription().FileName,
				CurrentGinkgoTestDescription().LineNumber,
				err.Error(),
				defaultStyle,
			)
		}
	})
})