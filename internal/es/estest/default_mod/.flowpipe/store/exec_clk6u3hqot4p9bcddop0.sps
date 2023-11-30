{
  "schema_version": "20221222",
  "start_time": "2023-11-30T11:16:31Z",
  "end_time": "2023-11-30T11:16:31Z",
  "layout": {
    "name": "pexec_clk6u3hqot4p9bcddopg",
    "panel_type": "dashboard",
    "children": [
      {
        "name": "execution_tree",
        "panel_type": "graph"
      }
    ]
  },
  "panels": {
    "edge_start_pexec_clk6u3hqot4p9bcddopg_to_step_transform.foo": {
      "dashboard": "pexec_clk6u3hqot4p9bcddopg",
      "name": "edge_start_pexec_clk6u3hqot4p9bcddopg_to_step_transform.foo",
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
            "from_id": "start_pexec_clk6u3hqot4p9bcddopg",
            "to_id": "step_transform.foo"
          }
        ]
      },
      "properties": {
        "name": "edge_start_pexec_clk6u3hqot4p9bcddopg_to_step_transform.foo"
      }
    },
    "edge_step_transform.foo_to_end_pexec_clk6u3hqot4p9bcddopg": {
      "dashboard": "pexec_clk6u3hqot4p9bcddopg",
      "name": "edge_step_transform.foo_to_end_pexec_clk6u3hqot4p9bcddopg",
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
            "to_id": "end_pexec_clk6u3hqot4p9bcddopg"
          }
        ]
      },
      "properties": {
        "name": "edge_step_transform.foo_to_end_pexec_clk6u3hqot4p9bcddopg"
      }
    },
    "end_pexec_clk6u3hqot4p9bcddopg": {
      "dashboard": "pexec_clk6u3hqot4p9bcddopg",
      "name": "end_pexec_clk6u3hqot4p9bcddopg",
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
            "id": "end_pexec_clk6u3hqot4p9bcddopg",
            "properties": {
              "Execution ID": "pexec_clk6u3hqot4p9bcddopg",
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
        "name": "end_pexec_clk6u3hqot4p9bcddopg"
      }
    },
    "execution_tree": {
      "dashboard": "pexec_clk6u3hqot4p9bcddopg",
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
            "from_id": "start_pexec_clk6u3hqot4p9bcddopg",
            "to_id": "step_transform.foo"
          },
          {
            "from_id": "step_transform.foo",
            "to_id": "end_pexec_clk6u3hqot4p9bcddopg"
          },
          {
            "id": "start_pexec_clk6u3hqot4p9bcddopg",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clk6u3hqot4p9bcddopg",
              "Status": "finished"
            },
            "title": "Start: default_mod.pipeline.pipes_echo"
          },
          {
            "id": "end_pexec_clk6u3hqot4p9bcddopg",
            "properties": {
              "Execution ID": "pexec_clk6u3hqot4p9bcddopg",
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
              "Execution ID": "sexec_clk6u3pqot4p9bcddoqg",
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
          "edge_start_pexec_clk6u3hqot4p9bcddopg_to_step_transform.foo",
          "edge_step_transform.foo_to_end_pexec_clk6u3hqot4p9bcddopg"
        ],
        "name": "execution",
        "nodes": [
          "start_pexec_clk6u3hqot4p9bcddopg",
          "end_pexec_clk6u3hqot4p9bcddopg",
          "step_transform.foo"
        ]
      }
    },
    "pexec_clk6u3hqot4p9bcddopg": {
      "dashboard": "pexec_clk6u3hqot4p9bcddopg",
      "name": "pexec_clk6u3hqot4p9bcddopg",
      "panel_type": "dashboard",
      "status": "complete",
      "title": "Pipeline Execution: pexec_clk6u3hqot4p9bcddopg",
      "data": {}
    },
    "start_pexec_clk6u3hqot4p9bcddopg": {
      "dashboard": "pexec_clk6u3hqot4p9bcddopg",
      "name": "start_pexec_clk6u3hqot4p9bcddopg",
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
            "id": "start_pexec_clk6u3hqot4p9bcddopg",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clk6u3hqot4p9bcddopg",
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
        "name": "start_pexec_clk6u3hqot4p9bcddopg"
      }
    },
    "step_transform.foo": {
      "dashboard": "pexec_clk6u3hqot4p9bcddopg",
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
              "Execution ID": "sexec_clk6u3pqot4p9bcddoqg",
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