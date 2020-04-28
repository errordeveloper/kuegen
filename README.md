# kue


[CUE](https://cuelang.org) offers a very fluent syntax for composing configuration data, it also provide validation.

Using CUE is a little challenging, especially because it's entirely up to user how they want to pick the data and
whether/how it would be written to files.

`kue` provides as simple tool for generating configuration with CUE template by introducting a few simple conventions.
The aim is to focus primarily on Kubernetes and GitOps use-cases, but it is also an just an experiment right now.
