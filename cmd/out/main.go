package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-deployment-resource/bosh"
	"github.com/cloudfoundry/bosh-deployment-resource/concourse"
	"github.com/cloudfoundry/bosh-deployment-resource/out"
	"io/ioutil"
)

func main() {
	fakeStdout, _ := ioutil.TempFile("", "stdout")
	realStdout := os.Stdout
	os.Stdout = fakeStdout

	fakeStderr, _ := ioutil.TempFile("", "stderr")
	os.Stderr = fakeStderr

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr,
			"not enough args - usage: %s <sources directory>\n",
			os.Args[0],
		)
		os.Exit(1)
	}

	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read configuration: %s\n", err)
		os.Exit(1)
	}

	outRequest, err := concourse.NewOutRequest(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid parameters: %s\n", err)
		os.Exit(1)
	}

	sourcesDir := os.Args[1]

	commandRunner := bosh.NewCommandRunner(outRequest.Source, os.Stderr)
	director := bosh.NewBoshDirector(outRequest.Source, commandRunner, sourcesDir, os.Stderr)

	outCommand := out.NewOutCommand(director, sourcesDir)
	outResponse, err := outCommand.Run(outRequest)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	concourseOutputFormatted, err := json.MarshalIndent(outResponse, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not generate version: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "%s", concourseOutputFormatted)
	fmt.Printf("%s", concourseOutputFormatted)

	fakeStdout.Close()
	os.Stdout = realStdout
	outContents, _ := ioutil.ReadFile(realStdout.Name())
	fmt.Print(string(outContents))

	fakeStderr.Close()
}
