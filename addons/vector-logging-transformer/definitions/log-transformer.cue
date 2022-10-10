// We put Definitions in definitions directory.
// References:
// - https://kubevela.net/docs/platform-engineers/cue/definition-edit
// - https://kubevela.net/docs/platform-engineers/addon/intro#definitions-directoryoptional
"log-transformer": {
	alias: "lt"
	annotations: {}
	attributes: {
		appliesToWorkloads: [
			"deployments.apps",
			"replicasets.apps",
			"statefulsets.apps",
		]
		conflictsWith: []
		podDisruptive:   false
		workloadRefPath: ""
	}
	description: "ETL transformer for application log"
	labels: {}
	type: "trait"
}
template: {

	parameter: {
		type: "nginx" | "apache"
		smapleRate: *100 | int
	}

  sourceName: context.appName+"_"+context.name+"_source"
  parserName: context.appName+"_"+context.name+"_parser"
  samplerName: context.appName+"_"+context.name+"_sampler"
  sinkName: context.appName+"_"+context.name+"_sink"


	outputs: vector_config: {
		apiVersion: "vector.oam.dev/v1alpha1"
		kind:       "Config"
		metadata: {
			name: context.name
		}
		spec: {
			role: "daemon"
			targetConfigMap: {
				namespace: "vector",
				name: "vector"
			}
			vectorConfig: {
				sources: "\(sourceName)": {
					type: "kubernetes_logs"
					extra_label_selector: "app.oam.dev/name="+ context.appName + ",app.oam.dev/component=" + context.name
				}
				transforms: {
					"\(parserName)": {
						inputs: [sourceName]
						type: "remap"
						if parameter.type == "apache" {
							source: """
              .message = parse_apache_log!(.message, format: "common")
              """
						}
						if parameter.type == "nginx" {
							source: """
              .message = parse_nginx_log!(.message, "combined")
              """
            }
				  }

					"\(samplerName)": {
						inputs: [parserName]
						type: "sampler"
						rate: parameter.smapleRate
					}
				}
				sinks:
				  "\(sinkName)": {
					  type: "loki"
					  inputs: [samplerName]
					  endpoint: "$LOKIURL"
					  compression: "none"
					  request:
					    concurrency:10
					  labels: {
								agent: "vector"
								stream:  "{{ stream }}"
								forward: "daemon"
								filename: "{{ file }}"
								namespace: "{{ kubernetes.pod_namespace }}"
							  pod: "{{ kubernetes.pod_name }}"
              }
            encoding: codec: "json"
				}
			}
		}
	}
}
