
<img width="67%" src="https://flowpipe-io-git-main-turbot.vercel.app/images/flowpipe_wordmark.svg">


<p>

[![mods](https://img.shields.io/badge/mods-23-blue)](https://hub-flowpipe-io-git-main-turbot.vercel.app/mods) &nbsp; 
[![pipelines](https://img.shields.io/badge/pipelines-154-blue)](https://hub-flowpipe-io-git-main-turbot.vercel.app/mods) &nbsp;
[![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-blue)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

[Flowpipe](https://flowpipe-io.vercel.app) is the universal cloud scripting engine. Use HCL to write pipelines that conduct workflow among people, applications, services, and data sources. 

With Flowpipe you can:

- Compose a pipeline, in HCL, with steps that use Flowpipe libraries, Steampipe queries, Lambda-compatible functions, and containers — all in a consistent way.
- Run your pipeline as a CLI command, or on a trigger driven by cron, query, or webhook.
- Connect people using libraries for email, Slack, and Teams.
- Deploy pipelines as single binary to your local machine, cloud VMs, container clusters, CI/CD pipelines.
- Commit your pipelines to repos just like all your other code.




 

## Flowpipe in action

<img width="524" src="https://steampipe.io/images/steampipe-sql-demo.gif" />



## Getting Started

### Binary

1. Download the approriate binary for your platform from the [Releases](https://github.com/turbot/flowpipe/releases)

1. Create a working directory and extract the binary
    ```
    $ mkdir flowpipe
    $ cd flowpipe
    $ cp ~/Downloads/flowpipe_0.0.1_darwin_amd64.tar.gz .
    $ tar -xzf flowpipe_0.0.1_darwin_amd64.tar.gz

    # Create the output directory (default to ./tmp)
    $ mkdir tmp
    ```

1. Run flowpipe specifying the pipeline directory
    ```
    $ ./flowpipe service start --mod-location ./pipeline
    ```

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

## Running

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

## Developer Setup

[Developer Setup](./docs/development-setup.md)




