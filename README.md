<a href="https://flowpipe.io"><img width="67%" src="https://flowpipe.io/images/flowpipe_wordmark.svg"></a>

<p>

[![libraries](https://img.shields.io/badge/mods-62-blue)](https://hub.flowpipe.io) &nbsp;
[![pipelines](https://img.shields.io/badge/pipelines-510-blue)](https://hub.flowpipe.io/mods) &nbsp;
[![slack](https://img.shields.io/badge/slack-2297-blue)](https://turbot.com/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp;
[![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-blue)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

[Flowpipe](https://flowpipe.io) is **Workflow for DevOps**. Use it automate cloud operations, coordinate people and pipelines, and build workflows as code.

**Pipelines**. A [pipeline](https://flowpipe.io/docs/flowpipe-hcl/pipeline) is a sequence of [steps](https://flowpipe.io/docs/flowpipe-hcl/pipeline) to do work. 

**Steps**. A step can [make an HTTP call](https://flowpipe.io/docs/flowpipe-hcl/step/http), [gather human input](https://flowpipe.io/docs/flowpipe-hcl/step/input), [send a message](https://flowpipe.io/docs/flowpipe-hcl/step/message), [run a query](https://flowpipe.io/docs/flowpipe-hcl/step/query), or [run a pipeline](https://flowpipe.io/docs/flowpipe-hcl/step/pipeline).

**Triggers**. A [trigger](https://flowpipe.io/docs/flowpipe-hcl/trigger) runs a pipeline when an event occurs, via a [webhook](https://flowpipe.io/docs/flowpipe-hcl/trigger/http), [query](https://flowpipe.io/docs/flowpipe-hcl/trigger/query), or [schedule](https://flowpipe.io/docs/flowpipe-hcl/trigger/schedule).

**Code, not clicks**. Our pipelines are [code](https://flowpipe.io/docs/build): version-controlled, composable, shareable, easy to edit — designed for the way you work. 

## Demo time!

**[Watch on YouTube →](https://www.youtube.com/watch?v=h4mWhMzaS7Y)**

[![Flowpipe Demo](https://flowpipe.io/images/flowpipe_hero_video_thumbnail.png)](https://www.youtube.com/watch?v=h4mWhMzaS7Y)

## Install Flowpipe

The <a href="https://flowpipe.io/downloads?utm_id=gfpreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">downloads</a> page shows you how but tl;dr:

Linux or WSL

```sh
sudo /bin/sh -c "$(curl -fsSL https://flowpipe.io/install/flowpipe.sh)"
```

MacOS

```sh
brew tap turbot/tap
brew install flowpipe
```

Now, [create and run your first pipeline →](https://flowpipe.io/docs).

## Flowpipe mods: libraries and samples

Flowpipe [library mods](https://hub.flowpipe.io/?type=library) are available for services including
  <a href="https://hub.flowpipe.io/mods/turbot/aws">AWS</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/azure">Azure</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/gcp">GCP</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/github">GitHub</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/jira">Jira</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/okta">Okta</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/pagerduty">PagerDuty</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/sendgrid">SendGrid</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/slack">Slack</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/teams">Teams</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/zendesk">Zendesk</a> ... and many more!

Check out [Flowpipe samples](https://hub.flowpipe.io/?type=sample) for ready-to-run samples that use various library mods.


## Open source and contributing
This repository is published under the [AGPL 3.0](https://www.gnu.org/licenses/agpl-3.0.html) license. Please see our [code of conduct](https://github.com/turbot/.github/blob/main/CODE_OF_CONDUCT.md). Contributors must sign our [Contributor License Agreement](https://turbot.com/open-source#cla) as part of their first pull request. We look forward to collaborating with you!

[Flowpipe](https://flowpipe.io) is a product produced from this open source software, exclusively by [Turbot HQ, Inc](https://turbot.com). It is distributed under our commercial terms. Others are allowed to make their own distribution of the software, but cannot use any of the Turbot trademarks, cloud services, etc. You can learn more in our [Open Source FAQ](https://turbot.com/open-source).

## Developing

<details>
<summary>Developing Flowpipe</summary>

Prerequisites:

- [Golang](https://golang.org/doc/install) Version 1.21 or higher.

Clone `github.com/flowpipe` and `github.com/turbot/pipe-fittings` repositories:

```sh
git clone git@github.com:turbot/flowpipe
git clone git@github.com:turbot/pipe-fittings
cd flowpipe
```

Build will build flowpipe binary in the current directory:

```sh
make
```

Check the version:
```sh
./flowpipe --version
```
```
Flowpipe v0.0.1-local.1
```

Flowpipe local version will always be `v0.0.1-local.1`. The real version is generated during the release process.

Try it!

```sh
./flowpipe pipeline list --mod-location ./internal/es/estest/test_suite_mod/
```
```
MOD                   NAME                                                                                                        DESCRIPTION
mod.mod_depend_a      mod_depend_a.pipeline.echo_one_depend_a
mod.test_suite_mod    test_suite_mod.pipeline.any_param
mod.test_suite_mod    test_suite_mod.pipeline.bad_email_with_expr
mod.test_suite_mod    test_suite_mod.pipeline.bad_http_ignored                                                                    Ignored bad HTTP step.
mod.test_suite_mod    test_suite_mod.pipeline.bad_http_not_ignored                                                                Pipeline with a HTTP step that will fail. Error is not ignored.
</snip>
```

Now run a simple pipeline:
```sh
./flowpipe pipeline run --mod-location ./internal/es/estest/test_suite_mod/ simple
```
```
[flowpipe] Execution ID: exec_clsm62ko47mjp5f74730
[simple] Starting pipeline
[simple.echo] Starting transform
[simple.echo] Output echo_1 = echo 1
[simple.echo] Output echo_2 = echo 2
[simple.echo] Complete 2ms
[simple] Output val = Hello World
[simple] Complete 12ms exec_clsm62ko47mjp5f74730
```

That's it! You're ready to start developing.

There are other third party tools that are required for the full development suite. These are not required for basic development. To make development easy, we have built a DevContainer that has all the required tools installed. See the Developer Setup section for more details.

</details>

<details>
<summary>Developer Setup</summary>

1. Install [Docker](https://docs.docker.com/get-docker/)

1. Install [VS Code](https://code.visualstudio.com/)

1. Pull the Dev Container: `docker pull ghcr.io/turbot/flowpipe-devcontainer:latest`

1. In VS Code install `devcontainer` extension.

1. Open `flowpipe` in `Dev Containers: Open Folder in Container...` option.

1. Run `make` to build the Flowpipe binary.

</details>

## Get involved

**[Join #flowpipe on Slack →](https://turbot.com/community/join)**

Want to help but don't know where to start? Pick up one of the `help wanted` issues:
* [Flowpipe](https://github.com/turbot/flowpipe/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22)

