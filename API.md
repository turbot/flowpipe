# Flowpipe API

All interaction with Flowpipe is done via the API. Even the CLI connects to the
API for it's commands and operations.

## Definition

GET  /v0/service

LIST /v0/pipeline
GET  /v0/pipeline/{pipeline_name}

LIST /v0/trigger
GET  /v0/trigger/{trigger_name}

LIST   /v0/pipeline_execution
POST   /v0/pipeline_execution
GET    /v0/pipeline_execution/{pexec_id}
PATCH  /v0/pipeline_execution/{pexec_id}
DELETE /v0/pipeline_execution/{pexec_id}

POST   /v0/hooks/{trigger_name}                // generic form
POST   /v0/hooks/my_webhook                    // example - short name of local mod
POST   /v0/hooks/local.trigger.http.my_webhook // example - fully qualified
POST   /v0/hooks/local.my_webhook              // example - partially qualified
POST   /v0/hooks/custom_name_for_my_webhook    // example - custom name

## CLI commands

List running pipelines:
```
$ flowpipe pipeline list
ID          PIPELINE               STATUS       DURATION
abcd1234    local.my_pipeline      running      4 secs
bdcd5413    local.my_pipeline      queued       1 secs
```

Run a pipeline:
```
$ flowpipe pipeline run local.mypipeline --input '{"your": "json"}'

Pipeline:  local.my_pipeline
Execution: abcd1234
Status:    queued

To watch:

  flowpipe pipeline watch abcd1234

To inspect:

  flowpipe pipeline inspect abcd1234

To pause:

  flowpipe pipeline pause abcd1324

To cancel:

  flowpipe pipeline cancel abcd1324

```

Pause an execution:
```
$ flowpipe pipeline pause abcd1234

Pipeline:  local.my_pipeline
Execution: abcd1234
Status:    paused

To resume, run:

  flowpipe pipeline resume abcd1324

To cancel, run:

  flowpipe pipeline cancel abcd1324

```

Cancel an execution:
```
$ flowpipe pipeline cancel abcd1234
...
```

List triggers:
```
$ flowpipe trigger list
TRIGGER                STATUS       EVENTS    FAILURES
local.my_webhook       enabled      55        0
local.my_cron          enabled      12        3
```

Disable a trigger:
```
$ flowpipe trigger disable local.my_webhook

Trigger:  local.my_webhook
Status:   disabled
Events:   55
Failures: 0

```

Enable a trigger:
```
$ flowpipe trigger enable local.my_webhook

Trigger:  local.my_webhook
Status:   enabled
Events:   55
Failures: 0

```





