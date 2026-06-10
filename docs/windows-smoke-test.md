# Windows smoke test

This smoke test was validated with Docker Desktop and PowerShell.

```powershell
go mod tidy
go test ./...
go build -o bin/dockyard.exe ./cmd/dockyard

.\bin\dockyard.exe doctor
.\bin\dockyard.exe init .\example-app
.\bin\dockyard.exe values template .\example-app -o .\deploy-values\local.yaml
.\bin\dockyard.exe lint .\example-app -f .\deploy-values\local.yaml
.\bin\dockyard.exe render .\example-app -f .\deploy-values\local.yaml --validate-compose
.\bin\dockyard.exe install example .\example-app -f .\deploy-values\local.yaml
.\bin\dockyard.exe status example --compose-ps
```

Test the HTTP endpoint with one of these commands:

```powershell
Invoke-WebRequest http://localhost:8080 -UseBasicParsing
```

or:

```powershell
curl.exe http://localhost:8080
```

Older Windows PowerShell aliases `curl` to `Invoke-WebRequest`, which may show a script parsing warning unless `-UseBasicParsing` is used.

Clean up:

```powershell
.\bin\dockyard.exe uninstall example
```

To include stopped containers in status output:

```powershell
.\bin\dockyard.exe status example --compose-ps --all
```
