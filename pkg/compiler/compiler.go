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
	instance := cue.Merge(cue.Build(load.Instances([]string{c.inputDirectory}, nil))...)
	if instance.Err != nil {
		return nil, instance.Err
	}
	return instance, nil
}
