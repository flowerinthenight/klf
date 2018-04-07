# Overview

I use `kubectl logs -f {pod-name} [container-name]` heavily when developing/debugging applications in Kubernetes. It quite difficult if I have multiple replicas though. `kubectl logs -l key=value` works to some extent but not with the `-f` flag. I just wish I could do something like `kubectl logs -f {service-name} [container-name]` or `kubectl logs -f {deployment-name} [container-name]`. This simple wrapper tool exactly does that.
