package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"cuelang.org/go/cue"
)

func doCompileAndValidate(r *cue.Runtime, filename string) (*cue.Instance, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// TODO: pass reader instead of data

	instance, err := r.Compile(filename, data)
	if err != nil {
		return nil, err
	}

	// TODO: validate the convetional schema

	return instance, nil

}

func logJSON(k interface{}, v cue.Value) {
	js, err := v.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s: %s\n", k, string(js))
}

func main() {
	var r cue.Runtime

	template, err := doCompileAndValidate(&r, "template.cue")
	if err != nil {
		log.Fatal(err)
	}

	instances, err := doCompileAndValidate(&r, "instances.cue")
	if err != nil {
		log.Fatal(err)
	}

	instancesObjects, err := instances.Lookup("instances").List()
	if err != nil {
		log.Fatal(err)
	}

	handlers := []func() error{}

	for instancesObjects.Next() {
		output := instancesObjects.Value().Lookup("output")
		parameters := instancesObjects.Value().Lookup("parameters")
		if err != nil {
			log.Fatal(err)
		}
		handlers = append(handlers, func() error {
			result, err := template.Fill(parameters, "parameters")
			if err != nil {
				return err
			}
			logJSON(output, result.Lookup("template"))

			return nil
		})
	}

	for _, call := range handlers {
		err := call()
		if err != nil {
			log.Fatal(err)
		}
	}
}
