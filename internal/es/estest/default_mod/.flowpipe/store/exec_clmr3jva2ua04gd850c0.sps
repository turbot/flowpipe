{
  "schema_version": "20221222",
  "start_time": "2023-12-04T11:02:42Z",
  "end_time": "2023-12-04T11:02:42Z",
  "layout": {
    "name": "pexec_clmr3jva2ua04gd850cg",
    "panel_type": "dashboard",
    "children": [
      {
        "name": "execution_tree",
        "panel_type": "graph"
      }
    ]
  },
  "panels": {
    "edge_start_pexec_clmr3jva2ua04gd850cg_to_step_transform.foo": {
      "dashboard": "pexec_clmr3jva2ua04gd850cg",
      "name": "edge_start_pexec_clmr3jva2ua04gd850cg_to_step_transform.foo",
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
            "from_id": "start_pexec_clmr3jva2ua04gd850cg",
            "to_id": "step_transform.foo"
          }
        ]
      },
      "properties": {
        "name": "edge_start_pexec_clmr3jva2ua04gd850cg_to_step_transform.foo"
      }
    },
    "edge_step_transform.foo_to_end_pexec_clmr3jva2ua04gd850cg": {
      "dashboard": "pexec_clmr3jva2ua04gd850cg",
      "name": "edge_step_transform.foo_to_end_pexec_clmr3jva2ua04gd850cg",
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
            "to_id": "end_pexec_clmr3jva2ua04gd850cg"
          }
        ]
      },
      "properties": {
        "name": "edge_step_transform.foo_to_end_pexec_clmr3jva2ua04gd850cg"
      }
    },
    "end_pexec_clmr3jva2ua04gd850cg": {
      "dashboard": "pexec_clmr3jva2ua04gd850cg",
      "name": "end_pexec_clmr3jva2ua04gd850cg",
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
            "id": "end_pexec_clmr3jva2ua04gd850cg",
            "properties": {
              "Execution ID": "pexec_clmr3jva2ua04gd850cg",
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
        "name": "end_pexec_clmr3jva2ua04gd850cg"
      }
    },
    "execution_tree": {
      "dashboard": "pexec_clmr3jva2ua04gd850cg",
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
            "from_id": "start_pexec_clmr3jva2ua04gd850cg",
            "to_id": "step_transform.foo"
          },
          {
            "from_id": "step_transform.foo",
            "to_id": "end_pexec_clmr3jva2ua04gd850cg"
          },
          {
            "id": "start_pexec_clmr3jva2ua04gd850cg",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clmr3jva2ua04gd850cg",
              "Status": "finished"
            },
            "title": "Start: default_mod.pipeline.pipes_echo"
          },
          {
            "id": "end_pexec_clmr3jva2ua04gd850cg",
            "properties": {
              "Execution ID": "pexec_clmr3jva2ua04gd850cg",
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
              "Execution ID": "sexec_clmr3kna2ua04gd850dg",
              "Input": {
                "value": "<redacted>"
              },
              "Output": {
                "status": "finished",
                "data": {
                  "value": "<redacted>"
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
          "edge_start_pexec_clmr3jva2ua04gd850cg_to_step_transform.foo",
          "edge_step_transform.foo_to_end_pexec_clmr3jva2ua04gd850cg"
        ],
        "name": "execution",
        "nodes": [
          "start_pexec_clmr3jva2ua04gd850cg",
          "end_pexec_clmr3jva2ua04gd850cg",
          "step_transform.foo"
        ]
      }
    },
    "pexec_clmr3jva2ua04gd850cg": {
      "dashboard": "pexec_clmr3jva2ua04gd850cg",
      "name": "pexec_clmr3jva2ua04gd850cg",
      "panel_type": "dashboard",
      "status": "complete",
      "title": "Pipeline Execution: pexec_clmr3jva2ua04gd850cg",
      "data": {}
    },
    "start_pexec_clmr3jva2ua04gd850cg": {
      "dashboard": "pexec_clmr3jva2ua04gd850cg",
      "name": "start_pexec_clmr3jva2ua04gd850cg",
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
            "id": "start_pexec_clmr3jva2ua04gd850cg",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clmr3jva2ua04gd850cg",
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
        "name": "start_pexec_clmr3jva2ua04gd850cg"
      }
    },
    "step_transform.foo": {
      "dashboard": "pexec_clmr3jva2ua04gd850cg",
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
              "Execution ID": "sexec_clmr3kna2ua04gd850dg",
              "Input": {
                "value": "<redacted>"
              },
              "Output": {
                "status": "finished",
                "data": {
                  "value": "<redacted>"
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