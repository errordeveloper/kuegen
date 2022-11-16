package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/errordeveloper/cue-utils/compiler"
)

const (
	instancesFilenameJSON = "instances.json"

	templateKey  = "template"
	instancesKey = "instances"

	parametersKey = "parameters"
	outputKey     = "output"
)

type generator struct {
	inputDirectory, outputDirectory string

	template  cue.Value
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

	g.compiler = compiler.NewCompiler()

	template, err := g.compiler.BuildAll(g.inputDirectory, ".")
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

	instancesKeyPath := cue.ParsePath(instancesKey)
	if err := instancesKeyPath.Err(); err != nil {
		return err
	}
	instancesIterator, err := g.template.LookupPath(instancesKeyPath).List()
	if err != nil {
		return err
	}

	outputKeyPath := cue.ParsePath(outputKey)
	if err := outputKeyPath.Err(); err != nil {
		return err
	}
	parametersKeyPath := cue.ParsePath(parametersKey)
	if err := parametersKeyPath.Err(); err != nil {
		return err
	}

	for instancesIterator.Next() {
		// TODO: try using decode method instead
		output, err := instancesIterator.Value().LookupPath(outputKeyPath).String()
		if err != nil {
			return err
		}

		parameters := instancesIterator.Value().LookupPath(parametersKeyPath)
		g.instances = append(g.instances, &templateInstance{
			output:     output,
			parameters: &parameters,
		})
	}

	return nil
}

func (g *generator) useInstancesJSON() bool {
	return g.fileExists(instancesFilenameJSON)
}

func (g *generator) fileExists(filename string) bool {
	info, err := os.Stat(filepath.Join(g.inputDirectory, filename))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (g *generator) WriteFiles() error {
	for _, ti := range g.instances {
		var parameters interface{}
		if ti.parameters != nil {
			parameters = ti.parameters
		} else {
			parameters = ti.parametersFromJSON
		}

		parametersKeyPath := cue.ParsePath(parametersKey)
		if err := parametersKeyPath.Err(); err != nil {
			return err
		}
		result := g.template.FillPath(parametersKeyPath, parameters)
		if err := result.Err(); err != nil {
			return err
		}

		templateKeyPath := cue.ParsePath(templateKey)
		if err := templateKeyPath.Err(); err != nil {
			return err
		}
		data, err := result.LookupPath(templateKeyPath).MarshalJSON()
		if err != nil {
			return err
		}

		obj := &unstructured.Unstructured{}

		useYAML := strings.HasSuffix(ti.output, ".yml") || strings.HasSuffix(ti.output, ".yaml")
		splitFiles := strings.Contains(ti.output, "%s")

		// TODO: can we get a map from cue? from a first look,
		// cue's MarshalJSON has a few special(?) internal methods
		if err := obj.UnmarshalJSON(data); err != nil {
			return fmt.Errorf("cannot convert JSON data to unstructured Kubernetes object: %w", err)
		}

		if !splitFiles {
			filename := filepath.Join(g.outputDirectory, ti.output)
			if err := writeFile(useYAML, obj.Object, filename); err != nil {
				return err
			}
			continue
		}

		if !obj.IsList() {
			return fmt.Errorf("cannot split files as object is not a list")
		}

		list, err := obj.ToList()
		if err != nil {
			return fmt.Errorf("cannot convert object to a list")
		}

		for index, item := range list.Items {
			derivedName := fmt.Sprintf("%05d-%s-%s", index, item.GetName(), strings.ToLower(item.GetKind()))

			filename := filepath.Join(g.outputDirectory, fmt.Sprintf(ti.output, derivedName))
			if err := writeFile(useYAML, item.Object, filename); err != nil {
				return err
			}
		}

	}
	return nil
}

func writeFile(useYAML bool, obj map[string]interface{}, filename string) error {
	var (
		data []byte
		err  error
	)

	if useYAML {
		data, err = yaml.Marshal(obj)
		if err != nil {
			return fmt.Errorf("gerating YAML: %w", err)
		}
	} else {
		data, err = json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return fmt.Errorf("gerating pretty JSON: %w", err)
		}
	}

	// TODO: determine mode based on umask?
	log.Printf("writing %s\n", filename)
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func main() {
	inputDirectory := flag.String("input-directory", ".", "input directory to read module definition from")
	outputDirectory := flag.String("output-directory", ".", "output directory to write generated manifest to")

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
