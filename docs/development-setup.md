# Flowpipe Development Setup

## Local Development

1. Clone both `flowpipe`, `flowpipe-sdk-go` and `terraform-components` in the following directory structure:
    ```
    workspace
       |
       |-- flowpipe
       |
       |-- flowpipe-sdk-go
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