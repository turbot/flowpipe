# Flowpipe

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




