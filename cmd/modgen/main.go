package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"github.com/errordeveloper/kue/pkg/compiler"
)

const (
	templateFilename  = "template.cue"
	instancesFilename = "instances.cue"

	templateKey  = "template"
	instancesKey = "instances"

	parametersKey = "parameters"
	outputKey     = "output"
)

type generator struct {
	inputDirectory, outputDirectory string

	template  *cue.Instance
	instances []*templateInstance
}

type templateInstance struct {
	output     string
	parameters cue.Value
}

func (g *generator) CompileAndValidate() error {
	// TODO: produce meanigful compilation errors
	// TODO: produce meanigful validation errors
	// TODO: validate the types match expections, e.g. objets not string

	c := compiler.NewCompiler(g.inputDirectory)

	template, err := c.Compile(templateFilename)
	if err != nil {
		return err
	}

	g.template = template

	instances, err := c.Compile(instancesFilename)
	if err != nil {
		return err
	}

	instancesIterator, err := instances.Lookup(instancesKey).List()
	if err != nil {
		return err
	}

	for instancesIterator.Next() {
		// TODO: try using decode method instead
		output, err := instancesIterator.Value().Lookup(outputKey).String()
		if err != nil {
			return err
		}

		g.instances = append(g.instances, &templateInstance{
			output:     output,
			parameters: instancesIterator.Value().Lookup(parametersKey),
		})
	}

	return nil
}

func (g *generator) WriteFiles() error {
	for _, ti := range g.instances {
		result, err := g.template.Fill(ti.parameters, parametersKey)
		if err != nil {
			return err
		}

		data, err := result.Lookup(templateKey).MarshalJSON()
		if err != nil {
			return err
		}

		// TODO: determine mode based on umask?
		log.Printf("writing %s\n", ti.output)
		if err := os.MkdirAll(filepath.Dir(ti.output), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(ti.output, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	inputDirectory := flag.String("input-directory", ".", "input directory to read module definition from")
	outputDirectory := flag.String("output-directory", ".", "out directory to write generated manifest to")

	flag.Parse()

	// TODO: add example that imports kubernetes types
	g := &generator{
		inputDirectory:  *inputDirectory,
		outputDirectory: *outputDirectory,
	}

	if err := g.CompileAndValidate(); err != nil {
		log.Fatal(err)
	}

	if err := g.WriteFiles(); err != nil {
		log.Fatal(err)
	}
}
