package example1

#ExampleTemplate :: {
	apiVersion: "v1"
	kind:       "Namespace"
	metadata: {
		name: "\(parameters.name)"
		labels: name: "\(parameters.name)"
	}
}

#ExampleParameters :: {
	name:      string
}

parameters: #ExampleParameters
template:   #ExampleTemplate
