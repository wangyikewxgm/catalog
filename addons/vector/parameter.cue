// parameter.cue is used to store addon parameters.
//
// You can use these parameters in template.cue or in resources/ by 'parameter.cue'
//
// For example, you can use parameters to allow the user to customize
// container images, ports, and etc.
parameter: {
	// +usage=Custom parameter description
	lokiEndpoint?: *"http://loki:3100" | string
	// +usage=The clusters to install
  clusters?: [...string]
}
