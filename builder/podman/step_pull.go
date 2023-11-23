package podman

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepPull struct{}

func (s *StepPull) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config, ok := state.Get("config").(*Config)
	if !ok {
		err := fmt.Errorf("error encountered obtaining podman config")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if !config.Pull {
		log.Println("Pull disabled, won't podman pull")
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Pulling Podman image: %s", config.Image))

	driver := state.Get("driver").(Driver)

	if config.Login {
		ui.Message("Logging in...")
		err := driver.Login(
			config.LoginServer,
			config.LoginUsername,
			config.LoginPassword)
		if err != nil {
			err := fmt.Errorf("Error logging in: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}


	if err := driver.Pull(config.Image); err != nil {
		err := fmt.Errorf("Error pulling Podman image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepPull) Cleanup(state multistep.StateBag) {
}
