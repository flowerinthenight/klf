## Overview

I use `kubectl logs -f {pod-name} [container-name]` heavily when developing/debugging applications in Kubernetes. It quite difficult if I have multiple replicas though. `kubectl logs -l key=value` works to some extent but not with the `-f` flag. I just wish I could do something like:
```bash
$ kubectl logs -f {service-name} [container-name]
```
or
```bash
$ kubectl logs -f {deployment-name} [container-name]
```

This simple wrapper tool does exactly just that.

## Installation

```bash
$ go get -u -v github.com/flowerinthenight/klf
```

## Usage

```bash
# Tail a service `svc1` with only one container in the pod:
$ klf tail svc1

Tail a deployment `deploy1` with only one container in the pod:
$ klf tail deploy1
```
