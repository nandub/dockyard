# Security Policy

Do not report suspected vulnerabilities in public issues until they have been triaged.

Dockyard should not be used to run untrusted packages without review. A Dockyard package can define containers, volumes, ports, and images that affect the host through Docker Compose.

Secure defaults:

- release state uses user-owned directories by default
- generated files are written with restrictive permissions where practical
- package archives reject symlinks, path traversal, and common secret-like files
- render diagnostics mask sensitive-looking keys
- install refuses HIGH policy findings unless `--allow-risk` is explicitly passed
