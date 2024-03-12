# Flowpipe

## v0.4.0 [tbd]

_What's new?_

* Microsoft Teams integration. [Documentation](https://flowpipe.io/docs/reference/config-files/integration/msteams).

_Bug fixes_

* Step output attribute should be called response not result. ([#789](https://github.com/turbot/flowpipe/issues/789))

## v0.3.2 [2023-03-08]

_Bug fixes_

* Multiselect Inputs with preselected Options now correctly pre-populate in Slack.
* Change detection in `throw` and `output` block in pipeline steps works correctly with ternary operators and will not trigger mod reload for white space changes.

## v0.3.1 [2024-03-07]

_Bug fixes_

* Multi-select option in input step now works. ([#776](https://github.com/turbot/flowpipe/issues/776)).
* Input step white space changes will not trigger mod reload. ([#297](https://github.com/turbot/pipe-fittings/issues/297)).

## v0.3.0 [2024-03-05] Human workflow, Slack and email messaging, Import Steampipe credentials, Concurrency controls.

_What's new?_

* Workflow - message step for easy notifications. [Documentation](https://flowpipe.io/docs/flowpipe-hcl/step/message).
* Workflow - input step for buttons, text and other data. [Documentation](https://flowpipe.io/docs/flowpipe-hcl/step/input).
* Workflow - simple, reusable integration and notifier configuration for HTTP, Slack and Email. [Documentation](https://flowpipe.io/docs/reference/config-files/integration).
* Import Steampipe connections as Flowpipe credentials. [Documentation](https://flowpipe.io/docs/reference/config-files/credential_import).
* Manage concurrency of [pipelines](https://flowpipe.io/docs/flowpipe-hcl/pipeline#arguments) and [steps](https://flowpipe.io/docs/flowpipe-hcl/step#common-step-arguments).
* New credential types: `alicloud` and `mastodon`.
* Shorter hash for HTTP triggers for simpler URLs.
* DuckDB support in query step & trigger.
* Step metadata, like `started_at` and `finished_at` added under a `flowpipe` attribute.
* Moved `flowpipe.db` into the mod-level `.flowpipe` directory.
* `connection_string` in query step and trigger renamed to `database`.

_Deprecation_

* Email step. Please use the new message step instead.

_Bug fixes_

* `log_level` workspace setting is now respected ([#618](https://github.com/turbot/flowpipe/issues/618)).
* Default `listen` flag should be network, not localhost ([#694](https://github.com/turbot/flowpipe/issues/694)).
* Trigger attributes are now validated ([#225](https://github.com/turbot/pipe-fittings/issues/255)).
* Pipeline output attributes are now validated ([#239](https://github.com/turbot/pipe-fittings/issues/239)).
* Pipeline param default value data type is now validated against the specified type ([#262](https://github.com/turbot/pipe-fittings/issues/262)).
* Removed titles when merging multiple error messages ([#263](https://github.com/turbot/pipe-fittings/issues/263)).
* Runtime resolution of pipeline reference and credentials are now working correctly. ([#732](https://github.com/turbot/flowpipe/issues/732)).
* Scheduled triggers are now re-scheduled when mod files have changed.
* File watcher reliability improvements.

## v0.2.3 [2024-02-13]

_Bug fixes_

* Only trigger pipeline failure after a step has completed all retries ([#630](https://github.com/turbot/flowpipe/issues/630)).
* `DOCKER_HOST`, `DOCKER_API_VERSION`, `DOCKER_CERT_PATH`, `DOCKER_TLS_VERIFY` environment variables are now correctly passed to the Docker client ([#651](https://github.com/turbot/flowpipe/issues/651)).
* Do not set memory_swappiness when using Podman ([#652](https://github.com/turbot/flowpipe/issues/652)).

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
