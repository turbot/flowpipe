{
  "schema_version": "20221222",
  "start_time": "2023-11-30T15:59:07Z",
  "end_time": "2023-11-30T15:59:07Z",
  "layout": {
    "name": "pexec_clkb2ipqot4kekkq23ug",
    "panel_type": "dashboard",
    "children": [
      {
        "name": "execution_tree",
        "panel_type": "graph"
      }
    ]
  },
  "panels": {
    "edge_start_pexec_clkb2ipqot4kekkq23ug_to_step_transform.foo": {
      "dashboard": "pexec_clkb2ipqot4kekkq23ug",
      "name": "edge_start_pexec_clkb2ipqot4kekkq23ug_to_step_transform.foo",
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
            "from_id": "start_pexec_clkb2ipqot4kekkq23ug",
            "to_id": "step_transform.foo"
          }
        ]
      },
      "properties": {
        "name": "edge_start_pexec_clkb2ipqot4kekkq23ug_to_step_transform.foo"
      }
    },
    "edge_step_transform.foo_to_end_pexec_clkb2ipqot4kekkq23ug": {
      "dashboard": "pexec_clkb2ipqot4kekkq23ug",
      "name": "edge_step_transform.foo_to_end_pexec_clkb2ipqot4kekkq23ug",
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
            "to_id": "end_pexec_clkb2ipqot4kekkq23ug"
          }
        ]
      },
      "properties": {
        "name": "edge_step_transform.foo_to_end_pexec_clkb2ipqot4kekkq23ug"
      }
    },
    "end_pexec_clkb2ipqot4kekkq23ug": {
      "dashboard": "pexec_clkb2ipqot4kekkq23ug",
      "name": "end_pexec_clkb2ipqot4kekkq23ug",
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
            "id": "end_pexec_clkb2ipqot4kekkq23ug",
            "properties": {
              "Execution ID": "pexec_clkb2ipqot4kekkq23ug",
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
        "name": "end_pexec_clkb2ipqot4kekkq23ug"
      }
    },
    "execution_tree": {
      "dashboard": "pexec_clkb2ipqot4kekkq23ug",
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
            "from_id": "start_pexec_clkb2ipqot4kekkq23ug",
            "to_id": "step_transform.foo"
          },
          {
            "from_id": "step_transform.foo",
            "to_id": "end_pexec_clkb2ipqot4kekkq23ug"
          },
          {
            "id": "start_pexec_clkb2ipqot4kekkq23ug",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clkb2ipqot4kekkq23ug",
              "Status": "finished"
            },
            "title": "Start: default_mod.pipeline.pipes_echo"
          },
          {
            "id": "end_pexec_clkb2ipqot4kekkq23ug",
            "properties": {
              "Execution ID": "pexec_clkb2ipqot4kekkq23ug",
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
              "Execution ID": "sexec_clkb2ipqot4kekkq23vg",
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
          "edge_start_pexec_clkb2ipqot4kekkq23ug_to_step_transform.foo",
          "edge_step_transform.foo_to_end_pexec_clkb2ipqot4kekkq23ug"
        ],
        "name": "execution",
        "nodes": [
          "start_pexec_clkb2ipqot4kekkq23ug",
          "end_pexec_clkb2ipqot4kekkq23ug",
          "step_transform.foo"
        ]
      }
    },
    "pexec_clkb2ipqot4kekkq23ug": {
      "dashboard": "pexec_clkb2ipqot4kekkq23ug",
      "name": "pexec_clkb2ipqot4kekkq23ug",
      "panel_type": "dashboard",
      "status": "complete",
      "title": "Pipeline Execution: pexec_clkb2ipqot4kekkq23ug",
      "data": {}
    },
    "start_pexec_clkb2ipqot4kekkq23ug": {
      "dashboard": "pexec_clkb2ipqot4kekkq23ug",
      "name": "start_pexec_clkb2ipqot4kekkq23ug",
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
            "id": "start_pexec_clkb2ipqot4kekkq23ug",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clkb2ipqot4kekkq23ug",
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
        "name": "start_pexec_clkb2ipqot4kekkq23ug"
      }
    },
    "step_transform.foo": {
      "dashboard": "pexec_clkb2ipqot4kekkq23ug",
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
              "Execution ID": "sexec_clkb2ipqot4kekkq23vg",
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