// Copyright 2019 Oath, Inc.
// Licensed under the terms of the Apache Version 2.0 License. See LICENSE file for terms.

// Package util helps to setup Kubernetes resource related operations.
// All kubernetes resource creation operation goes here.
// In addition, any common tests related library functions goes here.
package util

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	g "github.com/onsi/ginkgo"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultNamespace      = "k8s-sec-check"
	defaultServiceAccount = "k8s-sec-check"
)

// TargetNamespace represents Kubernetes namespace to run tests
var TargetNamespace = getTargetNamespace()

// TargetServiceAccount represents Kubernetes service account
var TargetServiceAccount = getTargetServiceAccount()

// getTargetNamespace returns the value of KUBE_NAMESPACE,
// or if that is not defined, set the default namespace
func getTargetNamespace() string {
	targetNamespace := os.Getenv("KUBE_NAMESPACE")
	if targetNamespace != "" {
		return targetNamespace
	}
	return defaultNamespace

}

// getTargetServiceAccount returns the value of KUBE_SERVICEACCOUNT,
// or if that is not defined, set the default service account
func getTargetServiceAccount() string {
	targetServiceAccount := os.Getenv("KUBE_SERVICEACCOUNT")
	if targetServiceAccount != "" {
		return targetServiceAccount
	}
	return defaultServiceAccount
}

// CreateDeployment creates kubernetes deployment
func CreateDeployment(clientset kubernetes.Interface, deployment *appsv1.Deployment, targetNamespace string) error {
	_, err := clientset.AppsV1().Deployments(targetNamespace).Create(deployment)
	if err != nil {
		return errors.New("Failed to create deployment: " + err.Error())
	}
	return nil
}

// DeleteDeployment deletes kubernetes deployment
func DeleteDeployment(clientset kubernetes.Interface, deploymentName string, targetNamespace string) error {
	propagationPolicy := metav1.DeletePropagationForeground
	err := clientset.AppsV1().Deployments(targetNamespace).Delete(deploymentName, &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil && !kerr.IsNotFound(err) {
		return errors.New("Failed to delete deployment: " + err.Error())
	}
	return nil
}

// ReplicaSetsByLabel returns rs list by labelSelectorValue and targetNamespace
func ReplicaSetsByLabel(clientset kubernetes.Interface, labelSelectorValue string, targetNamespace string) (*v1beta1.ReplicaSetList, error) {
	return clientset.ExtensionsV1beta1().
		ReplicaSets(targetNamespace).
		List(metav1.ListOptions{
			LabelSelector: labelSelectorValue,
		})
}

// IsPrivilegedContainerCreated gets the replicaset and match status failure
// condition and status failure reason
func IsPrivilegedContainerCreated(rsList *v1beta1.ReplicaSetList,
	statusReason string, statusCondition string) bool {
	for _, rs := range rsList.Items {
		for _, cond := range rs.Status.Conditions {
			if cond.Reason == statusReason &&
				cond.Type == v1beta1.ReplicaSetReplicaFailure &&
				strings.Contains(cond.Message, statusCondition) {
				return false
			}
		}
	}
	log.Println("failed to find the replicaSet for the deployment")
	return true
}

// CreatePod creates kubernetes pod
func CreatePod(clientset kubernetes.Interface, pod *v1.Pod, targetNamespace string) error {
	_, err := clientset.CoreV1().Pods(targetNamespace).Create(pod)
	if err != nil {
		return errors.New("Failed to create pod: " + err.Error())
	}
	return nil
}

// DeletePod deletes kubernetes pod
func DeletePod(clientset kubernetes.Interface, podName string, targetNamespace string) error {
	propagationPolicy := metav1.DeletePropagationForeground
	err := clientset.CoreV1().Pods(targetNamespace).Delete(podName, &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil && !kerr.IsNotFound(err) {
		return errors.New("Failed to delete pod: " + err.Error())
	}
	return nil
}

// CheckReadyReplicas will wait until the resource is fully rolled out with all replicas
func CheckReadyReplicas(clientset kubernetes.Interface, deploymentName string,
	targetNamespace string, retryCount int) error {
	for i := 0; i <= retryCount; i++ {
		deployment, err := clientset.AppsV1().Deployments(targetNamespace).
			Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			return errors.New("Failed to get deployment: " + err.Error())
		}
		log.Println(g.CurrentGinkgoTestDescription().TestText + ": waiting for the ready replicas...")
		// try every 30 seconds
		time.Sleep(30 * time.Second)

		// check if number of ready replica count is matching desired replicas.
		if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
			return nil
		}
	}

	// return error after multiple attempts
	return errors.New("CheckReadyReplicas status failed even after multiple " +
		"retryCount for deployment " + deploymentName)
}

// GetStatusCondition waits until status conditions are availabe for a
// given replica set. If not found after multiple retries, return an error
// otherwise return the replicasetList instance.
func GetStatusCondition(clientset kubernetes.Interface, deploymentName string) (*v1beta1.ReplicaSetList, error) {
	retryCount := 3
	for i := 0; i <= retryCount; i++ {
		rsList, err := ReplicaSetsByLabel(clientset,
			"k8s-app="+deploymentName, TargetNamespace)
		// fail if any other error happens while fetching replicaset
		if err != nil {
			log.Println("error from ReplicaSetsByLabel")
			return rsList, err
		}
		log.Printf(g.CurrentGinkgoTestDescription().TestText+
			": waiting for replicaset and status condition to be available "+
			"for deployment: %v\n", deploymentName)
		// sleep 30 seconds, wait for replication controller to create replicaset
		time.Sleep(30 * time.Second)
		if len(rsList.Items) != 0 && len(rsList.Items[0].Status.Conditions) != 0 {
			return rsList, nil
		}
	}
	return nil,
		errors.New("failed to fetch the replicaset status condition" +
			" after multiple attempts")
}
