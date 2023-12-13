
 <img width="67%" src="https://flowpipe-io-git-main-turbot.vercel.app/images/flowpipe_wordmark.svg">


<p>

[![mods](https://img.shields.io/badge/mods-47-blue)](https://hub-flowpipe-io-git-main-turbot.vercel.app/mods) &nbsp; 
[![pipelines](https://img.shields.io/badge/pipelines-532-blue)](https://hub-flowpipe-io-git-main-turbot.vercel.app/mods) &nbsp;
[![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-blue)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

[Flowpipe](https://flowpipe-io.vercel.app) is a cloud scripting engine. It provides automation and workflow to connect your clouds
to the people, systems and data that matter.

**Connect people and tools**. Connect your cloud data to people and systems using email, chat & APIs. Workflow steps can even run containers, custom functions, and more.

**Orchestrate your cloud**. Build simple steps into complex workflows. Run and test locally. Compose solutions across clouds using open source mods. 

**Respond to events**. Run workflows manually or on a schedule. Trigger pipelines from webhooks or changes in data.

**Code, not clicks**. Build and deploy DevOps workflows like infrastructure. Code in HCL and deploy from version control.

## Flowpipe in action

<img width="524" src="https://steampipe.io/images/steampipe-sql-demo.gif" />

## Getting Started

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

_Note: Flowpipe container and function steps require Docker to be running. (The CLI itself does not.)_

Now, **[create and run your first pipeline →](https://flowpipe.io/docs)**

## Libraries and samples

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

Check out [Flowpipe samples](https://hub.flowpipe.io/?type=samples) for ready-to-run samples that use various library mods.

## Developing

<details>
<summary>Developing Flowpipe</summary>

### VS Code Dev Container

1. If you are using VS Code, follow the Developer Setup instructions below. Using the Dev Container in VS Code will ensure that you have all the required tools and dependencies installed.

1. Run the following commands to build and start the Flowpipe service:
    ```bash
    # Starts the service, reads pipeline definition from the `pipelines` directory
    $ make
    FLOWPIPE_LOG_LEVEL=DEBUG go run . service start --mod-location ./pipeline
    2023-06-06T11:53:49.835Z        DEBUG   Manager starting
    2023-06-06T11:53:49.835Z        DEBUG   ES starting
    2023-06-06T11:53:49.835Z        DEBUG   Pipeline dir    {"dir": "./pipeline"}
    2023-06-06T11:53:49.835Z        DEBUG   Loading pipelines       {"directory": "./pipeline"}
    2023-06-06T11:53:49.835Z        DEBUG   Loaded pipeline {"name": "for_loop_using_http_request_body_json", "file": "pipeline/for_loop_using_http_request_body_json.yaml"}
    2023-06-06T11:53:49.847Z        DEBUG   Loaded pipeline {"name": "series_of_for_loop_steps", "file": "pipeline/series_of_for_loop_steps.yaml"}
    2023-06-06T11:53:49.858Z        DEBUG   Loaded pipeline {"name": "simple_parallel", "file": "pipeline/simple_parallel.yaml"}
    2023-06-06T11:53:49.879Z        DEBUG   Adding middleware       {"count": "1"}
    </snip>
    ```

### Running

1. In your API tool of choice (e.g. Postman, Insomnia, etc.) send a `GET` request to the following URL to check that the API server is running:
    ```
    http://localhost:7103/api/v0/service
    ```

1. Check the available pipelines by sending a `GET` request to the following URL:
    ```
    http://localhost:7103/api/v0/pipeline
    ```

1. Now run one of the pipeline by sending a `POST` to the following URL:
    ```
    http://localhost:7103/api/v0/pipeline/series_of_for_loop_steps
    ```

1. Copy the resulting `exec_<xxx>` ID and do a `GET` to the following URL:
    ```
    http://localhost:7103/api/v0/process/exec_chvgkvmu69j2b44q3e60
    ```
</details>

<details>
<summary>Developer Setup</summary>

### Flowpipe Development Setup


1. Clone `flowpipe`, `flowpipe-sdk-go`, `pipe-fittings` and `terraform-components` in the following directory structure:
    ```
    workspace
       |
       |-- flowpipe
       |
       |-- flowpipe-sdk-go
       |
       |-- pipe-fittings
       |
       |-- terraform-components
    ```

1. There should be a devcontainer ready for use in `ghcr.io`. To pull this devcontainer while we're still in private mode you need to create a class GitHub PAT with the following scopes:
    1. `read:packages`: must have
    1. `write:packages` & `delete:packages`: optional

1. Pull the devcontainer: `docker pull ghcr.io/turbot/flowpipe-devcontainer:latest`

1. In VS Code install `devcontainer` extension.

1. Open `flowpipe` in `Dev Containers: Open Folder in Container...` option. It will automatically open in a dev container where the `flowpipe-sdk-go` directory automatically mounted in the correct file structure.

1. In the terminal, run `go run . service start`

1. Check that API server is running.
</details>

## Open Source & Contributing
This repository is published under the [AGPL 3.0](https://www.gnu.org/licenses/agpl-3.0.html) license. Please see our [code of conduct](https://github.com/turbot/.github/blob/main/CODE_OF_CONDUCT.md). Contributors must sign our [Contributor License Agreement](https://turbot.com/open-source#cla) as part of their first pull request. We look forward to collaborating with you!

[Flowpipe](https://flowpipe.io) is a product produced from this open source software, exclusively by [Turbot HQ, Inc](https://turbot.com). It is distributed under our commercial terms. Others are allowed to make their own distribution of the software, but cannot use any of the Turbot trademarks, cloud services, etc. You can learn more in our [Open Source FAQ](https://turbot.com/open-source).

## Get Involved

**[Join #flowpipe on Slack →](https://turbot.com/community/join)**

Want to help but don't know where to start? Pick up one of the `help wanted` issues:
* [Flowpipe](https://github.com/turbot/flowpipe/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22)

