source "null" "basic-example" {
  communicator = "none"
}

build {
  sources = [
    "source.null.basic-example"
  ]

  provisioner "podman-my-provisioner" {
    mock = "my-mock-config"
  }
}
