# OCI Package Checklist

- Local path, archive, explicit `oci://`, catalog, and shorthand precedence preserved.
- Registry credentials remain outside Dockyard-owned state.
- Archive layer names and media types are preserved by the embedded ORAS Go client path.
- Archive path safety and forbidden-file checks remain intact.
- JSON output remains machine-readable.
- Catalog behavior is tested when changed.
- Registry validation run or explicitly marked not run.
- Signature/provenance claims are not overstated.
