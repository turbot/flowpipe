# Flowpipe

## v0.2.0 [tbd]

_What's new?_

* Query trigger. [See more in our documentation](https://flowpipe.io/docs/).
* Query step now supports:
    - MySQL
    - SQLite

_Enhancements_

* Improved output when running in `server` mode.
* Container step now supports `Source` in addition to `Image`. [See more in our documentation](https://flowpipe.io/docs/).

_Bug fixes_

* Implemented a more descriptive error message for server startup failures.
* Step arguments are now able to be referenced in the pipeline definition.
* Added missing `execution_mode` argument to HTTP Trigger. [#533](https://github.com/turbot/flowpipe/issues/533).

## v0.1.0 [2023-12-13]

Introducing Flowpipe, a cloud scripting engine. Automation and workflow to connect your clouds to the people, systems and data that matter. Pipelines for DevOps written in HCL.

Initial support for:
* Pipeline execution
* Steps: container, email, function, http, pipeline, query, sleep, transform
* Triggers: schedule, http
* Credential management
* Mod composition

Learn more at:
* Website - https://flowpipe.io
* Docs - https://flowpipe.io/docs
* Hub - https://hub.flowpipe.io
* Introduction - https://flowpipe.io/blog/introducing-flowpipe
