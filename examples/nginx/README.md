# nginx example

A minimal runnable Dockyard package for local smoke testing.

```powershell
dockyard values template ./examples/nginx -o ../deploy-values/nginx.local.yaml --force
dockyard lint ./examples/nginx -f ../deploy-values/nginx.local.yaml
dockyard install nginx-demo ./examples/nginx -f ../deploy-values/nginx.local.yaml
dockyard status nginx-demo --compose-ps
Invoke-WebRequest http://localhost:8080 -UseBasicParsing
dockyard uninstall nginx-demo
```
