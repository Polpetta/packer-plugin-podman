package podman

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepCommit commits the container to a image.
type StepCommit struct {
	imageId string
}

func (s *StepCommit) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config, ok := state.Get("config").(*Config)
	if !ok {
		err := fmt.Errorf("error encountered obtaining podman config")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	driver := state.Get("driver").(Driver)
	containerId := state.Get("container_id").(string)
	ui.Say("Committing the container")
	imageId, err := driver.Commit(containerId, config.Author, config.Changes, config.Message)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Save the container ID
	s.imageId = imageId
	state.Put("image_id", s.imageId)
	ui.Message(fmt.Sprintf("Image ID: %s", s.imageId))

	return multistep.ActionContinue
}

func (s *StepCommit) Cleanup(state multistep.StateBag) {}
