package podman

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type PodmanDriver struct {
	Ui  packersdk.Ui
	Ctx *interpolate.Context

	l sync.Mutex
}

func (d *PodmanDriver) DeleteImage(id string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("podman", "rmi", id)
	cmd.Stderr = &stderr

	log.Printf("Deleting image: %s", id)
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error deleting image: %s\nStderr: %s",
			err, stderr.String())
		return err
	}

	return nil
}

func (d *PodmanDriver) Commit(id string, author string, changes []string, message string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := []string{"commit"}
	if author != "" {
		args = append(args, "--author", author)
	}
	for _, change := range changes {
		args = append(args, "--change", change)
	}
	if message != "" {
		args = append(args, "--message", message)
	}
	args = append(args, id)

	log.Printf("Committing container with args: %v", args)
	cmd := exec.Command("podman", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error committing container: %s\nStderr: %s",
			err, stderr.String())
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *PodmanDriver) Export(id string, dst io.Writer) error {
	var stderr bytes.Buffer
	cmd := exec.Command("podman", "export", id)
	cmd.Stdout = dst
	cmd.Stderr = &stderr

	log.Printf("Exporting container: %s", id)
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error exporting: %s\nStderr: %s",
			err, stderr.String())
		return err
	}

	return nil
}

func (d *PodmanDriver) Import(path string, changes []string, repo string) (string, error) {
	var stdout, stderr bytes.Buffer

	args := []string{"import"}

	for _, change := range changes {
		args = append(args, "--change", change)
	}

	args = append(args, "-")
	args = append(args, repo)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()

	if err != nil {
		return "", err
	}

	// There should be only one artifact of the Podman builder
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	log.Printf("Importing tarball with args: %v", args)

	if err := cmd.Start(); err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		//nolint
		io.Copy(stdin, file)
	}()

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("Error importing container: %s\n\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *PodmanDriver) IPAddress(id string) (string, error) {
	var stderr, stdout bytes.Buffer
	cmd := exec.Command(
		"podman",
		"inspect",
		"--format",
		"{{ .NetworkSettings.IPAddress }}",
		id)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Error: %s\n\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *PodmanDriver) Sha256(id string) (string, error) {
	var stderr, stdout bytes.Buffer
	cmd := exec.Command(
		"podman",
		"inspect",
		"--format",
		"{{ .Id }}",
		id)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Error: %s\n\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *PodmanDriver) Cmd(id string) (string, error) {
	var stderr, stdout bytes.Buffer
	cmd := exec.Command(
		"podman",
		"inspect",
		"--format",
		"{{if .Config.Cmd}} {{json .Config.Cmd}} {{else}} [] {{end}}",
		id)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Error: %s\n\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *PodmanDriver) Entrypoint(id string) (string, error) {
	var stderr, stdout bytes.Buffer
	cmd := exec.Command(
		"podman",
		"inspect",
		"--format",
		"{{if .Config.Entrypoint}} {{json .Config.Entrypoint}} {{else}} [] {{end}}",
		id)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Error: %s\n\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *PodmanDriver) Login(repo, user, pass string) error {
	d.l.Lock()

	version_running, err := d.Version()
	if err != nil {
		d.l.Unlock()
		return err
	}

	// Version 17.07.0 of Podman adds support for the new
	// `--password-stdin` option which can be used to offer
	// password via the standard input, rather than passing
	// the password and/or token using a command line switch.
	constraint, err := version.NewConstraint(">= 17.07.0")
	if err != nil {
		d.l.Unlock()
		return err
	}

	cmd := exec.Command("podman")
	cmd.Args = append(cmd.Args, "login")

	if user != "" {
		cmd.Args = append(cmd.Args, "-u", user)
	}

	if pass != "" {
		if constraint.Check(version_running) {
			cmd.Args = append(cmd.Args, "--password-stdin")

			stdin, err := cmd.StdinPipe()
			if err != nil {
				d.l.Unlock()
				return err
			}
			_, err = io.WriteString(stdin, pass)
			if err != nil {
				return err
			}
			stdin.Close()
		} else {
			cmd.Args = append(cmd.Args, "-p", pass)
		}
	}

	if repo != "" {
		cmd.Args = append(cmd.Args, repo)
	}

	err = runAndStream(cmd, d.Ui)
	if err != nil {
		d.l.Unlock()
		return err
	}

	return nil
}

func (d *PodmanDriver) Logout(repo string) error {
	args := []string{"logout"}
	if repo != "" {
		args = append(args, repo)
	}

	cmd := exec.Command("podman", args...)
	err := runAndStream(cmd, d.Ui)
	d.l.Unlock()
	return err
}

