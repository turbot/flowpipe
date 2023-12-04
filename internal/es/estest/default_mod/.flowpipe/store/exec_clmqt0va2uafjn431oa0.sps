{
  "schema_version": "20221222",
  "start_time": "2023-12-04T10:48:35Z",
  "end_time": "2023-12-04T10:48:35Z",
  "layout": {
    "name": "pexec_clmqt0va2uafjn431oag",
    "panel_type": "dashboard",
    "children": [
      {
        "name": "execution_tree",
        "panel_type": "graph"
      }
    ]
  },
  "panels": {
    "edge_start_pexec_clmqt0va2uafjn431oag_to_step_transform.foo": {
      "dashboard": "pexec_clmqt0va2uafjn431oag",
      "name": "edge_start_pexec_clmqt0va2uafjn431oag_to_step_transform.foo",
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
            "from_id": "start_pexec_clmqt0va2uafjn431oag",
            "to_id": "step_transform.foo"
          }
        ]
      },
      "properties": {
        "name": "edge_start_pexec_clmqt0va2uafjn431oag_to_step_transform.foo"
      }
    },
    "edge_step_transform.foo_to_end_pexec_clmqt0va2uafjn431oag": {
      "dashboard": "pexec_clmqt0va2uafjn431oag",
      "name": "edge_step_transform.foo_to_end_pexec_clmqt0va2uafjn431oag",
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
            "to_id": "end_pexec_clmqt0va2uafjn431oag"
          }
        ]
      },
      "properties": {
        "name": "edge_step_transform.foo_to_end_pexec_clmqt0va2uafjn431oag"
      }
    },
    "end_pexec_clmqt0va2uafjn431oag": {
      "dashboard": "pexec_clmqt0va2uafjn431oag",
      "name": "end_pexec_clmqt0va2uafjn431oag",
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
            "id": "end_pexec_clmqt0va2uafjn431oag",
            "properties": {
              "Execution ID": "pexec_clmqt0va2uafjn431oag",
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
        "name": "end_pexec_clmqt0va2uafjn431oag"
      }
    },
    "execution_tree": {
      "dashboard": "pexec_clmqt0va2uafjn431oag",
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
            "from_id": "start_pexec_clmqt0va2uafjn431oag",
            "to_id": "step_transform.foo"
          },
          {
            "from_id": "step_transform.foo",
            "to_id": "end_pexec_clmqt0va2uafjn431oag"
          },
          {
            "id": "start_pexec_clmqt0va2uafjn431oag",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clmqt0va2uafjn431oag",
              "Status": "finished"
            },
            "title": "Start: default_mod.pipeline.pipes_echo"
          },
          {
            "id": "end_pexec_clmqt0va2uafjn431oag",
            "properties": {
              "Execution ID": "pexec_clmqt0va2uafjn431oag",
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
              "Execution ID": "sexec_clmqt0va2uafjn431obg",
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
          "edge_start_pexec_clmqt0va2uafjn431oag_to_step_transform.foo",
          "edge_step_transform.foo_to_end_pexec_clmqt0va2uafjn431oag"
        ],
        "name": "execution",
        "nodes": [
          "start_pexec_clmqt0va2uafjn431oag",
          "end_pexec_clmqt0va2uafjn431oag",
          "step_transform.foo"
        ]
      }
    },
    "pexec_clmqt0va2uafjn431oag": {
      "dashboard": "pexec_clmqt0va2uafjn431oag",
      "name": "pexec_clmqt0va2uafjn431oag",
      "panel_type": "dashboard",
      "status": "complete",
      "title": "Pipeline Execution: pexec_clmqt0va2uafjn431oag",
      "data": {}
    },
    "start_pexec_clmqt0va2uafjn431oag": {
      "dashboard": "pexec_clmqt0va2uafjn431oag",
      "name": "start_pexec_clmqt0va2uafjn431oag",
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
            "id": "start_pexec_clmqt0va2uafjn431oag",
            "properties": {
              "Args": null,
              "Execution ID": "pexec_clmqt0va2uafjn431oag",
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
        "name": "start_pexec_clmqt0va2uafjn431oag"
      }
    },
    "step_transform.foo": {
      "dashboard": "pexec_clmqt0va2uafjn431oag",
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
              "Execution ID": "sexec_clmqt0va2uafjn431obg",
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