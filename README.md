# Kubernetes Security Check (k8s-sec-check)

> Kubernetes Security Check automates the complex security checks based on CIS guidelines. 

## Table of Contents
- [Background](#background)
- [Architecture](#architecture)
- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Background

There are many tools in the open-source world that provide a way to certify the security of Kubernetes security and some tools are also implemented based on the detailed CIS guidelines. 
However, not all tools are covering complex security checks. For example. Pod Security Policy checks. Kubernetes Security Check automates the complex security checks based on CIS guidelines.
It can be further extended to add more checks as well. 

## Architecture
The Kube security check tool is a simple test suite based on Ginkgo. Once the binary is built, it can be run remotely by simply passing the KUBECONFIG environment variable which represents the path to a Kubernetes configuration file. 
Currently, it covers the following tests with respective Kubernetes fields: 
- [User impersonation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation)
  - Impersonate Kubernetes calls as a user
- Do not admit container with restricted volume e.g [flexVolume](https://kubernetes.io/docs/concepts/storage/volumes/#flexVolume), [hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath)
  - Volume - [AllowedHostPaths](https://kubernetes.io/docs/concepts/policy/pod-security-policy/#volumes-and-file-systems)
- [Pod Security Policy](https://kubernetes.io/docs/concepts/policy/pod-security-policy/): 
    - Do not admit privileged containers
       - Security Context: [privileged](https://kubernetes.io/docs/concepts/policy/pod-security-policy/#privileged)
    - Do not admit containers wishing to share the host process ID namespace
       - [hostPID](https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces)
    - Do not admit containers wishing to share the host IPC namespace
       - [hostIPC](https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces)
    - Do not admit containers wishing to share the host network namespace
       - [hostNetwork](https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces)
    - Do not admit containers with dangerous [capabilities](http://man7.org/linux/man-pages/man7/capabilities.7.html)
       -  [allowedCapabilities](https://kubernetes.io/docs/concepts/policy/pod-security-policy/#capabilities)

## Install

Make sure to set the relevant namespace, service account, and context in the kubeconfig file. 

``export KUBECONFIG=~/.kube/config``

Run the binary: 
``k8s-sec-check``

## Usage

If checks are being run remotely using the KUBECONFIG file, users must set the following environment variables. 

`KUBECONFIG`: Kubeconfig file absolute path. 
 - Set environment variable with `KUBECONFIG` to run the tests remotely. 
 - If the `KUBECONFIG` variable is not set, it sets to `INCLUSTERCONFIG` by default.

`KUBE_NAMESPACE` : Target Kubernetes namespace to run tests (default: `k8s-sec-check`)

`KUBE_SERVICEACCOUNT`: Target Kubernetes Service account to be used during tests. (default: `k8s-sec-check`)

## Maintainers
Core Team : omega-core@verizonmedia.com

## Contribute
Please refer to the [contributing file](Contributing.md) for information about how to get involved. We welcome issues, questions, and pull requests.

## License

Copyright 2019 Oath Inc. Licensed under the Apache License, Version 2.0 (the "License")