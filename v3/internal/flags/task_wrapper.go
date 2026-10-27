package flags

type Build struct {
	Common
	Tags       string `name:"tags" description:"Additional build tags to pass to the Go compiler (comma-separated)"`
	Obfuscated bool   `name:"obfuscated" description:"Build with garble and stable obfuscated binding IDs"`
	Dev        bool   `name:"dev" description:"Build a development binary (no production tag, no inlining, no strip or trimpath) and publish it to the output directory"`
	GarbleArgs string `name:"garble-args" description:"Additional arguments to pass to garble before the build command"`
	Profile    string `name:"profile" description:"Deprecated: specify the profile as the first build argument"`
	Targets    string `name:"targets" description:"Comma-separated platform/architecture targets (for example darwin/arm64,linux/amd64)"`
	Formats    string `name:"formats" description:"Comma-separated distribution formats for an anonymous build request"`
	Force      bool   `name:"force" description:"Ignore Wake cache entries"`
	Plan       bool   `name:"plan" description:"Validate and print the resolved build plan without changing files"`
	JSON       bool   `name:"json" description:"Emit plan output as JSON (requires --plan)"`
	Server     bool   `name:"server" description:"Build a server-mode binary (Go tag server, no GUI) named <binary>-server"`
}

type Dev struct {
	Common
}

type Package struct {
	Common
	Profile string `name:"profile" description:"Named profile to package"`
	Targets string `name:"targets" description:"Comma-separated platform/architecture targets"`
	Formats string `name:"formats" description:"Comma-separated package formats"`
	Force   bool   `name:"force" description:"Ignore Wake cache entries"`
}

type SignWrapper struct {
	Common
	Profile string `name:"profile" description:"Named profile to sign"`
	Targets string `name:"targets" description:"Comma-separated platform/architecture targets"`
	Formats string `name:"formats" description:"Comma-separated package formats to build and sign"`
}
