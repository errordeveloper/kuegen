# kuegen

`kuegen` is a simple config generator based on [CUE](https://cuelang.org).

`kuegen` is able to evaluate a CUE module as a template with a set of parameters
and will write output manifests for each set of parameters as JSON or YAML.
Each out manifest can be written to a single file as a list a list of resources,
or to multiple files with a single resource in each file.

`kuegen` establishes a very simple convention. It expects a CUE module to have
the following keys:

- `template` – contains data that will be evaluated with the given parameters
- `parameters` – contains a schema of the parameters (including default value)
- `instances` – contain one or more sets of parameters and output path

`kuegen` assumes that all objects to be handled are presented as Kubernetes-style
objects. This does allow for arbitrary schema to be used inside an object, e.g.
a nested YAML config in a `ConfigMap` can be templated with CUE, but it will need
to be a `ConfigMap` or another Kubernetes envelope and not a standalone file of
an arbitrary format.

## Example 1

Here is a simple module that defines a template for one namespace resource:

_`main.cue`_:
```
package example1

#ExampleTemplate :: {
	apiVersion: "v1"
	kind:       "Namespace"
	metadata: {
		name: parameters.name
		labels: name: parameters.name
	}
}

#ExampleParameters :: {
	name: string
}

parameters: #ExampleParameters
template:   #ExampleTemplate

instances: [
	{
		parameters: {
			name: "bar"
		}
		output: "bar.json"
	},
	{
		parameters: {
			name: "baz"
		}
		output: "baz.json"
	},
]
```

In this case everything is in a single file (`main.cue`), but multiple files can be
used for larger modules.

This module be evaluated by running:

```
$ kuegen -input-directory ./examples/1
2021/10/20 14:31:02 writing bar.json
2021/10/20 14:31:02 writing baz.json
$ cat bar.json
{
  "apiVersion": "v1",
  "kind": "Namespace",
  "metadata": {
    "labels": {
      "name": "bar"
    },
    "name": "bar"
  }
}
$ cat baz.json
{
  "apiVersion": "v1",
  "kind": "Namespace",
  "metadata": {
    "labels": {
      "name": "baz"
    },
    "name": "baz"
  }
}
$
```

## Example 2

Instances can be specified using JSON syntax also, which is convenient in some cases.
The JSON filename has to be `instances.json` and it must be located in the module's directory.

_`template.cue`_:
```
package example2

#ExampleTemplate :: {
	apiVersion: "v1"
	kind:       "Namespace"
	metadata: {
		name: parameters.name
		labels: name: parameters.name
	}
}

#ExampleParameters :: {
	name: string
}

parameters: #ExampleParameters
template:   #ExampleTemplate
```

_`instances.json`_:

```
{
    "instances": [
        {
            "parameters": {
                "name": "bar"
            },
            "output": "bar.json"
        },
        {
            "parameters": {
                "name": "baz"
            },
            "output": "baz.json"
        }
    ]
}
```

In this example the output will be the same as in _Example 1_.

## Example 3

Output manifest can be written to sepatare files when:
- value of `output` contains `%s`; AND
- `template` is a list (i.e. `kind: List`)

_`instances.json`_:
```
{
    "instances": [
        {
            "parameters": {
                "name": "foo"
            },
            "output": "%s.yaml"
        },
        {
            "parameters": {
                "name": "foo"
            },
            "output": "%s.json"
        }
    ]
}
```

Evaluating this will result in `00000-foo-namespace.yaml` and `00000-foo-namespace.json` being written.
The first digits in the filenames are based on the index of the object in the list, and are included in
order to preserve the order. Second segment is the name of the object and the third is its kind, if the
object is namespaced, name will be prefixed by that namespace.
More specifically, `%s` is substituted with either `<index>-<name>-<kind>` or `<index>-<namespace>-<name>-<kind>`.
Suffix given in the name of the output file will be used to determine encoding format.

If `-output-directory` flag is specified, the manifests will be written there instead of the current working
directory. The output path may contain directory names, which will be treated as relative to the directory that
was set with the `-output-directory` flag.
