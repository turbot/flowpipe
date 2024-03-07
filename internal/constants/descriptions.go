package constants

const (
	// FlowpipeShortDescription is a short description of the application used in the CLI.
	FlowpipeShortDescription = "Pipelines and workflows for DevSecOps"

	// FlowpipeLongDescription is a long description of the application used in the CLI.
	FlowpipeLongDescription = `Flowpipe: Workflow for DevOps

Automate cloud operations. Coordinate people and pipelines.
Build workflows as code.

Common commands:

  # Install a mod from the hub - https://hub.flowpipe.io
  flowpipe mod init
  flowpipe mod install github.com/turbot/flowpipe-mod-{example}

  # Start the server for triggers and interactive workflows
  flowpipe server

  # Run a pipeline while developing
  flowpipe pipeline list
  flowpipe pipeline run {pipeline_name}

Documentation:
  https://flowpipe.io/docs`
)
