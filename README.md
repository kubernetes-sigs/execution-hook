# Execution Hook

The ExecutionHook controller works with a snapshot controller to quiesce an application
beforing taking a snapshot and unquiesce an application afterwards.

## Usage

### Running on command line

For debugging, it is possible to run the hook controller on command line. For example,

```bash
hook -kubeconfig /var/run/kubernetes/admin.kubeconfig -v 5
```

### Running in a statefulset

It is necessary to create a new service account and give it enough privileges to run the hook controller. We provide .yaml files that deploy it. A real production deployment must customize them:

```bash
for i in $(find deploy/kubernetes -name '*.yaml'); do kubectl create -f $i; done
```

## Testing

Running Unit Tests:

```bash
go test -timeout 30s  github.com/kubernetes-csi/execution-hook/pkg/controller
```

## Dependency Management

```bash
dep ensure
```

To modify dependencies or versions change `./Gopkg.toml`

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

* [Slack channel](https://kubernetes.slack.com/messages/sig-storage)
* [Mailing list](https://groups.google.com/forum/#!forum/kubernetes-sig-storage)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
