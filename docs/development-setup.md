# Flowpipe Development Setup

## Local Development

1. Clone both `flowpipe` and `flowpipe-sdk-go` in the following directory structure:
    ```
    workspace
       |
       |-- flowpipe
       |
       |-- flowpipe-sdk-go
    ```

1. In `flowpipe\devcontainer` run the Makefile. It will build the local dev container.

1. In VS Code install `devcontainer` extension.

1. Open `flowpipe` in `Dev Containers: Open Folder in Container...` option. It will automatically open in a dev container where the `flowpipe-sdk-go` directory automatically mounted in the correct file structure.

1. In the terminal, run `go run . service start`

1. Check that API server is running.