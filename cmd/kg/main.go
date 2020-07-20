package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"

	"github.com/errordeveloper/kue/pkg/compiler"
)

const (
	templateFilename      = "template.cue"
	instancesFilename     = "instances.cue"
	instancesFilenameJSON = "instances.json"

	templateKey  = "template"
	instancesKey = "instances"

	parametersKey = "parameters"
	outputKey     = "output"
)

type generator struct {
	inputDirectory, outputDirectory string

	template  *cue.Instance
	instances []*templateInstance

	compiler *compiler.Compiler
}

type templateInstance struct {
	output             string
	parameters         *cue.Value
	parametersFromJSON map[string]interface{}
}

type InstancesFromJSON struct {
	Instances []struct {
		Parameters map[string]interface{} `json:"parameters"`
		Output     string                 `json:"output"`
	} `json:"instances"`
}

func (g *generator) CompileAndValidate() error {
	// TODO: produce meanigful compilation errors
	// TODO: produce meanigful validation errors
	// TODO: validate the types match expections, e.g. objets not string

	g.compiler = compiler.NewCompiler(g.inputDirectory)

	template, err := g.compiler.Compile(templateFilename)
	if err != nil {
		return err
	}

	g.template = template

	if g.useInstancesJSON() {
		filename := filepath.Join(g.inputDirectory, instancesFilenameJSON)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		obj := &InstancesFromJSON{}

		if err := json.Unmarshal(data, obj); err != nil {
			return fmt.Errorf("parsing %q: %w", filename, err)
		}

		for _, instance := range obj.Instances {
			g.instances = append(g.instances, &templateInstance{
				output:             instance.Output,
				parametersFromJSON: instance.Parameters,
			})

		}

		return nil
	}

	instances, err := g.compiler.Compile(instancesFilename)
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

		parameters := instancesIterator.Value().Lookup(parametersKey)
		g.instances = append(g.instances, &templateInstance{
			output:     output,
			parameters: &parameters,
		})
	}

	return nil
}

func (g *generator) useInstancesJSON() bool {
	return !g.fileExists(instancesFilename) && g.fileExists(instancesFilenameJSON)
}

func (g *generator) fileExists(filename string) bool {
	info, err := os.Stat(filepath.Join(g.inputDirectory, filename))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (g *generator) WriteFiles(prettyJSON bool) error {
	for _, ti := range g.instances {
		var parameters interface{}
		if ti.parameters != nil {
			parameters = ti.parameters
		} else {
			parameters = ti.parametersFromJSON
		}
		result, err := g.template.Fill(parameters, parametersKey)
		if err != nil {
			return err
		}

		data, err := result.Lookup(templateKey).MarshalJSON()
		if err != nil {
			return err
		}

		if prettyJSON {
			// TODO: can we get a map from cue? from a first look,
			// cue's MarshalJSON has a few special(?) internal methods
			temp := map[string]interface{}{}
			err := json.Unmarshal(data, &temp)
			if err != nil {
				return fmt.Errorf("gerating pretty JSON: %w", err)
			}
			data, err = json.MarshalIndent(temp, "", "  ")
			if err != nil {
				return fmt.Errorf("gerating pretty JSON: %w", err)
			}
		}
		// TODO: determine mode based on umask?
		filename := filepath.Join(g.outputDirectory, ti.output)
		log.Printf("writing %s\n", filename)
		if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	inputDirectory := flag.String("input-directory", ".", "input directory to read module definition from")
	outputDirectory := flag.String("output-directory", ".", "output directory to write generated manifest to")
	prettyJSON := flag.Bool("pretty-json", true, "write pretty JSON manifest")

	flag.Parse()

	// TODO: add example that imports kubernetes types
	g := &generator{
		inputDirectory:  *inputDirectory,
		outputDirectory: *outputDirectory,
	}

	if err := g.CompileAndValidate(); err != nil {
		log.Fatal(err)
	}

	if err := g.WriteFiles(*prettyJSON); err != nil {
		log.Fatal(err)
	}
}
