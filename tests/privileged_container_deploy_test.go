// Copyright 2019 Oath, Inc.
// Licensed under the terms of the Apache Version 2.0 License. See LICENSE file for terms.

package tests

import (
	"fmt"

	"github.com/yahoo/k8s-sec-check/client"
	"github.com/yahoo/k8s-sec-check/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//Test case(s):
//  Do not admit privileged containers
//  Do not admit containers wishing to share the host network namespace
//  Do not admit containers wishing to share the host IPC namespace
//  Do not admit containers wishing to share the host process ID namespace

//	Create Privileged pod with host network, host PID and host IPC
//	and assert that it fails to create the pod and return the
//	appropriate error for each. Following document has information on privileged
//  container creation.

// https://kubernetes.io/docs/concepts/policy/pod-security-policy/#privileged
// https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces

//Sample error output:
//	status:
//	conditions:
//		- lastTransitionTime: 2019-05-15T00:21:08Z
//	message: 'pods "nginx-privileged-container-deployment-test-7c68b45968-" is forbidden:
//		unable to validate against any pod security policy: [spec.securityContext.hostNetwork:
//		Invalid value: true: Host network is not allowed to be used spec.securityContext.hostPID:
//		Invalid value: true: Host PID is not allowed to be used spec.securityContext.hostIPC:
//		Invalid value: true: Host IPC is not allowed to be used spec.containers[0].securityContext.privileged:
//		Invalid value: true: Privileged containers are not allowed spec.containers[0].securityContext.containers[0].hostPort:
//		Invalid value: 4080: Host port 4080 is not allowed to be used. Allowed ports:
//		[] spec.securityContext.hostNetwork: Invalid value: true: Host network is not
//		allowed to be used spec.securityContext.hostPID: Invalid value: true: Host PID
//		is not allowed to be used spec.securityContext.hostIPC: Invalid value: true:
//		Host IPC is not allowed to be used spec.containers[0].securityContext.privileged:
//		Invalid value: true: Privileged containers are not allowed spec.containers[0].securityContext.containers[0].hostPort:
//		Invalid value: 4080: Host port 4080 is not allowed to be used. Allowed ports:
//		[]]'
//	reason: FailedCreate
//		status: "True"
//	type: ReplicaFailure

var _ = Describe("creating a deployment", func() {

	var deploymentName = "nginx-privileged-container-deploy-test"

	Context("with Privileged container", func() {

		It("should return an error on creating of replicaset", func() {
			// set privileged container and host network, pid and ipc.
			deployment := GetNginxDeploymentSpec(util.TargetNamespace, deploymentName, 1, true)
			deployment.Spec.Template.Spec.HostNetwork = true
			deployment.Spec.Template.Spec.HostPID = true
			deployment.Spec.Template.Spec.HostIPC = true

			// create deployment with privilege true and replicacount set to 1
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
			// assert on Host network
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring(
					"spec.securityContext.hostNetwork: Invalid value: true: " +
						"Host network is not allowed to be used"))
			// assert on host PID
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring("spec.securityContext.hostPID: Invalid value: true: " +
					"Host PID is not allowed to be used"))
			// assert on host IPC
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring("spec.securityContext.hostIPC: Invalid value: true: " +
					"Host IPC is not allowed to be used"))
			// assert if operation is forbidden and do not admit the operation
			Expect(rsList.Items[0].Status.Conditions[0].Message).To(ContainSubstring("is forbidden: "))
			// assert on Privileged container
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring("spec.containers[0].securityContext.privileged: Invalid value: true: " +
					"Privileged containers are not allowed"))
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
