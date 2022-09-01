package main

vector: {
	type: "helm"
	name: "vector"
	properties: {
		repoType: "helm"
		url:      "https://helm.vector.dev"
		chart:    "ivector-agent"
		version:  "0.21.3"
		values: {
			 existingConfigMap: "vector-config"
		}
	}
}
