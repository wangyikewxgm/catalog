package main

import "encoding/yaml"

vectorDeamonsetConfig: {
	type: "k8s-objects"
	name: "vector-config"
	properties: {
		objects: [{
			apiVersion: "v1"
			kind:       "ConfigMap"
			metadata: {
				name: "vector-config"
			}
			data: yaml.Marshal(_vectorDeamonsetConfig)
		}]
	}
}

_vectorDeamonsetConfig: {
	data_dir: "/vector-data-dir"
	api: {
		enabled:    true
		address:    "127.0.0.1:8686"
		playground: false
	}
	sources: pod_stdout_log: {
		type:                 "kubernetes_logs"
		extra_label_selector: "log-collector==vector"
	}
	sinks: loki: {
		type: "loki"
		inputs: [
			"pod_stdout_log",
		]
		endpoint:    parameter.lokiEndpoint
		compression: "none"
		request: concurrency: 10
		labels: {
			log_type:      "stdout"
			pod_namespace: "{{ kubernetes.pod_namespace }}"
			pod_name:      "{{ kubernetes.pod_name }}"
		}
		encoding: codec: "json"
	}
}
