{
  "schema_version": "20221222",
  "start_time": "2023-11-30T10:50:19Z",
  "end_time": "2023-11-30T10:50:19Z",
  "layout": {
    "name": "pexec_clk6hqpqot4o7a4sehs0",
    "panel_type": "dashboard",
    "children": [
      {
        "name": "execution_tree",
        "panel_type": "graph"
      }
    ]
  },
  "panels": {
    "edge_start_pexec_clk6hqpqot4o7a4sehs0_to_step_transform.foo": {
      "dashboard": "pexec_clk6hqpqot4o7a4sehs0",
      "name": "edge_start_pexec_clk6hqpqot4o7a4sehs0_to_step_transform.foo",
      "panel_type": "edge",
      "status": "complete",
      "data": {
        "columns": [
          {
            "name": "from_id",
            "data_type": "TEXT"
          },
          {
            "name": "to_id",
            "data_type": "TEXT"
          }
        ],
        "rows": [
          {
            "from_id": "start_pexec_clk6hqpqot4o7a4sehs0",
            "to_id": "step_transform.foo"
          }
        ]
      },
      "properties": {
        "name": "edge_start_pexec_clk6hqpqot4o7a4sehs0_to_step_transform.foo"
      }
    },
    "edge_step_transform.foo_to_end_pexec_clk6hqpqot4o7a4sehs0": {
      "dashboard": "pexec_clk6hqpqot4o7a4sehs0",
      "name": "edge_step_transform.foo_to_end_pexec_clk6hqpqot4o7a4sehs0",
      "panel_type": "edge",
      "status": "complete",
      "data": {
        "columns": [
          {
            "name": "from_id",
            "data_type": "TEXT"
          },
          {
            "name": "to_id",
            "data_type": "TEXT"
          }
        ],
        "rows": [
          {
            "from_id": "step_transform.foo",
            "to_id": "end_pexec_clk6hqpqot4o7a4sehs0"
          }
        ]
      },
      "properties": {
        "name": "edge_step_transform.foo_to_end_pexec_clk6hqpqot4o7a4sehs0"
      }
    },
    "end_pexec_clk6hqpqot4o7a4sehs0": {
      "dashboard": "pexec_clk6hqpqot4o7a4sehs0",
      "name": "end_pexec_clk6hqpqot4o7a4sehs0",
      "panel_type": "node",
      "status": "complete",
      "title": "End: default_mod.pipeline.pipes_echo",
      "data": {
        "columns": [
          {
            "name": "id",
            "data_type": "TEXT"
          },
          {
            "name": "title",
            "data_type": "TEXT"
          },
          {
            "name": "properties",
            "data_type": "JSONB"
          }
        ],
        "rows": [
          {
            "id": "end_pexec_clk6hqpqot4o7a4sehs0",
            "properties": {
              "Execution ID": "pexec_clk6hqpqot4o7a4sehs0",
              "Output": {
                "foo": "foo"
              },
              "Status": "finished"
            },
            "title": "End: default_mod.pipeline.pipes_echo"
          }
        ]
      },
      "properties": {
        "category": {
          "color": "green",
          "icon": "valve",
          "name": "pipeline",
          "title": "Pipeline"
        },
        "name": "end_pexec_clk6hqpqot4o7a4sehs0"
      }
    },
    "execution_tree": {
      "dashboard": "pexec_clk6hqpqot4o7a4sehs0",
      "name": "execution_tree",
      "panel_type": "graph",
      "status": "complete",
      "title": "Execution",
      "display_type": "graph",
      "data": {
        "columns": [
          {
            "name": "from_id",
            "data_type": "TEXT"
          },
          {
            "name": "to_id",
            "data_type": "TEXT"
          },
          {
            "name": "id",
            "data_type": "TEXT"
          },
          {
            "name": "title",
            "data_type": "TEXT"
          },
          {
            "name": "properties",
            "data_type": "JSONB"
          }
        ],
        "rows": [
          {
            "from_id": "start_pexec_clk6hqpqot4o7a4sehs0",
            "to_id": "step_transform.foo"
          },
          {
            "from_id": "step_transform.foo",
            "to_id": "end_pexec_clk6hqpqot4o7a4sehs0"
          },
          {
            "id": "start_pexec_clk6hqpqot4o7a4sehs0",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clk6hqpqot4o7a4sehs0",
              "Status": "finished"
            },
            "title": "Start: default_mod.pipeline.pipes_echo"
          },
          {
            "id": "end_pexec_clk6hqpqot4o7a4sehs0",
            "properties": {
              "Execution ID": "pexec_clk6hqpqot4o7a4sehs0",
              "Output": {
                "foo": "foo"
              },
              "Status": "finished"
            },
            "title": "End: default_mod.pipeline.pipes_echo"
          },
          {
            "id": "step_transform.foo",
            "properties": {
              "Execution ID": "sexec_clk6hqpqot4o7a4seht0",
              "Input": {
                "value": "foo"
              },
              "Output": {
                "status": "finished",
                "data": {
                  "value": "foo"
                }
              },
              "Status": "finished"
            },
            "title": "0 = "
          }
        ]
      },
      "properties": {
        "direction": "TD",
        "edges": [
          "edge_start_pexec_clk6hqpqot4o7a4sehs0_to_step_transform.foo",
          "edge_step_transform.foo_to_end_pexec_clk6hqpqot4o7a4sehs0"
        ],
        "name": "execution",
        "nodes": [
          "start_pexec_clk6hqpqot4o7a4sehs0",
          "end_pexec_clk6hqpqot4o7a4sehs0",
          "step_transform.foo"
        ]
      }
    },
    "pexec_clk6hqpqot4o7a4sehs0": {
      "dashboard": "pexec_clk6hqpqot4o7a4sehs0",
      "name": "pexec_clk6hqpqot4o7a4sehs0",
      "panel_type": "dashboard",
      "status": "complete",
      "title": "Pipeline Execution: pexec_clk6hqpqot4o7a4sehs0",
      "data": {}
    },
    "start_pexec_clk6hqpqot4o7a4sehs0": {
      "dashboard": "pexec_clk6hqpqot4o7a4sehs0",
      "name": "start_pexec_clk6hqpqot4o7a4sehs0",
      "panel_type": "node",
      "status": "complete",
      "title": "Start: default_mod.pipeline.pipes_echo",
      "data": {
        "columns": [
          {
            "name": "id",
            "data_type": "TEXT"
          },
          {
            "name": "title",
            "data_type": "TEXT"
          },
          {
            "name": "properties",
            "data_type": "JSONB"
          }
        ],
        "rows": [
          {
            "id": "start_pexec_clk6hqpqot4o7a4sehs0",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clk6hqpqot4o7a4sehs0",
              "Status": "finished"
            },
            "title": "Start: default_mod.pipeline.pipes_echo"
          }
        ]
      },
      "properties": {
        "category": {
          "color": "green",
          "icon": "valve",
          "name": "pipeline",
          "title": "Pipeline"
        },
        "name": "start_pexec_clk6hqpqot4o7a4sehs0"
      }
    },
    "step_transform.foo": {
      "dashboard": "pexec_clk6hqpqot4o7a4sehs0",
      "name": "step_transform.foo",
      "panel_type": "node",
      "status": "complete",
      "title": "transform.foo",
      "data": {
        "columns": [
          {
            "name": "id",
            "data_type": "TEXT"
          },
          {
            "name": "title",
            "data_type": "TEXT"
          },
          {
            "name": "properties",
            "data_type": "JSONB"
          }
        ],
        "rows": [
          {
            "id": "step_transform.foo",
            "properties": {
              "Execution ID": "sexec_clk6hqpqot4o7a4seht0",
              "Input": {
                "value": "foo"
              },
              "Output": {
                "status": "finished",
                "data": {
                  "value": "foo"
                }
              },
              "Status": "finished"
            },
            "title": "0 = "
          }
        ]
      },
      "properties": {
        "category": {
          "color": "red",
          "icon": "priority_high",
          "name": "transform",
          "title": "transform"
        },
        "name": "step_transform.foo"
      }
    }
  }
}