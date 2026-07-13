# OCI Package Checklist

- Local path, archive, explicit `oci://`, catalog, and shorthand precedence preserved.
- ORAS authentication remains external.
- Absolute archive paths are not passed to ORAS.
- Archive path safety and forbidden-file checks remain intact.
- JSON output remains machine-readable.
- Catalog behavior is tested when changed.
- Registry validation run or explicitly marked not run.
- Signature/provenance claims are not overstated.
