// Copyright 2019 Oath, Inc.
// Licensed under the terms of the Apache Version 2.0 License. See LICENSE file for terms.

package tests

import (
	"fmt"

	"github.com/yahoo/k8s-cis-check/client"
	"github.com/yahoo/k8s-cis-check/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
)

//Test case:
//  Do not admit containers with dangerous capabilities

//	Create Privileged pod with capabilities,
//	and assert that it fails to create the pod and return the
//	appropriate error for each capabilities. Following document has
//	information on capabilties on k8s container.

//http://man7.org/linux/man-pages/man7/capabilities.7.html
//https://kubernetes.io/docs/concepts/policy/pod-security-policy/#capabilities
//https://docs.docker.com/engine/reference/run/#runtime-privilege-and-linux-capabilities

//sample output:
//status:
//conditions:
//	- lastTransitionTime: 2019-05-15T21:00:04Z
//message: 'pods "nginx-privileged-container-capability-not-allowed-deploy-test-69fc5b5b7d-"
//	is forbidden: unable to validate against any pod security policy: [
//	capabilities.add: Invalid value: "NET_ADMIN": capability may not be added
//	capabilities.add: Invalid value: "NET_RAW": capability may not be added
//	capabilities.add: Invalid value: "SYS_PTRACE": capability may not be added
//	capabilities.add: Invalid value: "SYS_ADMIN": capability may not be added
//	spec.containers[0].securityContext.privileged: Invalid value: true: Privileged containers are not allowed
//	capabilities.add: Invalid value: "NET_ADMIN": capability may not be added
//	capabilities.add: Invalid value: "NET_RAW": capability may not be added
//	capabilities.add: Invalid value: "SYS_PTRACE": capability may not be added
//	capabilities.add: Invalid value: "SYS_ADMIN": capability may not be added
//	spec.containers[0].securityContext.privileged: Invalid value: true: Privileged containers are not allowed
//	capabilities.add: Invalid value: "NET_ADMIN": capability may not be added
//	capabilities.add: Invalid value: "NET_RAW": capability may not be added
//	capabilities.add: Invalid value: "SYS_PTRACE": capability may not be added
//	capabilities.add: Invalid value: "SYS_ADMIN": capability may not be added]'
//reason: FailedCreate
//	status: "True"
//	type: ReplicaFailure

var _ = Describe("creating a deployment", func() {

	var deploymentName = "nginx-privileged-container-capability-not-allowed-deploy-test"

	Context("with Privileged container with capabilities", func() {

		It("should return an error on creating of replicaset", func() {
			// set the deployment with privilege true and replicacount and other linux capabilities
			deployment := GetNginxDeploymentSpec(util.TargetNamespace, deploymentName, 1, true)
			deployment.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities = &v1.Capabilities{}
			deployment.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities.
				Add = []v1.Capability{"NET_ADMIN", "NET_RAW", "SYS_PTRACE", "SYS_ADMIN", "KILL"}

			// create deployment with above set configs
			err := util.CreateDeployment(client.KubernetesClient, deployment, util.TargetNamespace)
			Ω(err).Should(BeNil())

			// find replicaSet for the deployment by label
			// retry 3 times, every 10 seconds
			rsList, err := util.GetStatusCondition(client.KubernetesClient, deploymentName)
			if err != nil {
				Fail(CurrentGinkgoTestDescription().TestText + ":" + err.Error())
			}

			// if not error, but failed to find the replicaset, fail.
			Expect(len(rsList.Items) == 1).Should(Equal(true))
			fmt.Println(len(rsList.Items[0].Status.Conditions))
			Ω(rsList.Items[0].Status).ShouldNot(BeNil())
			Expect(len(rsList.Items[0].Status.Conditions) == 1).Should(Equal(true))
			// check if privileged container is failed to create with
			// failure reason and condition
			Expect(rsList.Items[0].Status.Conditions[0].Reason).To(Equal("FailedCreate"))
			Expect(rsList.Items[0].Status.Conditions[0].Type).To(Equal(v1beta1.ReplicaSetReplicaFailure))
			Expect(rsList.Items[0].Status.Conditions[0].Status).To(Equal(v1.ConditionStatus("True")))
			// assert if operation is forbidden and do not admit the operation
			Expect(rsList.Items[0].Status.Conditions[0].Message).To(ContainSubstring("is forbidden: "))
			// assert on Privileged container

			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring("Privileged containers are not allowed"))
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring("capabilities.add: Invalid value: \"NET_ADMIN\": capability may not be added"))
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring("capabilities.add: Invalid value: \"NET_RAW\": capability may not be added"))
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring("capabilities.add: Invalid value: \"SYS_PTRACE\": capability may not be added"))
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring("capabilities.add: Invalid value: \"SYS_ADMIN\": capability may not be added"))
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring("capabilities.add: Invalid value: \"KILL\": capability may not be added"))

		})
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
