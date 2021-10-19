# kue


[CUE](https://cuelang.org) offers a very fluent syntax for composing configuration data, it also has many great features
enabling extensive validation of configuration data as part template definition.

Using CUE can be a little challenging at first, especially because it's entirely up to the user how they want to consume
the results and whether any data would be written to files.

`kue` provides as simple tool for generating configuration with CUE template by introducting a few simple conventions.
The aim is to focus primarily on Kubernetes and GitOps use-cases, but it is just an experiment right now.
