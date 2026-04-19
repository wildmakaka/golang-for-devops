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

<br/>

```shell
$ make manifests
```



<br/>

```shell
$ make install
```


<br/>

### OperatorFramework

https://sdk.operatorframework.io/

<br/>

### MetaController

https://metacontroller.github.io/metacontroller/


