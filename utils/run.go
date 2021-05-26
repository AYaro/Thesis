package utils

import (
	"bytes"
	"log"
	"os"
	"os/exec"

	"github.com/urfave/cli"
)

func RunInDir(cmd, dir string) ([]byte, error) {
	if os.Getenv("DEBUG") != "" {
		log.Println(cmd)
	}

	command := exec.Command("bash", "-c", cmd)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		return output, cli.NewExitError(err, 1)
	}
	return output, nil
}

func Run(cmd string) ([]byte, error) {
	return RunInDir(cmd, ".")
}

func RunCommand(cmd string, path string) error {
	if os.Getenv("DEBUG") != "" {
		log.Println(cmd)
	}

	command := exec.Command("bash", "-c", cmd)
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Dir = path

	err := command.Run()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	return nil
}

func RunWithInput(cmd string, input []byte) ([]byte, error) {
	command := exec.Command("bash", "-c", cmd)
	command.Stdin = bytes.NewReader(input)
	data, err := command.Output()
	if err != nil {
		return data, cli.NewExitError(err.Error(), 1)
	}
	return data, nil
}
