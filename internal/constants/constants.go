package constants

const (
	// Name is the name of the application.
	Name = "flowpipe"

	// ShortDescription is a short description of the application used in the CLI.
	ShortDescription = "Pipelines and workflows for DevSecOps"

	// LongDescription is a long description of the application used in the CLI.
	LongDescription = "Flowpipe is a distributed, fault-tolerant, and highly-available workflow engine. See https://flowpipe.io for more information."
)

func ApplicationName() string {
	return Name
}
