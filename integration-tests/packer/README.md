## Creating a new version

Prerequisites:
- Packer
- hcloud API token
- Pulumi (aligned to the repo’s version)
- Talos CLI (aligned to the repo’s version)

Run the following command to build a new image version:

```
go run ./run-packer.go -var=talos_version=v1.12.0 -template ./hcloud-talos.pkr.hcl
```

`talos_version` defaults to `v1.12.0` in `hcloud-talos.pkr.hcl`; override with `-var talos_version=<version>` if needed.

When upgrading Talos or Pulumi versions in this repo, also bump the versions in `hcloud-talos.pkr.hcl` so the test images stay in sync with the integration framework.***
