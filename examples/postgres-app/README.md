# postgres-app example

A multi-container example showing environment-backed secret values.

```powershell
dockyard values template ./examples/postgres-app -o ../deploy-values/postgres-app.local.yaml --force
dockyard env template ./examples/postgres-app --sensitive-only -o ../deploy-values/postgres-app.env.example --force
Copy-Item ../deploy-values/postgres-app.env.example ../deploy-values/postgres-app.env
# edit ../deploy-values/postgres-app.env and set POSTGRES_APP_PASSWORD
dockyard config ./examples/postgres-app -f ../deploy-values/postgres-app.local.yaml --env-file ../deploy-values/postgres-app.env
dockyard install postgres-demo ./examples/postgres-app -f ../deploy-values/postgres-app.local.yaml --env-file ../deploy-values/postgres-app.env
```
