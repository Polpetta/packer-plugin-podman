packer {
  required_plugins {
    podman = {
      version = ">=v0.1.0"
      source  = "github.com/hashicorp/podman"
    }
  }
}

source "podman-my-builder" "foo-example" {
  mock = local.foo
}

source "podman-my-builder" "bar-example" {
  mock = local.bar
}

build {
  sources = [
    "source.podman-my-builder.foo-example",
  ]

  source "source.podman-my-builder.bar-example" {
    name = "bar"
  }

  provisioner "podman-my-provisioner" {
    only = ["podman-my-builder.foo-example"]
    mock = "foo: ${local.foo}"
  }

  provisioner "podman-my-provisioner" {
    only = ["podman-my-builder.bar"]
    mock = "bar: ${local.bar}"
  }

  post-processor "podman-my-post-processor" {
    mock = "post-processor mock-config"
  }
}
