[![Test](https://github.com/h3poteto/node-manager/actions/workflows/test.yml/badge.svg)](https://github.com/h3poteto/node-manager/actions/workflows/test.yml)
[![Docker](https://github.com/h3poteto/node-manager/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/h3poteto/node-manager/actions/workflows/docker-publish.yml)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/h3poteto/node-manager)](https://github.com/h3poteto/node-manager/releases)

# Node Manager
This is a custom controller to manage nodes in your Kubernetes clusters. This controller has two main functions.

1. Refresh nodes according to specified schedule
1. Replenish nodes when nodes are terminated unexpectedly

If it is an immutable system, it doesn't matter when the nodes are replaced. But master/worker nodes stay alive if you don't care for anything. So this controller replaces nodes in a certain period.

And if you don't introduce [cluster-autoscaler](https://github.com/kubernetes/autoscaler), node compensation is left to the cloud provider. Maybe you want to avoid running out of nodes, even if the number of nodes is fixed. Node Manager replenishes master/worker nodes in this case instead of cluster-autoscaler.

**Currently this controller supports only AWS.**

## Install
You can install this controller and custom resource using helm.

```
$ helm repo add h3poteto-stable https://h3poteto.github.io/charts/stable
$ helm install my-node-manager --namespace kube-system h3poteto-stable/node-manager
```

Default values of the chart don't refresh and replenish, so please refer [chart](https://github.com/h3poteto/charts/tree/master/stable/node-manager#configuration) to set values.

## Development
Please prepare a Kubernetes cluster to install this, and export `KUBECONFIG`.

```
$ export KUBECONFIG=$HOME/.kube/config
```

At first, install CRDs in your cluster.

```
$ make install
```

Next, run controller in you machine.

```
$ make run
```

## License
The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
