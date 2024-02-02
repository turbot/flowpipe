# Flowpipe

## v0.2.2 [2024-02-02]

_Bug fixes_

* Build error no longer suppressed in container and function steps ([#625](https://github.com/turbot/flowpipe/issues/625)).
* Handles complex data types in step output ([#626](https://github.com/turbot/flowpipe/issues/626)).

## v0.2.1 [2024-01-29]

_Bug fixes_

* Map MySQL query results to correct types ([#604](https://github.com/turbot/flowpipe/issues/604)).
* Handle null values in query trigger results ([#611](https://github.com/turbot/flowpipe/issues/611)).
* Convert binary data in query results to a string.
* Docker containers now clear the cache to get correct parameters ([#561](https://github.com/turbot/flowpipe/issues/561)).
* Improved error message when Flowpipe CLI port is already in use ([#603](https://github.com/turbot/flowpipe/issues/603)).

## v0.2.0 [2024-01-24]

_What's new?_

* Query trigger type to watch & event on to database changes. [Documentation](https://flowpipe.io/docs/flowpipe-hcl/trigger/query).
* HTTP trigger can now handle both GET and POST methods. [Documentation](https://flowpipe.io/docs/flowpipe-hcl/trigger/http).
* Query steps & triggers now support Postgres, MySQL, SQLite and Postgres.
* Define container step using a `source` argument for inline image definitions.
* Add a `timeout` to pipeline steps.
* Enable or disable triggers using `enabled` attribute.
* Improved and expanded output for `flowpipe server`.
* Improved and standardized output for CLI `list` and `show` commands.
* Expanded intervals available in schedule and query triggers (e.g. `5m`, `10m`, etc).
* New credential types: BitBucket, Datadog, Freshdesk, JumpCloud, ServiceNow, Turbot Guardrails.

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
