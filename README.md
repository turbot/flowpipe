
 <img width="67%" src="https://flowpipe-io-git-main-turbot.vercel.app/images/flowpipe_wordmark.svg">


<p>

[![mods](https://img.shields.io/badge/mods-47-blue)](https://hub-flowpipe-io-git-main-turbot.vercel.app/mods) &nbsp; 
[![pipelines](https://img.shields.io/badge/pipelines-500-blue)](https://hub-flowpipe-io-git-main-turbot.vercel.app/mods) &nbsp;
[![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-blue)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

[Flowpipe](https://flowpipe-io.vercel.app) is the universal cloud scripting engine. It provides automation and workflow to connect your clouds
to the people, systems and data that matter.

With Flowpipe you can:

**Connect people and tools** → Flowpipe connects cloud data to both people and systems. You build workflows against APIs and databases, and can even run containers and Lambda-compatible functions as pipeline steps.

**Orchestrate your cloud** → Combine simple pipeline steps into workflows with control logic and error handling. Deploy a single binary to cloud VMs, container clusters, and CI/CD pipelines.

**Respond to events** → Trigger workflows — manually, via webhooks or queries, or on a schedule — that notify people. 

**Automate with code, not clicks** → Stop emailing spreadsheets! With Flowpipe, ops workflows are another kind of IaC. Use HCL to define workflows as code, and use Git to collaborate with your team and our community.

## Flowpipe in action

<img width="524" src="https://steampipe.io/images/steampipe-sql-demo.gif" />

## Getting Started

<details>
<summary>Ensure that Docker is installed and running.</summary>
Flowpipe's container support requires Docker.
</details>

<details>
<summary>Install the Flowpipe binary.</summary>

1. Download the approriate binary for your platform from the [Releases](https://github.com/turbot/flowpipe/releases).

1. Create a working directory and extract the binary

    ```
    $ mkdir flowpipe
    $ cd flowpipe
    $ cp ~/Downloads/flowpipe_0.0.1_darwin_amd64.tar.gz .
    $ tar xzf flowpipe_darwin_amd64.tar.gz
    ```

1. Add that directory to your path, or move the `flowpipe` binary to a location (e.g. `/usr/local/bin`) that is on the path.

1. Verify the installation.

```
$ flowpipe -v
Flowpipe v0.1.0
```
</details>

<details>
<summary>Install and use a Flowpipe mod.</summary>

1. Choose a mod from [Flowpipe hub](https://hub.flowpipe.io), e.g. [flowpipe-mod-github](https://hub.flowpipe.io/mods/turbot/github).

1. On the hub page for your mod, click `Install Mod` and follow the instructions to clone the mod.

1. Run `flowpipe pipeline list` in the directory you cloned:

    ```
    $ flowpipe pipeline list
    MOD           NAME                                                DESCRIPTION
    mod.github    github.pipeline.add_issue_assignees                 Add assignees to an issue.
    mod.github    github.pipeline.close_issue                         Close an issue with the given ID.
    mod.github    github.pipeline.close_pull_request                  Closes a pull request.
    ```

    ```
    $ flowpipe pipeline run get_current_user
    [get_current_user] Output user = {
    "login": "jsmyth",
    ...
    }
    ```
</details>

<details>
<summary>Start the server and trigger a pipeline.</summary>

You only need to start the server if you're running a pipeline that responds to a webhook. In that case:

1. Create this setup in a directory.

    In `mod.fp`:

    ```hcl
    mod "local" {
        title = "trigger_test"    
    }
    ```

    In `trigger_hello.fp`:

    ```hcl
    mod "hello" {
        trigger "http" "hello_webhook" {
            pipeline = pipeline.hello
        }

        pipeline "hello" {
            output "hello" {
                value = "hello"
            }
    }
    ```

1. Run `flowpipe server`, specifying your mod location.

    ```
    $ ./flowpipe server --mod-location ~/YOUR_MOD
    ```


1. Run `flowpipe trigger list` to list your triggers.

    ```
    $ flowpipe trigger list
    PIPELINE                TYPE    NAME                                DESCRIPTION    URL
                                            SCHEDULE
    local.pipeline.hello    http    local.trigger.http.hello_webhook                   /hook/local.trigger.http.hello_webhook/f08b...b41f8
    ```

4. Use `curl` to test the webhook.

    ```
    curl /hook/local.trigger.http.hello_webhook/f08b...b41f8
    ```
</details>

## Libraries and samples

The Flowpipe hub lists and documents [library mods](https://hub.flowpipe.io/mods) for services including 
  <a href="https://hub.flowpipe.io/mods/turbot/aws">AWS</a>, 
  <a href="https://hub.flowpipe.io/mods/turbot/azure">Azure</a>, 
  <a href="https://hub.flowpipe.io/mods/turbot/gcp">GCP</a>, 
  <a href="https://hub.flowpipe.io/mods/turbot/github">GitHub</a>, 
  <a href="https://hub.flowpipe.io/mods/turbot/googleworkspace">GoogleWorkspace</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/jira">Jira</a>, 
  <a href="https://hub.flowpipe.io/mods/turbot/okta">Okta</a>, 
  <a href="https://hub.flowpipe.io/mods/turbot/pagerduty">PagerDuty</a>, 
  <a href="https://hub.flowpipe.io/mods/turbot/salesforce">Salesforce</a>, 
  <a href="https://hub.flowpipe.io/mods/turbot/sendgrid">SendGrid</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/Slack">Slack</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/teams">Teams</a>,
  <a href="https://hub.flowpipe.io/mods/turbot/zendesk">Zendesk</a> ... and many more!

The [Flowpipe samples](https://github.com/turbot/flowpipe-samples) repo contains ready-to-run samples that use various library mods.

## Community

We thrive on feedback and community involvement!

Join our [Slack community](https://turbot.com/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) or open a [GitHub issue](https://github.com/turbot/flowpipe/issues/new/choose).

## License

Flowpipe is distributed under [AGPL-3.0](https://github.com/turbot/flowpipe/blob/main/LICENSE). Flowpipe mods are distributed under [Apache-2.0](https://github.com/turbot/flowpipe-mod-github/blob/main/LICENSE).

## Contributor license agreement
To safeguard the legal integrity of our projects and facilitate their sustainable growth, we require a [Contributor License Agreement (CLA)](https://turbot.com/legal/cla-faq) for contributions to `turbot/flowpipe`, `turbot/flowpipe-docs`, and `turbot/pipe-fittings`. The `turbot/flowpipe-mod-*` repos do not require a CLA.

## Developing

**consider moving all this into docs/development-setup.md?**

<details>
<summary>Developing Flowpipe</summary>

### VS Code Dev Container

1. If you are using VS Code, follow the [Developer Setup](#developer-setup) instructions below. Using the Dev Container in VS Code will ensure that you have all the required tools and dependencies installed.

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

#### Developer Setup

[Developer Setup](./docs/development-setup.md)
</details>



