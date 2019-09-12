// Copyright 2019 Oath, Inc.
// Licensed under the terms of the Apache Version 2.0 License. See LICENSE file for terms.
package tests

import (
	"github.com/yahoo/k8s-cis-check/client"
	"github.com/yahoo/k8s-cis-check/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//Test case:
//  Do not admit privileged containers

//	Create Privileged pod with host network, host PID and host IPC
//	and assert that it fails to create the pod and return the
//	appropriate error for each.

//Sample error output:
//	Failed to create pod: pods "nginx-privileged-container-pod-test" is forbidden:
//	unable to validate against any pod security policy:
//	[spec.securityContext.hostNetwork: Invalid value: true: Host network is not allowed to be used
//	spec.securityContext.hostPID: Invalid value: true: Host PID is not allowed to be used
//	spec.securityContext.hostIPC: Invalid value: true: Host IPC is not allowed to be used
//	spec.containers[0].securityContext.containers[0].hostPort: Invalid value: 4080:
//	Host port 4080 is not allowed to be used. Allowed ports: []
//	spec.securityContext.hostNetwork: Invalid value: true: Host network is not allowed to be used
//	spec.securityContext.hostPID: Invalid value: true: Host PID is not allowed to be used
//	spec.securityContext.hostIPC: Invalid value: true: Host IPC is not allowed to be used
//	spec.containers[0].securityContext.containers[0].hostPort: Invalid value: 4080:
//	Host port 4080 is not allowed to be used. Allowed ports: []]

var _ = Describe("creating a pod", func() {

	var PodName = "nginx-privileged-container-pod-test"

	Context("with Privileged security context", func() {

		It("should return an error on creating pod", func() {
			// create pod with privilege true
			pod := GetNginxPodSpec(util.TargetNamespace, PodName, false)
			pod.Spec.HostNetwork = true
			pod.Spec.HostPID = true
			pod.Spec.HostIPC = true
			err := util.CreatePod(client.KubernetesClient, pod, util.TargetNamespace)
			// assert for an existence of an error
			Ω(err).ShouldNot(BeNil())
			// assert on Host network
			Expect(err.Error()).To(
				ContainSubstring("spec.securityContext.hostNetwork: Invalid value: true: " +
					"Host network is not allowed to be used"))
			// assert on host PID
			Expect(err.Error()).To(
				ContainSubstring("spec.securityContext.hostPID: Invalid value: true: " +
					"Host PID is not allowed to be used"))
			// assert on host IPC
			Expect(err.Error()).To(
				ContainSubstring("spec.securityContext.hostIPC: Invalid value: true: " +
					"Host IPC is not allowed to be used"))
			// assert if operation is forbidden and do not admit the operation
			Expect(err.Error()).To(ContainSubstring("is forbidden: "))
		})
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			// delete the pod once the test is complete, if test created the pod successfully
			err := util.DeletePod(client.KubernetesClient, PodName, util.TargetNamespace)
			if err != nil {
				GinkgoT().Logf("%s Failed in teardown. %v:%d err: %v %s",
					redColor,
					CurrentGinkgoTestDescription().FileName,
					CurrentGinkgoTestDescription().LineNumber,
					err.Error(),
					defaultStyle,
				)
			}
		}
	})
})
