// Copyright 2019 Oath, Inc.
// Licensed under the terms of the Apache Version 2.0 License. See LICENSE file for terms.

package tests

import (
	"log"

	"github.com/yahoo/k8s-cis-check/client"
	"github.com/yahoo/k8s-cis-check/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
)

//Test case:
//  Do not admit container with restricted volume e.g flexVolume, hostPath

//	Create Privileged deployment with restricted volume such as hostpath and flex volume
//	and assert that it fails to create the pod and return the
//	appropriate error for each. Following document has information on privileged
//  container creation.

// https://kubernetes.io/docs/concepts/policy/pod-security-policy/#privileged
// https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes
// https://kubernetes.io/docs/concepts/policy/pod-security-policy/#volumes-and-file-systems

//Sample error output:
//	status:
//	conditions:
//	- lastTransitionTime: 2019-05-22T18:37:47Z
//	message: 'pods "nginx-privileged-container-deployment-test-6585dc5475-" is forbidden:
//	unable to validate against any pod security policy: [
// 	spec.volumes[0]: Invalid value: "hostPath": hostPath volumes are not allowed to be used
// 	spec.volumes[1]: Invalid value: "flexVolume": flexVolume volumes are not allowed to be used
// 	spec.containers[0].securityContext.privileged: Invalid value: true: Privileged containers are not allowed
// 	spec.volumes[0]: Invalid value: "hostPath": hostPath volumes are not allowed to be used
// 	spec.securityContext.volumes[1].driver: Invalid value: "kubernetes.io/lvm": Flexvolume driver is not allowed to be used
//	spec.containers[0].securityContext.privileged: Invalid value: true: Privileged containers are not allowed
//	]'
//	reason: FailedCreate
//	status: "True"
//	type: ReplicaFailure

var _ = Describe("creating a deployment", func() {

	var deploymentName = "nginx-volume-deploy-test"
	var deployment *appsv1.Deployment

	BeforeEach(func() {
		// create the deployment instance
		// set privileged container with replica count to 1
		deployment = GetNginxDeploymentSpec(util.TargetNamespace, deploymentName, 1, true)

		// host path of type directory. note: you can't get the address of a constant.
		t := v1.HostPathDirectory

		// set restricted privileged host path volume and flex volume
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
			{
				Name:      deploymentName + "hostpath",
				MountPath: "/datahostpath",
			},
			{
				Name:      deploymentName + "flex",
				MountPath: "/dataflex",
			},
		}
		deployment.Spec.Template.Spec.Volumes = []v1.Volume{
			{
				Name: deploymentName + "hostpath",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: "/datahostpath",
						Type: &t,
					},
				},
			},
			{
				Name: deploymentName + "flex",
				VolumeSource: v1.VolumeSource{
					FlexVolume: &v1.FlexVolumeSource{
						Driver: "kubernetes.io/lvm",
						FSType: "ext4",
					},
				},
			},
		}
	})

	Context("with Privileged container, host path and flex volumes", func() {

		It("should return an error on creating of replicaset", func() {

			// create deployment with privilege true and set volumes
			err := util.CreateDeployment(client.KubernetesClient, deployment, util.TargetNamespace)
			// error should be nil
			Ω(err).Should(BeNil())

			// once deployment is created, find replicaSet for the deployment by label
			rsList, err := util.GetStatusCondition(client.KubernetesClient, deploymentName)
			if err != nil {
				Fail(CurrentGinkgoTestDescription().TestText + ":" + err.Error())
			}

			// if not error, but failed to find the replica set, fail.
			Expect(len(rsList.Items) == 1).Should(Equal(true))
			log.Println(len(rsList.Items[0].Status.Conditions))

			// status should not be nil
			Ω(rsList.Items[0].Status).ShouldNot(BeNil())

			// match conditions count
			Expect(len(rsList.Items[0].Status.Conditions) == 1).Should(Equal(true))

			// check if privileged container is failed to create with
			// failure reason and condition
			Expect(rsList.Items[0].Status.Conditions[0].Reason).To(Equal("FailedCreate"))
			Expect(rsList.Items[0].Status.Conditions[0].Type).To(Equal(v1beta1.ReplicaSetReplicaFailure))
			Expect(rsList.Items[0].Status.Conditions[0].Status).To(Equal(v1.ConditionStatus("True")))

			// assert on host path volume
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring(
					"\"hostPath\": " + "hostPath volumes are not allowed to be used"))

			// assert on flex volume
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring(
					"\"flexVolume\": flexVolume volumes are not allowed to be used"))

			// assert of flex volume driver
			Expect(rsList.Items[0].Status.Conditions[0].Message).
				To(ContainSubstring(
					"\"kubernetes.io/lvm\": Flexvolume driver is not allowed to be used"))

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
