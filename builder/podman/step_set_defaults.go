package podman

import (
	"context"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"strings"
)

type StepSetDefaults struct{}

func (s *StepSetDefaults) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(*PodmanDriver)
	config := state.Get("config").(*Config)

	// Fetch default CMD and ENTRYPOINT
	defaultCmd, _ := driver.Cmd(config.Image)
	defaultEntrypoint, _ := driver.Entrypoint(config.Image)

	// Set defaults if not provided by the user
	hasCmd, hasEntrypoint := false, false
	for _, change := range config.Changes {
		if strings.HasPrefix(change, "CMD") {
			hasCmd = true
		} else if strings.HasPrefix(change, "ENTRYPOINT") {
			hasEntrypoint = true
		}
	}

	if !hasCmd && defaultCmd != "" {
		config.Changes = append(config.Changes, "CMD "+defaultCmd)
	}
	if !hasEntrypoint && defaultEntrypoint != "" {
		config.Changes = append(config.Changes, "ENTRYPOINT "+defaultEntrypoint)
	}

	return multistep.ActionContinue
}

func (s *StepSetDefaults) Cleanup(state multistep.StateBag) {}
