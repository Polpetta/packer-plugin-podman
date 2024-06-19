module packer-plugin-podman

go 1.16

require (
	github.com/hashicorp/go-version v1.7.0
	github.com/hashicorp/hcl/v2 v2.20.1
	github.com/hashicorp/packer-plugin-sdk v0.5.4
	github.com/mitchellh/mapstructure v1.5.0
	github.com/zclconf/go-cty v1.14.2
)

replace github.com/zclconf/go-cty => github.com/nywilken/go-cty v1.13.3 // added by packer-sdc fix as noted in github.com/hashicorp/packer-plugin-sdk/issues/187
