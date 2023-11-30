{
  "schema_version": "20221222",
  "start_time": "2023-11-30T10:45:33Z",
  "end_time": "2023-11-30T10:45:33Z",
  "layout": {
    "name": "pexec_clk6fj9qot4nq9q78rog",
    "panel_type": "dashboard",
    "children": [
      {
        "name": "execution_tree",
        "panel_type": "graph"
      }
    ]
  },
  "panels": {
    "edge_start_pexec_clk6fj9qot4nq9q78rog_to_step_transform.foo": {
      "dashboard": "pexec_clk6fj9qot4nq9q78rog",
      "name": "edge_start_pexec_clk6fj9qot4nq9q78rog_to_step_transform.foo",
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
            "from_id": "start_pexec_clk6fj9qot4nq9q78rog",
            "to_id": "step_transform.foo"
          }
        ]
      },
      "properties": {
        "name": "edge_start_pexec_clk6fj9qot4nq9q78rog_to_step_transform.foo"
      }
    },
    "edge_step_transform.foo_to_end_pexec_clk6fj9qot4nq9q78rog": {
      "dashboard": "pexec_clk6fj9qot4nq9q78rog",
      "name": "edge_step_transform.foo_to_end_pexec_clk6fj9qot4nq9q78rog",
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
            "to_id": "end_pexec_clk6fj9qot4nq9q78rog"
          }
        ]
      },
      "properties": {
        "name": "edge_step_transform.foo_to_end_pexec_clk6fj9qot4nq9q78rog"
      }
    },
    "end_pexec_clk6fj9qot4nq9q78rog": {
      "dashboard": "pexec_clk6fj9qot4nq9q78rog",
      "name": "end_pexec_clk6fj9qot4nq9q78rog",
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
            "id": "end_pexec_clk6fj9qot4nq9q78rog",
            "properties": {
              "Execution ID": "pexec_clk6fj9qot4nq9q78rog",
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
        "name": "end_pexec_clk6fj9qot4nq9q78rog"
      }
    },
    "execution_tree": {
      "dashboard": "pexec_clk6fj9qot4nq9q78rog",
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
            "from_id": "start_pexec_clk6fj9qot4nq9q78rog",
            "to_id": "step_transform.foo"
          },
          {
            "from_id": "step_transform.foo",
            "to_id": "end_pexec_clk6fj9qot4nq9q78rog"
          },
          {
            "id": "start_pexec_clk6fj9qot4nq9q78rog",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clk6fj9qot4nq9q78rog",
              "Status": "finished"
            },
            "title": "Start: default_mod.pipeline.pipes_echo"
          },
          {
            "id": "end_pexec_clk6fj9qot4nq9q78rog",
            "properties": {
              "Execution ID": "pexec_clk6fj9qot4nq9q78rog",
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
              "Execution ID": "sexec_clk6fj9qot4nq9q78rpg",
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
          "edge_start_pexec_clk6fj9qot4nq9q78rog_to_step_transform.foo",
          "edge_step_transform.foo_to_end_pexec_clk6fj9qot4nq9q78rog"
        ],
        "name": "execution",
        "nodes": [
          "start_pexec_clk6fj9qot4nq9q78rog",
          "end_pexec_clk6fj9qot4nq9q78rog",
          "step_transform.foo"
        ]
      }
    },
    "pexec_clk6fj9qot4nq9q78rog": {
      "dashboard": "pexec_clk6fj9qot4nq9q78rog",
      "name": "pexec_clk6fj9qot4nq9q78rog",
      "panel_type": "dashboard",
      "status": "complete",
      "title": "Pipeline Execution: pexec_clk6fj9qot4nq9q78rog",
      "data": {}
    },
    "start_pexec_clk6fj9qot4nq9q78rog": {
      "dashboard": "pexec_clk6fj9qot4nq9q78rog",
      "name": "start_pexec_clk6fj9qot4nq9q78rog",
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
            "id": "start_pexec_clk6fj9qot4nq9q78rog",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clk6fj9qot4nq9q78rog",
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
        "name": "start_pexec_clk6fj9qot4nq9q78rog"
      }
    },
    "step_transform.foo": {
      "dashboard": "pexec_clk6fj9qot4nq9q78rog",
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
              "Execution ID": "sexec_clk6fj9qot4nq9q78rpg",
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