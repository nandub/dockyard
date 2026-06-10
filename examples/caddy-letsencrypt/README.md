# Caddy Let's Encrypt Example

This package demonstrates automatic HTTPS with Caddy using standard Docker Compose.

## Configurable parameters

- `caddy.image`
- `caddy.tag`
- `caddy.email`
- `site.hostname`
- `upstream.url`
- `service.httpPort`
- `service.httpsPort`

## Notes

Caddy obtains certificates only when the hostname resolves to this host and ports 80/443 are reachable from the public internet.

For local testing, use `dockyard render` or `dockyard config` first. Do not expect Let's Encrypt issuance to work with `localhost`.

## Example

```powershell
.\bin\dockyard.exe values template .\examples\caddy-letsencrypt -o ..\deploy-values\caddy.yaml
.\bin\dockyard.exe config .\examples\caddy-letsencrypt -f ..\deploy-values\caddy.yaml
```
