# Flowpipe

## Developer Setup

([Developer Setup](./docs/development-setup.md))

## Communication

Clients, including the flowpipe CLI, communicate with the flowpipe server via
a REST API. Requests may be sent to any node, but the node will forward the
request to the leader of the pipeline's raft group.

Communication between nodes, including for raft coordination, uses gRPC with
a protobuf message format.

(Given multiple raft groups, I'm not sure how the communication for those will
work - I hope it can fit into the single gRPC server model as well.)

It will be common for client requests to be received via the HTTPS REST API
and then result in gRPC commands within the cluster. The REST API is for client
communication, the gRPC API is for cluster communication.


## Consensus

A flowpipe cluster is a collection of nodes.

Nodes are discovered by the cluster (e.g. via DNS) and form a raft group. This
raft group is used to maintain cluster wide information only.

The basic unit of work in flowpipe is a pipeline. Pipelines are executed on
nodes. Pipelines can be arranged into groups. Each node may choose which pipeline
groups it can execute.

Each pipeline group is assigned a new raft group. This raft group is used to
manage pipeline execution and ensure high availability. Only the leader of the
raft group can execute a pipeline, followers are used as backup if the leader
fails.

There is a raft group to coordinate pipelines and results. It includes only the
leader of each pipeline group.







## Server

Start the service:
```
flowpipe service start
```

Get service status:
```
flowpipe service status
```

Add a new node to an existing cluster:
```
flowpipe service start --server https://localhost:7103
```

If the node has existed previously and you want to rejoin the cluster, just
provide the node ID:
```
flowpipe service start --server https://localhost:7103 --node fpn_a1b2c3d4e5f6
```

During development it can be helpful to run multiple node processes on the
same machine. To avoid conflicts, you can specify a port:
```
flowpipe service start
  --server https://localhost:7103 \
  --raft-address 8102 \
  --https-address 8103
```

Remove a node from a cluster:
```
flowpipe service remove-node --server https://localhost:7103 --node fpn_a1b2c3d4e5f6
```


## Development

Flowpipe pipelines can be run on demand from the command line without needing a
server:
```
flowpipe run my_pipeline
```

You can also run a pipeline on a server from the command line:
```
flowpipe run --server https://localhost:7103 my_pipeline
```
