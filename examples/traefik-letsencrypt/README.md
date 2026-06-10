# Traefik Let's Encrypt Example

This package demonstrates automatic HTTPS using Traefik Docker provider labels.

## Configurable parameters

- `traefik.image`
- `traefik.tag`
- `traefik.email`
- `traefik.useStagingCA`
- `site.hostname`
- `service.httpPort`
- `service.httpsPort`
- `whoami.image`
- `whoami.tag`

## Security notes

This example mounts the Docker socket read-only so Traefik can discover containers and labels. That is powerful and security-sensitive.

The package policy explicitly allows Docker socket mounts for this example. Do not copy that policy into packages that do not need Docker API access.

## Let's Encrypt notes

The example uses the Let's Encrypt staging CA by default in the Compose command to avoid rate limits during testing. Review the command before production use.
