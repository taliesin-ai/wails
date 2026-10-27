package flags

type SetupDocker struct {
	Build bool `name:"build" description:"Build the wails-cross image locally from the embedded Dockerfile instead of pulling it"`
}
