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
}

variable "talos_schematic_id" {
  type    = string
}

locals {
  image_arm = "https://factory.talos.dev/image/${var.talos_schematic_id}/${var.talos_version}/hcloud-arm64.raw.xz"
  image_x86 = "https://factory.talos.dev/image/${var.talos_schematic_id}/${var.talos_version}/hcloud-amd64.raw.xz"

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

# Source for the Talos ARM image
source "hcloud" "talos-arm" {
  rescue       = "linux64"
  image        = "debian-11"
  location     = "fsn1"
  server_type  = "cax11"
  ssh_username = "root"

  snapshot_name   = "Talos Linux ${var.talos_version} (schemaID: ${var.talos_schematic_id}) for testing purposes"
  snapshot_labels = {
    os      = "talos",
    version = "${var.talos_version}",
    arch    = "arm",
    testing = "true"
    creator = "pulumi-talos-cluster"
  }
}

# Source for the Talos x86 image
source "hcloud" "talos-x86" {
  rescue       = "linux64"
  image        = "debian-11"
  server_type  = "cx22"
  location     = "fsn1"
  ssh_username = "root"
  snapshot_name   = "Talos Linux ${var.talos_version} (schemaID: ${var.talos_schematic_id}) for testing purposes"

  snapshot_labels = {
    os      = "talos",
    version = "${var.talos_version}",
    arch    = "x86",
    testing = "true"
    creator = "pulumi-talos-cluster"
  }
}

# Build the Talos ARM snapshot
build {
  sources = ["source.hcloud.talos-arm"]

  # Download the Talos ARM image
  provisioner "shell" {
    inline = ["${local.download_image}${local.image_arm}"]
  }

  # Write the Talos ARM image to the disk
  provisioner "shell" {
    inline = [local.write_image]
  }

  # Clean-up
  provisioner "shell" {
    inline = [local.clean_up]
  }
}

# Build the Talos x86 snapshot
build {
  sources = ["source.hcloud.talos-x86"]

  # Download the Talos x86 image
  provisioner "shell" {
    inline = ["${local.download_image}${local.image_x86}"]
  }

  # Write the Talos x86 image to the disk
  provisioner "shell" {
    inline = [local.write_image]
  }

  # Clean-up
  provisioner "shell" {
    inline = [local.clean_up]
  }
}
