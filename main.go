package main

import (
	"fmt"
	"os"
	"packer-plugin-podman/builder/podman"
	podmanVersion "packer-plugin-podman/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("podman", new(podman.Builder))
	pps.SetVersion(podmanVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
