<a href="https://flowpipe.io"><img width="67%" src="https://flowpipe.io/images/flowpipe_wordmark_outline.png"></a>

[![libraries](https://img.shields.io/badge/mods-76-blue)](https://hub.flowpipe.io) &nbsp;
[![pipelines](https://img.shields.io/badge/pipelines-1041-blue)](https://hub.flowpipe.io/mods) &nbsp;
[![slack](https://img.shields.io/badge/slack-2695-blue)](https://turbot.com/community/join) &nbsp;
[![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-blue)](https://turbot.com)

## Workflow for DevOps

[Flowpipe](https://flowpipe.io) enables automation and workflow to connect your clouds to the people, systems and data that matter.

**Pipelines**. A [pipeline](https://flowpipe.io/docs/flowpipe-hcl/pipeline) is a sequence of [steps](https://flowpipe.io/docs/flowpipe-hcl/pipeline) to do work.

**Steps**. A step can [make an HTTP call](https://flowpipe.io/docs/flowpipe-hcl/step/http), [gather human input](https://flowpipe.io/docs/flowpipe-hcl/step/input), [send a message](https://flowpipe.io/docs/flowpipe-hcl/step/message), [run a query](https://flowpipe.io/docs/flowpipe-hcl/step/query), or [run a pipeline](https://flowpipe.io/docs/flowpipe-hcl/step/pipeline).

**Triggers**. A [trigger](https://flowpipe.io/docs/flowpipe-hcl/trigger) runs a pipeline when an event occurs, via a [webhook](https://flowpipe.io/docs/flowpipe-hcl/trigger/http), [query](https://flowpipe.io/docs/flowpipe-hcl/trigger/query), or [schedule](https://flowpipe.io/docs/flowpipe-hcl/trigger/schedule).

**Code, not clicks**. Our pipelines are [code](https://flowpipe.io/docs/build): version-controlled, composable, shareable, easy to edit — designed for the way you work.

## Demo time!

**[Watch on YouTube →](https://www.youtube.com/watch?v=h4mWhMzaS7Y)**

<a href="https://www.youtube.com/watch?v=h4mWhMzaS7Y"><img width="500" alt="flowpipe demo" src="https://flowpipe.io/images/flowpipe_hero_video_thumbnail.png" /></a>

## Documentation

See the [documentation](https://flowpipe.io/docs) for:

- [Running Flowpipe](https://flowpipe.io/docs/run)
- [CLI commands](https://flowpipe.io/docs/reference/cli)
- [HCL reference](https://flowpipe.io/docs/flowpipe-hcl)
- [Configuration](https://flowpipe.io/docs/reference/config-files)
- [Building mods](https://flowpipe.io/docs/build)


## Install Flowpipe

Install Flowpipe from the [downloads](https://flowpipe.io/downloads) page:

```sh
# MacOS
brew install turbot/tap/flowpipe
```

```sh
# Linux or Windows (WSL)
sudo /bin/sh -c "$(curl -fsSL https://flowpipe.io/install/flowpipe.sh)"
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

## Developing

If you want to help develop the Flowpipe binary, these are the steps to build it.

<details>
<summary>Clone</summary>

Clone [github.com/flowpipe](https://github.com/turbot/flowpipe) and [github.com/turbot/pipe-fittings](https://github.com/turbot/pipe-fittings).

```sh
git clone git@github.com:turbot/flowpipe
git clone git@github.com:turbot/pipe-fittings
```
</details>

<details>
<summary>Build</summary>

```sh
cd flowpipe
make
```

The Flowpipe binary lands in the current directory.

</details>

<details>
<summary>Check the install</summary>

```sh
./flowpipe --version

./flowpipe --help
```
</details>

<details>
<summary>Try it!</summary>


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
</details>

<details>
<summary>DevContainer</summary>

There are other third party tools that are required for the full suite that are not required for initial development tasks. We have built a [DevContainer](https://containers.dev/) that has all the required tools installed.

1. Install [Docker](https://docs.docker.com/get-docker/)

1. Install [VS Code](https://code.visualstudio.com/)

1. Pull the Dev Container: `docker pull ghcr.io/turbot/flowpipe-devcontainer:latest`

1. In VS Code install `devcontainer` extension.

1. Open `flowpipe` in `Dev Containers: Open Folder in Container...` option.

1. Run `make` to build the Flowpipe binary.

[Flowpipe DevContainer](https://github.com/turbot/flowpipe/pkgs/container/flowpipe-devcontainer) bundles the following:

* [Java](https://openjdk.org/)
* [Apache Maven](https://maven.apache.org/)
* [Swag](https://github.com/swaggo/swag)
* [MailHog](https://github.com/mailhog/MailHog)
* [OpenAPI Generator](https://github.com/OpenAPITools/openapi-generator)

</details>


If you're interested in developing [Flowpipe mods](https://hub.flowpipe.io), see our [documentation for mod developers](https://flowpipe.io/docs/build).

## Turbot Pipes

Bring your team to [Turbot Pipes](https://turbot.com/pipes) to use Flowpipe together in the cloud.

## Open source and contributing
This repository is published under the [AGPL 3.0](https://www.gnu.org/licenses/agpl-3.0.html) license. Please see our [code of conduct](https://github.com/turbot/.github/blob/main/CODE_OF_CONDUCT.md). Contributors must sign our [Contributor License Agreement](https://turbot.com/open-source#cla) as part of their first pull request. We look forward to collaborating with you!

[Flowpipe](https://flowpipe.io) is a product produced from this open source software, exclusively by [Turbot HQ, Inc](https://turbot.com). It is distributed under our commercial terms. Others are allowed to make their own distribution of the software, but cannot use any of the Turbot trademarks, cloud services, etc. You can learn more in our [Open Source FAQ](https://turbot.com/open-source).


## Get involved

**[Join #flowpipe on Slack →](https://turbot.com/community/join)**


