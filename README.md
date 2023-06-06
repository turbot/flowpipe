# Flowpipe

## Getting Started

1. If you are using VS Code, follow the [Developer Setup](#developer-setup) instructions below. Using the Dev Container in VS Code will ensure that you have all the required tools and dependencies installed.

1. Run the following commands to build and start the Flowpipe service:
    ```bash
    # Starts the service using debug logging
    $ make
    ```

1. In your API tool of choice (e.g. Postman, Insomnia, etc.) send a `GET` request to the following URL to check that the API server is running:
    ```
    https://localhost:7103/api/v0/service
    ``` 

1. Check the available pipelines by sending a `GET` request to the following URL:
    ```
    https://localhost:7103/api/v0/pipeline
    ```

1. Now run one of the pipeline by sending a `POST` to the following URL:
    ```
    https://localhost:7103/api/v0/pipeline/series_of_for_loop_steps
    ```

1. Copy the resulting `exec_<xxx>` ID and do a `GET` to the following URL:
    ```
    https://localhost:7103/api/v0/process/exec_chvgkvmu69j2b44q3e60
    ```

## Developer Setup

[Developer Setup](./docs/development-setup.md)




