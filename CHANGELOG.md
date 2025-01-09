# Flowpipe

## v1.1.0 [2024-12-23]

_What's new?_

* Improved CLI load time for environments with many connection resources.
* Updated Go to v1.23.

_Bug fixes_

## v1.0.2 [2024-10-29]

_Bug fixes_

* Event jsonl output file deletion is now handled correctly. ([#960](https://github.com/turbot/flowpipe/issues/960)).
* `trigger run` command now exits when the execution is paused. ([#962](https://github.com/turbot/flowpipe/issues/962)).

## v1.0.1 [2024-10-25]

_Bug fixes_

* Fix crashing cases when using `--output json`. ([#594](https://github.com/turbot/pipe-fittings/issues/594)).
* Coerce variables set in interactive console to their declared type. ([#595](https://github.com/turbot/pipe-fittings/issues/595)).
* Nested pipelines now correctly pauses parent pipelines. ([#955](https://github.com/turbot/flowpipe/issues/955)).
* Pipeline with `max_concurrency` setting is now automatically paused and will successfully resume. ([#957](https://github.com/turbot/flowpipe/issues/957)).
* `form_url` is now sanitized.


## v1.0.0 [2024-10-22]

_What's new?_

* `connection` resource to manage credentials. [Documentation](https://flowpipe.io/docs/reference/config-files/connection).
* `connection` and `notifier` types for variables and params. ([#871](https://github.com/turbot/flowpipe/issues/871))
* `enum` validation for [variables](https://flowpipe.io/docs/flowpipe-hcl/variable#variable-types) and [params](https://flowpipe.io/docs/flowpipe-hcl/pipeline#arguments-1).
* Defined exit codes for various CLI operations. [Documentation](https://flowpipe.io/docs/reference/cli#exit-codes).

_Bug fixes_

* Passing pipeline references to nested mods for execution. ([#908](https://github.com/turbot/flowpipe/issues/908))
* Do not crash if pipeline reference is set to a string. ([#911](https://github.com/turbot/flowpipe/issues/911))

_Deprecation_

* `credential` and `credential_import` are deprecated to be replaced with `connection` and `connection_import`.

## v0.9.1 [2024-09-09]

_Bug fixes_

* `trigger` introspection output correctly shows `param` attribute. ([#900](https://github.com/turbot/flowpipe/issues/900))

## v0.9.0 [2024-09-04]

_What's new?_

* `tags` attribute in `pipeline param` and `mod variable` resources. ([#898](https://github.com/turbot/flowpipe/issues/898)).
* Updated `Docker` dependency to v27.1.2.

## v0.8.1 [2024-08-30]

_Bug fixes_

* `source` attribute in function step is now evaluated relative to the its mod directory rather than the root mod directory. ([#895](https://github.com/turbot/flowpipe/issues/895)).

## v0.8.0 [2024-08-26]

_What's new?_

* `trigger list` command includes triggers from root mod's immediate dependencies. ([#892](https://github.com/turbot/flowpipe/issues/892)).

_Bug fixes_

* Function step will no longer randomly fail in slower host machines. ([#888](https://github.com/turbot/flowpipe/issues/888)).
* Mod variable definition now matches Powerpipe's definition. ([#889](https://github.com/turbot/flowpipe/issues/889)).

## v0.7.1 [2024-08-14]

_Bug fixes_

* Complex nested map data types in `pipeline param` no longer fails with a `mismatched types` error. ([#879](https://github.com/turbot/flowpipe/issues/879)).

## v0.7.0 [2024-08-14]

_What's new?_

* On-demand trigger execution. ([#864](https://github.com/turbot/flowpipe/issues/864)).
* `param` support for trigger. ([#840](https://github.com/turbot/flowpipe/issues/840)).

_Bug fixes_

* Complex data type in `pipeline param` no longer fails with a `mismatched types` error. ([#879](https://github.com/turbot/flowpipe/issues/879)).
* Pipeline `param` default value is not nested in a `map` data type. ([#880](https://github.com/turbot/flowpipe/issues/880)).

## v0.6.1 [2024-08-05]

_Bug fixes_

* The `variable` command no longer fails if the `.flowpipe` directory in the user's home directory is not created yet. ([#872](https://github.com/turbot/flowpipe/issues/872)).

## v0.6.0 [2024-07-24]

_What's new?_

* Interactive workflows in the terminal via console integration. [Blog](https://flowpipe.io/blog/interactive-workflows-pipelines).
* Simplified progress output for `flowpipe pipeline run` command when running in [Client](https://flowpipe.io/docs/run#operating-modes) mode and not using the `--verbose` arg.
* `--data-dir` parameter to specify the location of the event store database. ([#852](https://github.com/turbot/flowpipe/issues/852)).
* `--execution-id` parameter to specify custom execution id for pipeline run. ([#856](https://github.com/turbot/flowpipe/issues/856)).
* Update `Go` version to v1.22.4.

_Bug fixes_

* Return a non-zero exit code if there's a failure. ([#855](https://github.com/turbot/flowpipe/issues/855)).
* `loop` block now respect the `if` step attribute. ([#858](https://github.com/turbot/flowpipe/issues/858)).

## v0.5.0 [2024-06-02]

_What's new?_

* Add support for installing mods from a branch or from the local file system. ([#849](https://github.com/turbot/flowpipe/issues/849)).

    To install from a branch:
    ```
    flowpipe mod install github.com/turbot/flowpipe-mod-aws-thrifty#main
    ```
    To reference a mod in the local file system:
    ```
    flowpipe mod install ../mods/local_mod_folder
    ```

- Add `--pull` flag to `mod` command to control the mod update strategy. ([#849](https://github.com/turbot/flowpipe/issues/849)). Possible update strategies are:

    - `full` - check branch and tags for both latest and accuracy
    - `latest` - update everything to latest, but only branches - not tags - are commit checked (which is the same as latest)
    - `development` - update branches and broken constraints to latest, leave satisfied constraints unchanged
    - `minimal` - only update broken constraints, do not check branches for new commits

* Variable list and show commands. ([#373](https://github.com/turbot/flowpipe/issues/373))

_Bug fixes_

* Pipeline references declared in subsequent files are correctly identified and processed.
* Preserves pipeline params ordering as specified in the pipeline definition. ([#408](https://github.com/turbot/pipe-fittings/issues/408))

## v0.4.6 [2024-05-14]

_Bug fixes_

* Load `locals` in order of dependency. ([#399](https://github.com/turbot/pipe-fittings/issues/399)).

## v0.4.5 [2024-05-10]

_Bug fixes_

* Pipeline execution no longer stalls when concurrency limit is applied and if clause returns false. ([#836](https://github.com/turbot/flowpipe/issues/836)).
* Trigger's common attributes (title, description, tags, documentation) allow functions and expresions. ([#394](https://github.com/turbot/pipe-fittings/issues/394)).

## v0.4.4 [2024-04-23]

_Bug fixes_

* Param can be used in query step's args attribute. ([#830](https://github.com/turbot/flowpipe/issues/830)).
* File watcher now correctly detect changes in the `loop` block. ([#808](https://github.com/turbot/flowpipe/issues/808)).
* Duplicate step names are now detected and reported as an error. ([#820](https://github.com/turbot/flowpipe/issues/820)).
* Better error message for invalid notifier reference. ([#826](https://github.com/turbot/flowpipe/issues/826)).


## v0.4.3 [2024-04-01]

_Bug fixes_

* Lazy create `flowpipe.db`. ([#808](https://github.com/turbot/flowpipe/issues/808)).
* Respect `max_concurrency` in `pipeline` and `input` steps. ([#815](https://github.com/turbot/flowpipe/issues/815)).
* Misleading error message for invalid step dependencies. ([#816](https://github.com/turbot/flowpipe/issues/816)).
* HTTP integration address is shown correctly at the beginning of each input step loop. ([#818](https://github.com/turbot/flowpipe/issues/818)).

## v0.4.2 [2024-03-26]

_Bug fixes_

* `loop` block now works in `container`, `function`, `message` and `input` steps.
* Use HCL expressions in `max_concurrency` step argument. ([#800](https://github.com/turbot/flowpipe/issues/800)).
* `throw`, `retry` and `error` block now works for `input` step.

## v0.4.1 [2024-03-19]

_Bug fixes_

* Input step respects the `max_concurrency` argument. ([#798](https://github.com/turbot/flowpipe/issues/798)).
* Erroneous error message detecting a missing credential where there isn't one.
* HCL `try()` function should be evaluated at runtime rather than parse time.
* Integration and input step URLs should use the provided custom host & port. ([#792](https://github.com/turbot/flowpipe/issues/792)).
* Shows filename and line number for invalid step references.

## v0.4.0 [2024-03-14]

_What's new?_

* Microsoft Teams integration. [Documentation](https://flowpipe.io/docs/reference/config-files/integration/msteams).

_Bug fixes_

* Step output attribute should be called `response` not `result`. ([#789](https://github.com/turbot/flowpipe/issues/789))
* Pipeline execution should not fail when a string argument is passed with double quotes. ([#791](https://github.com/turbot/flowpipe/issues/791))

## v0.3.2 [2024-03-11]

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
