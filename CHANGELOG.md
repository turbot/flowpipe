# Flowpipe

## v0.2.0 [tbd]

_What's new?_

* Query Trigger. [See more in our documentation](https://flowpipe.io/docs/flowpipe-hcl/trigger/query).
* Added Query Step's database supports for:
    - MySQL
    - SQLite
    - DuckDB
* Added credentials support for the following plugins:
    - BitBucket
    - Datadog
    - Freshdesk
    - JumpCloud
    - ServiceNow
    - Turbot Guardrails

_Enhancements_

* Improved output when running in `server` mode.
* Improved Container and Function step build performance.
* Added `source` argument to Container Step in addition to `image`. [See more in our documentation](https://flowpipe.io/docs/flowpipe-hcl/step/container#arguments).
* Added `timeout` argument to Pipeline steps.
* Added `method` block for HTTP Trigger.
* Added `enabled` attribute to Flowpipe Triggers.
* Improved output for `list` and `show` commands.
* New intervals (`5m`, `10m`, `15m`, `30m`, `60m`, `1h`, `2h`, `4h`, `6h`, `12h`, `24h`) are now supported for the Schedule and Query Triggers.

_Bug fixes_

* Implemented a more descriptive error message for server startup failures.
* Fixed Step Arguments unable to be referenced in the Pipeline definition.
* Added missing `execution_mode` argument to HTTP Trigger ([#533](https://github.com/turbot/flowpipe/issues/533)).
* Fixed `args` arguments unable to be updated in the Pipeline Step loop block ([#559](https://github.com/turbot/flowpipe/issues/559)).
* Fixed an issue in the bootstrap process for identifying the config path.

## v0.1.1 [2024-01-09]

_Bug fixes_

* Removed inaccurate SQL Query string validation to check for arguments. ([#516](https://github.com/turbot/flowpipe/issues/516))

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
