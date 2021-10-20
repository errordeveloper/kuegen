package example2

#ExampleTemplate :: {
	kind:       "List"
	apiVersion: "v1"
	items: [
		{
			apiVersion: "v1"
			kind:       "Namespace"
			metadata: {
				name: parameters.name
				labels: name: parameters.name
			}
		},
	]
}

#ExampleParameters :: {
	name: string
}

parameters: #ExampleParameters
template:   #ExampleTemplate
