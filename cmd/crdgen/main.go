package main

import (
	"flag"
	"log"

	"cuelang.org/go/cue"
	"github.com/errordeveloper/kue/cmd/crdgen/types"
	"github.com/errordeveloper/kue/pkg/compiler"
)

const (
	templateFilename = "template.cue"

	templateKey = "template"
	resourceKey = "resource"
)

type generator struct {
	inputDirectory string

	template *cue.Instance
}

func (g *generator) CompileAndValidate() error {
	// TODO: produce meanigful compilation errors
	// TODO: produce meanigful validation errors
	// TODO: validate the types match expections, e.g. objets not string

	c := compiler.NewCompiler(g.inputDirectory)

	template, err := c.BuildAll()
	if err != nil {
		return err
	}

	g.template = template

	return nil
}

func (g *generator) WriteOutput() error {

	cluster := types.Cluster{}
	cluster.Metadata.Name = "foo1"
	cluster.Metadata.Namespace = "defuault"
	cluster.Spec.Location = "us-central1-a"

	result, err := g.template.Fill(cluster, resourceKey)
	if err != nil {
		return err
	}

	data, err := result.Lookup(templateKey).MarshalJSON()
	if err != nil {
		return err
	}

	log.Println(string(data))

	return nil
}

func main() {

	flag.Parse()

	g := &generator{
		inputDirectory: ".",
	}

	if err := g.CompileAndValidate(); err != nil {
		log.Fatal(err)
	}

	if err := g.WriteOutput(); err != nil {
		log.Fatal(err)
	}
}
