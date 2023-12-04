{
  "schema_version": "20221222",
  "start_time": "2023-12-04T11:01:17Z",
  "end_time": "2023-12-04T11:01:17Z",
  "layout": {
    "name": "pexec_clmr2rfa2ua03efq2rjg",
    "panel_type": "dashboard",
    "children": [
      {
        "name": "execution_tree",
        "panel_type": "graph"
      }
    ]
  },
  "panels": {
    "edge_start_pexec_clmr2rfa2ua03efq2rjg_to_step_transform.foo": {
      "dashboard": "pexec_clmr2rfa2ua03efq2rjg",
      "name": "edge_start_pexec_clmr2rfa2ua03efq2rjg_to_step_transform.foo",
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
            "from_id": "start_pexec_clmr2rfa2ua03efq2rjg",
            "to_id": "step_transform.foo"
          }
        ]
      },
      "properties": {
        "name": "edge_start_pexec_clmr2rfa2ua03efq2rjg_to_step_transform.foo"
      }
    },
    "edge_step_transform.foo_to_end_pexec_clmr2rfa2ua03efq2rjg": {
      "dashboard": "pexec_clmr2rfa2ua03efq2rjg",
      "name": "edge_step_transform.foo_to_end_pexec_clmr2rfa2ua03efq2rjg",
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
            "to_id": "end_pexec_clmr2rfa2ua03efq2rjg"
          }
        ]
      },
      "properties": {
        "name": "edge_step_transform.foo_to_end_pexec_clmr2rfa2ua03efq2rjg"
      }
    },
    "end_pexec_clmr2rfa2ua03efq2rjg": {
      "dashboard": "pexec_clmr2rfa2ua03efq2rjg",
      "name": "end_pexec_clmr2rfa2ua03efq2rjg",
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
            "id": "end_pexec_clmr2rfa2ua03efq2rjg",
            "properties": {
              "Execution ID": "pexec_clmr2rfa2ua03efq2rjg",
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
        "name": "end_pexec_clmr2rfa2ua03efq2rjg"
      }
    },
    "execution_tree": {
      "dashboard": "pexec_clmr2rfa2ua03efq2rjg",
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
            "from_id": "start_pexec_clmr2rfa2ua03efq2rjg",
            "to_id": "step_transform.foo"
          },
          {
            "from_id": "step_transform.foo",
            "to_id": "end_pexec_clmr2rfa2ua03efq2rjg"
          },
          {
            "id": "start_pexec_clmr2rfa2ua03efq2rjg",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clmr2rfa2ua03efq2rjg",
              "Status": "finished"
            },
            "title": "Start: default_mod.pipeline.pipes_echo"
          },
          {
            "id": "end_pexec_clmr2rfa2ua03efq2rjg",
            "properties": {
              "Execution ID": "pexec_clmr2rfa2ua03efq2rjg",
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
              "Execution ID": "sexec_clmr2tfa2ua03efq2rkg",
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
          "edge_start_pexec_clmr2rfa2ua03efq2rjg_to_step_transform.foo",
          "edge_step_transform.foo_to_end_pexec_clmr2rfa2ua03efq2rjg"
        ],
        "name": "execution",
        "nodes": [
          "start_pexec_clmr2rfa2ua03efq2rjg",
          "end_pexec_clmr2rfa2ua03efq2rjg",
          "step_transform.foo"
        ]
      }
    },
    "pexec_clmr2rfa2ua03efq2rjg": {
      "dashboard": "pexec_clmr2rfa2ua03efq2rjg",
      "name": "pexec_clmr2rfa2ua03efq2rjg",
      "panel_type": "dashboard",
      "status": "complete",
      "title": "Pipeline Execution: pexec_clmr2rfa2ua03efq2rjg",
      "data": {}
    },
    "start_pexec_clmr2rfa2ua03efq2rjg": {
      "dashboard": "pexec_clmr2rfa2ua03efq2rjg",
      "name": "start_pexec_clmr2rfa2ua03efq2rjg",
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
            "id": "start_pexec_clmr2rfa2ua03efq2rjg",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clmr2rfa2ua03efq2rjg",
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
        "name": "start_pexec_clmr2rfa2ua03efq2rjg"
      }
    },
    "step_transform.foo": {
      "dashboard": "pexec_clmr2rfa2ua03efq2rjg",
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
              "Execution ID": "sexec_clmr2tfa2ua03efq2rkg",
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