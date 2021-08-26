# Packer Plugin for Podman

This repository contains a Packer Plugin for Podman. It directly takes the source code from
[github.com/hashicorp/packer-plugin-docker](https://github.com/hashicorp/packer-plugin-docker) and it "remixes" it to
make it work with Podman. I've simply taken the default scaffolder, run sed on it (`scaffolder -> podman`) and the used
the code from the Docker plugin into this one and add a few features here and there. Therefore, the copyright of this
project is the same of [github.com/hashicorp/packer-plugin-docker](https://github.com/hashicorp/packer-plugin-docker),
and most of the work has to be given credit to HashiCorp for the hard work done with Docker that I've shamelessly taken 
in order to make this plugin.

## Packer plugin projects

Here's a non exaustive list of Packer plugins that you can checkout:

* [github.com/hashicorp/packer-plugin-docker](https://github.com/hashicorp/packer-plugin-docker)
* [github.com/exoscale/packer-plugin-exoscale](https://github.com/exoscale/packer-plugin-exoscale)
* [github.com/sylviamoss/packer-plugin-comment](https://github.com/sylviamoss/packer-plugin-comment)
* [github.com/hashicorp/packer-plugin-hashicups](https://github.com/hashicorp/packer-plugin-hashicups)

Looking at their code will give you good examples.

## Running Acceptance Tests

Make sure to install the plugin with `go build .` and to have Packer installed locally.
Then source the built binary to the plugin path with `cp packer-plugin-podman ~/.packer.d/plugins/packer-plugin-podman`
Once everything needed is set up, run:
```
PACKER_ACC=1 go test -count 1 -v ./... -timeout=120m
```

This will run the acceptance tests for all plugins in this set.

## Test Plugin Example Action

This project configures a [manually triggered plugin test action](/.github/workflows/test-plugin-example.yml).
By default, the action will run Packer at the latest version to init, validate, and build the example configuration
within the [example](example) folder. This is useful to quickly test a basic template of your plugin against Packer.

The example must contain the `required_plugins` block and require your plugin at the latest or any other released version.
This will help test and validate plugin releases.

# Requirements

-	[packer-plugin-sdk](https://github.com/hashicorp/packer-plugin-sdk) >= v0.1.0
-	[Go](https://golang.org/doc/install) >= 1.16

## Packer Compatibility
This podman template is compatible with Packer >= v1.7.0
