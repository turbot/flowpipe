# flowpipe

Flowpipe is "pipelines as code", defining workflows and other tasks
that are performed in a sequence.

## How it works

A mod defines a collection of pipelines and triggers. The mod may be
loaded to then start listening for events and executing pipelines.

## Event Sourcing and CQRS

Here is the sequence when starting:
* Queue - waiting to load the mod
* Load - Load the full mod definition including pipelines and triggers
* Start - Start execution of the mod
* Plan - Plan the next tasks to be executed for the mod

The mod can end in a few ways:
* Finish - Everything is done, clean shutdown
* Fail - Something went wrong, forced shutdown

When running a pipeline (regardless of the trigger), the sequence is:
* Queue - waiting to start the pipeline
* Load - Load the pipeline definition
* Start - Start execution of the pipeline
* Plan - Plan the next steps to be executed for the pipeline
* Queue Step - Queue the step to be executed
* Load Step - Load the step definition
* Start Step - Start execution of the step
* Execute Step - Execute the step
* Finish Step - Finish execution of the step
* Fail Step - Fail execution of the step due to an error


## Runtime identifiers

The mod is running, waiting for triggers.
Each trigger starts a pipeline, which has a unique ID.
Each step in the pipeline has a unique ID.

The IDs above are nested, giving a StackID.