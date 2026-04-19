# Kubernetes Operator

<br/>

### KubeBuilder

https://kubebuilder.io/quick-start.html

<br/>

```shell
$ kubebuilder init --domain slurm.io --repo slurm.io/cronjob
```

<br/>

```shell
$ kubebuilder create api --group batch --version v1 --kind CronJob
INFO Create Resource [y/n] 
y
INFO Create Controller [y/n] 
y
```

<br/>

**Меняем:**  

```
cronjob_types.go
cronjob_controller.go
```

<br/>

**Деплоим оператор:**

В Makefile поменял

```
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@latest)

kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5@v5.4.3)
```

<br/>

```shell
$ make manifests
```

<br/>

```shell
$ make install
```

<br/>

```shell
$ kubectl get cronjobs.batch.slurm.io
No resources found in default namespace.
```

<br/>

```shell
$ make run
```

<br/>

```shell
$ kubectl apply -f config/samples/batch_v1_cronjob.yaml 
cronjob.batch.slurm.io/cronjob-sample created
```

<br/>

```shell
$ kubectl get cronjobs.batch.slurm.io
NAME             AGE
cronjob-sample   22s
```

<br/>

```shell
$ kubectl describe cronjobs.batch.slurm.io/cronjob-sample
```

<br/>

```
^C
```

<br/>

```shell
$ make docker-build docker-push IMG=webmakaka/cronjob-controller:1.0.0
```

<br/>

### OperatorFramework

https://sdk.operatorframework.io/

<br/>

### MetaController

https://metacontroller.github.io/metacontroller/


