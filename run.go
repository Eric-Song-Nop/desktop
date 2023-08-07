package desktop

import (
	"fmt"
	"io/ioutil"
	"os"
)

// RunScript creates a temporary run script consisting of a bash script which
// removes itself and executes the specified command.
func RunScript(exec string) (string, error) {
	runScript, err := ioutil.TempFile("", "run-*")
	if err != nil {
		return "", err
	}

	_, err = runScript.WriteString("#!/bin/sh\n")
	if err != nil {
		runScript.Close()
		return "", err
	}
	_, err = runScript.WriteString(fmt.Sprintf("rm %s\n", runScript.Name()))
	if err != nil {
		runScript.Close()
		return "", err
	}
	_, err = runScript.WriteString("exec " + exec + "\n")
	if err != nil {
		runScript.Close()
		return "", err
	}

	err = runScript.Close()
	if err != nil {
		return "", err
	}

	err = os.Chmod(runScript.Name(), 0744)
	if err != nil {
		return "", err
	}

	return runScript.Name(), nil
}
