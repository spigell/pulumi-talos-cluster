# hcloud.pkr.hcl
packer {
  required_plugins {
    hcloud = {
      version = "v1.6.0"
      source  = "github.com/hetznercloud/hcloud"
    }
  }
}

variable "talos_version" {
  type    = string
  default = "v1.12.0"
}

variable "talos_schematic_id" {
  type    = string
  # With cloudflared extension
  default = "09dbcadc567d93b02a1610c70d651fadbe56aeac3aaca36bc488a38f3fffe99d"
  # The default one
  # default = "376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba"

}

variable "target" {
  description = "Which Talos image to bake (hcloud-arm64, hcloud-amd64, metal-arm64, metal-amd64)"
  type        = string
  default     = "hcloud-amd64"
}

locals {
  image_url = "https://factory.talos.dev/image/${var.talos_schematic_id}/${var.talos_version}/${var.target}.raw.xz"

  arch    = length(regexall("arm", var.target))   > 0 ? "arm"   : "x86"
  variant = length(regexall("metal", var.target)) > 0 ? "metal" : "cloud"

  # Add local variables for inline shell commands
  download_image = "wget -q --timeout=5 --waitretry=5 --tries=5 --retry-connrefused --inet4-only -O /tmp/talos.raw.xz "

  write_image = <<-EOT
    set -ex
    echo 'Talos image loaded, writing to disk... '
    xz -d -c /tmp/talos.raw.xz | dd of=/dev/sda && sync
    echo 'done.'
  EOT

  clean_up = <<-EOT
    set -ex
    echo "Cleaning-up..."
    rm -rf /etc/ssh/ssh_host_*
  EOT
}

source "hcloud" "talos" {
  rescue       = "linux64"
  image        = "debian-11"
  server_type  = local.arch == "arm" ? "cax11" : "cx22"
  location     = "fsn1"
  ssh_username = "root"

  snapshot_name   = "Talos Linux ${var.talos_version} ${local.variant} ${local.arch} (schemaID: ${var.talos_schematic_id})"
  snapshot_labels = {
    os      = "talos"
    version = var.talos_version
    arch    = local.arch
    variant = local.variant
    creator = "pulumi-talos-cluster"
  }
}


build {
  sources = ["source.hcloud.talos"]

  provisioner "shell" {
    inline = ["${local.download_image}${local.image_url}"]
  }
  provisioner "shell" { inline = [local.write_image] }
  provisioner "shell" { inline = [local.clean_up] }
}
