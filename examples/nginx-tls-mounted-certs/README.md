# Nginx TLS with Mounted Certificates

This package demonstrates TLS using operator-provided certificate, key, and nginx configuration files.

## Configurable parameters

- `image.repository`
- `image.tag`
- `service.httpPort`
- `service.httpsPort`
- `tls.certificatePath`
- `tls.privateKeyPath`
- `tls.nginxConfigPath`

## Important notes

Use absolute host paths for mounted files. Relative paths are resolved by Docker Compose relative to the rendered release directory, not necessarily your source package directory.

This example intentionally allows host-path mounts because mounted TLS files are the point of the example. Use careful file permissions on the private key.

## Minimal nginx TLS config

Create a file such as `C:\\dockyard-certs\\default.conf` or `/opt/dockyard-certs/default.conf`:

```nginx
server {
    listen 80;
    listen 443 ssl;
    server_name example.com;

    ssl_certificate /etc/nginx/certs/fullchain.pem;
    ssl_certificate_key /etc/nginx/certs/privkey.pem;

    location / {
        return 200 "hello from nginx tls example\n";
    }
}
```
