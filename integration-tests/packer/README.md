## Creating a new version

Run the following command to build a new image version:

```
go run ./run-packer.go -var=talos_version=v1.11.0 -template ./hcloud-talos.pkr.hcl
```
