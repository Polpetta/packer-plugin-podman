package podman

import (
	"context"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

const BuilderId = "podman.builder"

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	err = config.Decode(&b.config, &config.DecodeOpts{
		PluginType:  "packer.builder.podman",
		Interpolate: true,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}
	// Return the placeholder for the generated data that will become available to provisioners and post-processors.
	// If the builder doesn't generate any data, just return an empty slice of string: []string{}
	buildGeneratedData := []string{"GeneratedMockData"}
	return buildGeneratedData, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	driver := &PodmanDriver{Ctx: &b.config.ctx, Ui: ui}
	if err := driver.Verify(); err != nil {
		return nil, err
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	generatedData := &packerbuilderdata.GeneratedData{State: state}

	// Setup the driver that will talk to Podman
	state.Put("driver", driver)

	steps := []multistep.Step{
		&StepTempDir{},
		&StepPull{},
		&StepRun{},
		&communicator.StepConnect{
			Config: &b.config.Comm,
			Host: commHost(b.config.Comm.Host()),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
			CustomConnect: map[string]multistep.Step{
				"podman": &StepConnectPodman{},
			},
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
	}

	if b.config.Discard {
		log.Print("[DEBUG] Container will be discarded")
	} else if b.config.Commit {
		log.Print("[DEBUG] Container will be committed")
		steps = append(steps,
			new(StepCommit),
			&StepSetGeneratedData{ // Adds ImageSha256 variable available after StepCommit
				GeneratedData: generatedData,
			})
	} else if b.config.ExportPath != "" {
		log.Printf("[DEBUG] Container will be exported to %s", b.config.ExportPath)
		steps = append(steps, new(StepExport))
	} else {
		return nil, errArtifactNotUsed
	}

	steps = append(steps,
		nil,
		new(commonsteps.StepProvision),
	)

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	artifact := &Artifact{
		// Add the builder generated data to the artifact StateData so that post-processors
		// can access them.
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	return artifact, nil
}
