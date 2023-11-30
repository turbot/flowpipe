{
  "schema_version": "20221222",
  "start_time": "2023-11-30T16:02:53Z",
  "end_time": "2023-11-30T16:02:53Z",
  "layout": {
    "name": "pexec_clkb4b9qot4kk3hhlgk0",
    "panel_type": "dashboard",
    "children": [
      {
        "name": "execution_tree",
        "panel_type": "graph"
      }
    ]
  },
  "panels": {
    "edge_start_pexec_clkb4b9qot4kk3hhlgk0_to_step_transform.foo": {
      "dashboard": "pexec_clkb4b9qot4kk3hhlgk0",
      "name": "edge_start_pexec_clkb4b9qot4kk3hhlgk0_to_step_transform.foo",
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
            "from_id": "start_pexec_clkb4b9qot4kk3hhlgk0",
            "to_id": "step_transform.foo"
          }
        ]
      },
      "properties": {
        "name": "edge_start_pexec_clkb4b9qot4kk3hhlgk0_to_step_transform.foo"
      }
    },
    "edge_step_transform.foo_to_end_pexec_clkb4b9qot4kk3hhlgk0": {
      "dashboard": "pexec_clkb4b9qot4kk3hhlgk0",
      "name": "edge_step_transform.foo_to_end_pexec_clkb4b9qot4kk3hhlgk0",
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
            "to_id": "end_pexec_clkb4b9qot4kk3hhlgk0"
          }
        ]
      },
      "properties": {
        "name": "edge_step_transform.foo_to_end_pexec_clkb4b9qot4kk3hhlgk0"
      }
    },
    "end_pexec_clkb4b9qot4kk3hhlgk0": {
      "dashboard": "pexec_clkb4b9qot4kk3hhlgk0",
      "name": "end_pexec_clkb4b9qot4kk3hhlgk0",
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
            "id": "end_pexec_clkb4b9qot4kk3hhlgk0",
            "properties": {
              "Execution ID": "pexec_clkb4b9qot4kk3hhlgk0",
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
        "name": "end_pexec_clkb4b9qot4kk3hhlgk0"
      }
    },
    "execution_tree": {
      "dashboard": "pexec_clkb4b9qot4kk3hhlgk0",
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
            "from_id": "start_pexec_clkb4b9qot4kk3hhlgk0",
            "to_id": "step_transform.foo"
          },
          {
            "from_id": "step_transform.foo",
            "to_id": "end_pexec_clkb4b9qot4kk3hhlgk0"
          },
          {
            "id": "start_pexec_clkb4b9qot4kk3hhlgk0",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clkb4b9qot4kk3hhlgk0",
              "Status": "finished"
            },
            "title": "Start: default_mod.pipeline.pipes_echo"
          },
          {
            "id": "end_pexec_clkb4b9qot4kk3hhlgk0",
            "properties": {
              "Execution ID": "pexec_clkb4b9qot4kk3hhlgk0",
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
              "Execution ID": "sexec_clkb4b9qot4kk3hhlgl0",
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
          "edge_start_pexec_clkb4b9qot4kk3hhlgk0_to_step_transform.foo",
          "edge_step_transform.foo_to_end_pexec_clkb4b9qot4kk3hhlgk0"
        ],
        "name": "execution",
        "nodes": [
          "start_pexec_clkb4b9qot4kk3hhlgk0",
          "end_pexec_clkb4b9qot4kk3hhlgk0",
          "step_transform.foo"
        ]
      }
    },
    "pexec_clkb4b9qot4kk3hhlgk0": {
      "dashboard": "pexec_clkb4b9qot4kk3hhlgk0",
      "name": "pexec_clkb4b9qot4kk3hhlgk0",
      "panel_type": "dashboard",
      "status": "complete",
      "title": "Pipeline Execution: pexec_clkb4b9qot4kk3hhlgk0",
      "data": {}
    },
    "start_pexec_clkb4b9qot4kk3hhlgk0": {
      "dashboard": "pexec_clkb4b9qot4kk3hhlgk0",
      "name": "start_pexec_clkb4b9qot4kk3hhlgk0",
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
            "id": "start_pexec_clkb4b9qot4kk3hhlgk0",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clkb4b9qot4kk3hhlgk0",
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
        "name": "start_pexec_clkb4b9qot4kk3hhlgk0"
      }
    },
    "step_transform.foo": {
      "dashboard": "pexec_clkb4b9qot4kk3hhlgk0",
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
              "Execution ID": "sexec_clkb4b9qot4kk3hhlgl0",
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