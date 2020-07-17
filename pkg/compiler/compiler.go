package compiler

import (
	"bufio"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
)

type Compiler struct {
	inputDirectory string

	runtime *cue.Runtime
}

func NewCompiler(inputDirectory string) *Compiler {
	return &Compiler{
		inputDirectory: inputDirectory,
		runtime:        &cue.Runtime{},
	}
}

func (c *Compiler) Compile(filename string) (*cue.Instance, error) {
	filePath := filepath.Join(c.inputDirectory, filename)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	instance, err := c.runtime.Compile(filePath, bufio.NewReader(file))
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (c *Compiler) BuildAll() (*cue.Instance, error) {
	loadedInstances := load.Instances([]string{c.inputDirectory}, nil)
	for _, loadedInstance := range loadedInstances {
		if loadedInstance.Err != nil {
			return nil, loadedInstance.Err
		}
	}

	builtInstances := cue.Build(loadedInstances)
	for _, builtInstance := range builtInstances {
		if builtInstance.Err != nil {
			return nil, builtInstance.Err
		}
		if err := builtInstance.Value().Validate(); err != nil {
			return nil, err
		}
	}

	mergedInstance := cue.Merge(builtInstances...)
	if mergedInstance.Err != nil {
		return nil, mergedInstance.Err
	}
	return mergedInstance, nil
}
