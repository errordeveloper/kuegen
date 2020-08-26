package example2

ClusterTemplate :: {
	kind:       "List"
	apiVersion: "v1"
	items: [
		{
			apiVersion: "container.cnrm.cloud.google.com/v1beta1"
			kind:       "ContainerCluster"
			metadata: {
				namespace: "\(parameters.namespace)"
				name:      "\(parameters.name)"
				labels: cluster:                                               "\(parameters.name)"
				annotations: "cnrm.cloud.google.com/remove-default-node-pool": "false"
			}
			spec: {
				location: "us-central1-a"
				networkRef: name:    "\(parameters.name)"
				subnetworkRef: name: "\(parameters.name)"
				initialNodeCount:  1
				loggingService:    "logging.googleapis.com/kubernetes"
				monitoringService: "monitoring.googleapis.com/kubernetes"
				masterAuth: clientCertificateConfig: issueClientCertificate: false
			}
		}, {
			apiVersion: "compute.cnrm.cloud.google.com/v1beta1"
			kind:       "ComputeNetwork"
			metadata: {
				namespace: "\(parameters.namespace)"
				name:      "\(parameters.name)"
				labels: cluster: "\(parameters.name)"
			}
			spec: {
				routingMode:                 "REGIONAL"
				autoCreateSubnetworks:       false
				deleteDefaultRoutesOnCreate: false
			}
		}, {
			apiVersion: "compute.cnrm.cloud.google.com/v1beta1"
			kind:       "ComputeSubnetwork"
			metadata: {
				namespace: "\(parameters.namespace)"
				name:      "\(parameters.name)"
				labels: cluster: "\(parameters.name)"
			}
			spec: {
				ipCidrRange: "10.128.0.0/20"
				region:      "us-central1"
				networkRef: name: "\(parameters.name)"
			}
		},
	]
}

ClusterParameters :: {
	namespace: string
	name:      string
}

parameters: ClusterParameters
template:   ClusterTemplate
