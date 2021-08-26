package main

import (
	"fmt"
	"os"
	"packer-plugin-podman/builder/podman"
	podmanData "packer-plugin-podman/datasource/podman"
	podmanPP "packer-plugin-podman/post-processor/podman"
	podmanProv "packer-plugin-podman/provisioner/podman"
	podmanVersion "packer-plugin-podman/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("my-builder", new(podman.Builder))
	pps.RegisterProvisioner("my-provisioner", new(podmanProv.Provisioner))
	pps.RegisterPostProcessor("my-post-processor", new(podmanPP.PostProcessor))
	pps.RegisterDatasource("my-datasource", new(podmanData.Datasource))
	pps.SetVersion(podmanVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