func (d *PodmanDriver) Pull(image string) error {
	cmd := exec.Command("podman", "pull", image)
	return runAndStream(cmd, d.Ui)
}

func (d *PodmanDriver) Push(name string) error {
	cmd := exec.Command("podman", "push", name)
	return runAndStream(cmd, d.Ui)
}

func (d *PodmanDriver) SaveImage(id string, dst io.Writer) error {
	var stderr bytes.Buffer
	cmd := exec.Command("podman", "save", id)
	cmd.Stdout = dst
	cmd.Stderr = &stderr

	log.Printf("Exporting image: %s", id)
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error exporting: %s\nStderr: %s",
			err, stderr.String())
		return err
	}

	return nil
}

func (d *PodmanDriver) StartContainer(config *ContainerConfig) (string, error) {
	// Build up the template data
	var tplData startContainerTemplate
	tplData.Image = config.Image
	ictx := *d.Ctx
	ictx.Data = &tplData

	// Args that we're going to pass to Podman
	args := []string{"run"}
	for _, v := range config.Device {
		args = append(args, "--device", v)
	}
	for _, v := range config.CapAdd {
		args = append(args, "--cap-add", v)
	}
	for _, v := range config.CapDrop {
		args = append(args, "--cap-drop", v)
	}
	if config.Privileged {
		args = append(args, "--privileged")
	}
	args = append(args, fmt.Sprintf("--systemd=%s", config.Systemd))
	for _, v := range config.TmpFs {
		args = append(args, "--tmpfs", v)
	}
	for host, guest := range config.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", host, guest))
	}
	for _, v := range config.RunCommand {
		v, err := interpolate.Render(v, &ictx)
		if err != nil {
			return "", err
		}

		args = append(args, v)
	}
	d.Ui.Message(fmt.Sprintf(
		"Run command: podman %s", strings.Join(args, " ")))

	// Start the container
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("podman", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Starting container with args: %v", args)
	if err := cmd.Start(); err != nil {
		return "", err
	}

	log.Println("Waiting for container to finish starting")
	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("Podman exited with a non-zero exit status.\nStderr: %s",
				stderr.String())
		}

		return "", err
	}

	// Capture the container ID, which is alone on stdout
	return strings.TrimSpace(stdout.String()), nil
}

func (d *PodmanDriver) StopContainer(id string) error {
	if err := exec.Command("podman", "stop", id).Run(); err != nil {
		return err
	}
	return nil
}

func (d *PodmanDriver) KillContainer(id string) error {
	if err := exec.Command("podman", "kill", id).Run(); err != nil {
		return err
	}

	return exec.Command("podman", "rm", id).Run()
}

func (d *PodmanDriver) TagImage(id string, repo string, force bool) error {
	args := []string{"tag"}

	// detect running podman version before tagging
	// flag `force` for podman tagging was removed after Podman 1.12.0
	// to keep its backward compatibility, we are not going to remove `force`
	// option, but to ignore it when Podman version >= 1.12.0
	//
	// for more detail, please refer to the following links:
	// - https://docs.podman.com/engine/deprecated/#/f-flag-on-podman-tag
	// - https://github.com/podman/podman/pull/23090
	version_running, err := d.Version()
	if err != nil {
		return err
	}

	version_deprecated, err := version.NewVersion("1.12.0")
	if err != nil {
		// should never reach this line
		return err
	}

	if force {
		if version_running.LessThan(version_deprecated) {
			args = append(args, "-f")
		} else {
			// do nothing if Podman version >= 1.12.0
			log.Printf("[WARN] option: \"force\" will be ignored here")
			log.Printf("since it was removed after Podman 1.12.0 released")
		}
	}
	args = append(args, id, repo)

	var stderr bytes.Buffer
	cmd := exec.Command("podman", args...)
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error tagging image: %s\nStderr: %s",
			err, stderr.String())
		return err
	}

	return nil
}

func (d *PodmanDriver) Verify() error {
	if _, err := exec.LookPath("podman"); err != nil {
		return err
	}

	return nil
}

func (d *PodmanDriver) Version() (*version.Version, error) {
	output, err := exec.Command("podman", "-v").Output()
	if err != nil {
		return nil, err
	}

	match := regexp.MustCompile(version.VersionRegexpRaw).FindSubmatch(output)
	if match == nil {
		return nil, fmt.Errorf("unknown version: %s", output)
	}

	return version.NewVersion(string(match[0]))
}
