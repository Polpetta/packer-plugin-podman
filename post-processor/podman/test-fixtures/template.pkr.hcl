source "null" "basic-example" {
  communicator = "none"
}

build {
  sources = [
    "source.null.basic-example"
  ]

  post-processor "podman-my-post-processor" {
    mock = "my-mock-config"
  }
}
