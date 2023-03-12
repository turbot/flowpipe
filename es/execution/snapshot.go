package execution

import (
	"fmt"
	"time"
)

type Snapshot struct {
	SchemaVersion string                   `json:"schema_version"`
	StartTime     string                   `json:"start_time"`
	EndTime       string                   `json:"end_time"`
	Layout        SnapshotLayout           `json:"layout"`
	Panels        map[string]SnapshotPanel `json:"panels"`
}

type SnapshotLayout struct {
	Name      string           `json:"name"`
	PanelType string           `json:"panel_type"`
	Children  []SnapshotLayout `json:"children,omitempty"`
}

type SnapshotPanel struct {
	Dashboard   string                 `json:"dashboard"`
	Name        string                 `json:"name"`
	PanelType   string                 `json:"panel_type"`
	Status      string                 `json:"status"`
	Title       string                 `json:"title,omitempty"`
	DisplayType string                 `json:"display_type,omitempty"`
	Width       int                    `json:"width,omitempty"`
	Data        SnapshotPanelData      `json:"data,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

type SnapshotPanelData struct {
	Columns []SnapshotPanelDataColumn `json:"columns,omitempty"`
	Rows    []SnapshotPanelDataRow    `json:"rows,omitempty"`
}

type SnapshotPanelDataColumn struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

type SnapshotPanelDataRow map[string]interface{}

func (ex *Execution) Snapshot(pipelineExecutionID string) (*Snapshot, error) {
	pe, ok := ex.PipelineExecutions[pipelineExecutionID]
	if !ok {
		return nil, fmt.Errorf("pipeline execution %s not found", pipelineExecutionID)
	}

	snapshot := &Snapshot{
		SchemaVersion: "20221222",
		StartTime:     time.Now().UTC().Format(time.RFC3339),
		Layout: SnapshotLayout{
			Name:      pe.ID,
			PanelType: "dashboard",
			Children: []SnapshotLayout{
				{
					Name:      "execution_tree",
					PanelType: "graph",
				},
				{
					Name:      "step_definitions",
					PanelType: "table",
				},
			},
		},
		Panels: map[string]SnapshotPanel{},
	}

	dashboardPanel := SnapshotPanel{
		Dashboard: pe.ID,
		Name:      pe.ID,
		PanelType: "dashboard",
		Status:    "complete",
		Title:     fmt.Sprintf("Pipeline Execution: %s", pe.ID),
	}

	executionPanel := SnapshotPanel{
		Dashboard:   pe.ID,
		Name:        "execution_tree",
		PanelType:   "graph",
		Status:      "complete",
		Title:       "Execution",
		DisplayType: "graph",
		/*
				Data: SnapshotPanelData{
					Columns: []SnapshotPanelDataColumn{
						{Name: "from_id", DataType: "TEXT"},
						{Name: "to_id", DataType: "TEXT"},
						{Name: "id", DataType: "TEXT"},
						{Name: "title", DataType: "TEXT"},
						{Name: "properties", DataType: "JSONB"},
					},
			},
		*/
		Properties: map[string]interface{}{
			"name":      "execution",
			"direction": "TD",
		},
	}

	tablePanel := SnapshotPanel{
		Dashboard: pe.ID,
		Name:      "step_definitions",
		PanelType: "table",
		Status:    "complete",
		Title:     "Step Definitions",
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "name", DataType: "TEXT"},
				{Name: "type", DataType: "TEXT"},
				{Name: "depends_on", DataType: "JSONB"},
			},
		},
		Properties: map[string]interface{}{
			"name": "step_definitions",
		},
	}

	// Cacluate the rows
	var edgeName string
	nodeNames := []string{}
	edgeNames := []string{}
	tableRows := []SnapshotPanelDataRow{}

	pd, err := ex.PipelineDefinition(pe.ID)
	if err != nil {
		return nil, err
	}

	nodeNames = append(nodeNames, "start_"+pe.ID)
	snapshot.Panels["start_"+pe.ID] = SnapshotPanel{
		Dashboard: pe.ID,
		Name:      "start_" + pe.ID,
		PanelType: "node",
		Status:    "complete",
		Title:     "Start: " + pd.Name,
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "id", DataType: "TEXT"},
				{Name: "title", DataType: "TEXT"},
				{Name: "properties", DataType: "JSONB"},
			},
			Rows: []SnapshotPanelDataRow{
				{
					"id":    "start_" + pe.ID,
					"title": "Start: " + pd.Name,
					"properties": map[string]interface{}{
						"Execution ID": pe.ID,
						"Input":        pe.Input,
						"Status":       pe.Status,
					},
				},
			},
		},
		Properties: map[string]interface{}{
			"name":     "start_" + pe.ID,
			"category": Category("pipeline"),
		},
	}

	nodeNames = append(nodeNames, "end_"+pe.ID)
	snapshot.Panels["end_"+pe.ID] = SnapshotPanel{
		Dashboard: pe.ID,
		Name:      "end_" + pe.ID,
		PanelType: "node",
		Status:    "complete",
		Title:     "End: " + pd.Name,
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "id", DataType: "TEXT"},
				{Name: "title", DataType: "TEXT"},
				{Name: "properties", DataType: "JSONB"},
			},
			Rows: []SnapshotPanelDataRow{
				{
					"id":    "end_" + pe.ID,
					"title": "End: " + pd.Name,
					"properties": map[string]interface{}{
						"Execution ID": pe.ID,
						"Status":       pe.Status,
					},
				},
			},
		},
		Properties: map[string]interface{}{
			"name":     "end_" + pe.ID,
			"category": Category("pipeline"),
		},
	}

	dependedOn := map[string]bool{}

	// Check each step definition in the pipeline
	for _, sd := range pd.Steps {

		// Find all executions for the step
		executionsForStep := []*StepExecution{}
		for _, se := range ex.StepExecutions {
			if se.PipelineExecutionID != pe.ID {
				continue
			}
			if se.Name != sd.Name {
				continue
			}
			executionsForStep = append(executionsForStep, se)
		}

		stepToID := sd.Name

		if len(executionsForStep) == 0 {

			nodeNames = append(nodeNames, sd.Name)
			snapshot.Panels[sd.Name] = SnapshotPanel{
				Dashboard: pe.ID,
				Name:      sd.Name,
				PanelType: "node",
				Status:    "complete",
				Title:     sd.Name,
				Data: SnapshotPanelData{
					Columns: []SnapshotPanelDataColumn{
						{Name: "id", DataType: "TEXT"},
						{Name: "title", DataType: "TEXT"},
						{Name: "properties", DataType: "JSONB"},
					},
					Rows: []SnapshotPanelDataRow{
						{
							"id":    sd.Name,
							"title": sd.Name,
							"properties": map[string]interface{}{
								"Type": sd.Type,
							},
						},
					},
				},
				Properties: map[string]interface{}{
					"name":     sd.Name,
					"category": Category(sd.Type),
				},
			}

		} else if len(executionsForStep) <= 1 {
			// Single execution
			se := executionsForStep[0]

			nodeNames = append(nodeNames, sd.Name)
			snapshot.Panels[sd.Name] = SnapshotPanel{
				Dashboard: pe.ID,
				Name:      sd.Name,
				PanelType: "node",
				Status:    "complete",
				Title:     sd.Name,
				Data: SnapshotPanelData{
					Columns: []SnapshotPanelDataColumn{
						{Name: "id", DataType: "TEXT"},
						{Name: "title", DataType: "TEXT"},
						{Name: "properties", DataType: "JSONB"},
					},
					Rows: []SnapshotPanelDataRow{
						{
							"id":    sd.Name,
							"title": sd.Name,
							"properties": map[string]interface{}{
								"Execution ID": se.ID,
								"Input":        se.Input,
								"Output":       se.Output,
								"Status":       se.Status,
								"Type":         sd.Type,
							},
						},
					},
				},
				Properties: map[string]interface{}{
					"name":     sd.Name,
					"category": Category(sd.Type),
				},
			}

		} else {
			// Multiple executions. Add a start node for the step, execution nodes
			// in the middle, all converging back into an end node.

			// Start of step
			stepToID = "start_" + sd.Name

			// Start of step
			nodeNames = append(nodeNames, "start_"+sd.Name)
			snapshot.Panels["start_"+sd.Name] = SnapshotPanel{
				Dashboard: pe.ID,
				Name:      "start_" + sd.Name,
				PanelType: "node",
				Status:    "complete",
				Title:     "Start: " + sd.Name,
				Data: SnapshotPanelData{
					Columns: []SnapshotPanelDataColumn{
						{Name: "id", DataType: "TEXT"},
						{Name: "title", DataType: "TEXT"},
						{Name: "properties", DataType: "JSONB"},
					},
					Rows: []SnapshotPanelDataRow{
						{
							"id":    "start_" + sd.Name,
							"title": "Start: " + sd.Name,
							"properties": map[string]interface{}{
								"Type": sd.Type,
							},
						},
					},
				},
				Properties: map[string]interface{}{
					"name":     "start_" + sd.Name,
					"category": Category(sd.Type),
				},
			}

			// End of step
			nodeNames = append(nodeNames, sd.Name)
			snapshot.Panels[sd.Name] = SnapshotPanel{
				Dashboard: pe.ID,
				Name:      sd.Name,
				PanelType: "node",
				Status:    "complete",
				Title:     "End: " + sd.Name,
				Data: SnapshotPanelData{
					Columns: []SnapshotPanelDataColumn{
						{Name: "id", DataType: "TEXT"},
						{Name: "title", DataType: "TEXT"},
						{Name: "properties", DataType: "JSONB"},
					},
					Rows: []SnapshotPanelDataRow{
						{
							"id":    sd.Name,
							"title": "End: " + sd.Name,
							"properties": map[string]interface{}{
								"Type": sd.Type,
							},
						},
					},
				},
				Properties: map[string]interface{}{
					"name":     sd.Name,
					"category": Category(sd.Type),
				},
			}

			for _, se := range executionsForStep {

				// Node for the step execution
				nodeNames = append(nodeNames, se.ID)
				snapshot.Panels[se.ID] = SnapshotPanel{
					Dashboard: pe.ID,
					Name:      se.ID,
					PanelType: "node",
					Status:    "complete",
					Title:     sd.Name + ": " + se.ID,
					Data: SnapshotPanelData{
						Columns: []SnapshotPanelDataColumn{
							{Name: "id", DataType: "TEXT"},
							{Name: "title", DataType: "TEXT"},
							{Name: "properties", DataType: "JSONB"},
						},
						Rows: []SnapshotPanelDataRow{
							{
								"id":    se.ID,
								"title": sd.Name,
								"properties": map[string]interface{}{
									"Execution ID": se.ID,
									"Input":        se.Input,
									"Output":       se.Output,
									"Status":       se.Status,
									"Type":         sd.Type,
								},
							},
						},
					},
					Properties: map[string]interface{}{
						"name":     se.ID,
						"category": Category(sd.Type),
					},
				}

				// Edge from step start to execution
				edgeName = "start_" + sd.Name + "_to_" + se.ID
				edgeNames = append(edgeNames, edgeName)
				snapshot.Panels[edgeName] = SnapshotPanel{
					Dashboard: pe.ID,
					Name:      edgeName,
					PanelType: "edge",
					Status:    "complete",
					Data: SnapshotPanelData{
						Columns: []SnapshotPanelDataColumn{
							{Name: "from_id", DataType: "TEXT"},
							{Name: "to_id", DataType: "TEXT"},
						},
						Rows: []SnapshotPanelDataRow{
							{
								"from_id": "start_" + sd.Name,
								"to_id":   se.ID,
							},
						},
					},
					Properties: map[string]interface{}{
						"name": edgeName,
					},
				}

				// Edge from step start to execution
				edgeName = se.ID + "_to_" + sd.Name
				edgeNames = append(edgeNames, edgeName)
				snapshot.Panels[edgeName] = SnapshotPanel{
					Dashboard: pe.ID,
					Name:      edgeName,
					PanelType: "edge",
					Status:    "complete",
					Data: SnapshotPanelData{
						Columns: []SnapshotPanelDataColumn{
							{Name: "from_id", DataType: "TEXT"},
							{Name: "to_id", DataType: "TEXT"},
						},
						Rows: []SnapshotPanelDataRow{
							{
								"from_id": se.ID,
								"to_id":   sd.Name,
							},
						},
					},
					Properties: map[string]interface{}{
						"name": edgeName,
					},
				}

			}
		}

		if len(sd.DependsOn) > 0 {
			for _, dep := range sd.DependsOn {
				dependedOn[dep] = true

				// Edge between dependencies
				edgeName = dep + "_to_" + stepToID
				edgeNames = append(edgeNames, edgeName)
				snapshot.Panels[edgeName] = SnapshotPanel{
					Dashboard: pe.ID,
					Name:      edgeName,
					PanelType: "edge",
					Status:    "complete",
					Data: SnapshotPanelData{
						Columns: []SnapshotPanelDataColumn{
							{Name: "from_id", DataType: "TEXT"},
							{Name: "to_id", DataType: "TEXT"},
						},
						Rows: []SnapshotPanelDataRow{
							{
								"from_id": dep,
								"to_id":   stepToID,
							},
						},
					},
					Properties: map[string]interface{}{
						"name": edgeName,
					},
				}

			}
		} else {

			// Edge from pipeline start to step
			edgeName = "start_" + pe.ID + "_to_" + stepToID
			edgeNames = append(edgeNames, edgeName)
			snapshot.Panels[edgeName] = SnapshotPanel{
				Dashboard: pe.ID,
				Name:      edgeName,
				PanelType: "edge",
				Status:    "complete",
				Data: SnapshotPanelData{
					Columns: []SnapshotPanelDataColumn{
						{Name: "from_id", DataType: "TEXT"},
						{Name: "to_id", DataType: "TEXT"},
					},
					Rows: []SnapshotPanelDataRow{
						{
							"from_id": "start_" + pe.ID,
							"to_id":   stepToID,
						},
					},
				},
				Properties: map[string]interface{}{
					"name": edgeName,
				},
			}

		}

		tableRows = append(tableRows, SnapshotPanelDataRow{
			"name":       sd.Name,
			"type":       sd.Type,
			"depends_on": sd.DependsOn,
		})
	}

	// Add edges from steps that are not depended on to the end
	for _, sd := range pd.Steps {
		if dependedOn[sd.Name] {
			continue
		}

		// Edge from pipeline start to step
		edgeName = sd.Name + "_to_" + "end_" + pe.ID
		edgeNames = append(edgeNames, edgeName)
		snapshot.Panels[edgeName] = SnapshotPanel{
			Dashboard: pe.ID,
			Name:      edgeName,
			PanelType: "edge",
			Status:    "complete",
			Data: SnapshotPanelData{
				Columns: []SnapshotPanelDataColumn{
					{Name: "from_id", DataType: "TEXT"},
					{Name: "to_id", DataType: "TEXT"},
				},
				Rows: []SnapshotPanelDataRow{
					{
						"from_id": sd.Name,
						"to_id":   "end_" + pe.ID,
					},
				},
			},
			Properties: map[string]interface{}{
				"name": edgeName,
			},
		}

	}

	// Copy node and edge data into the graph for UI debugging
	graphColumns := []SnapshotPanelDataColumn{}
	graphColumnNames := map[string]bool{}
	graphRows := []SnapshotPanelDataRow{}
	for _, edge := range edgeNames {
		for _, col := range snapshot.Panels[edge].Data.Columns {
			if graphColumnNames[col.Name] {
				// Already added
				continue
			}
			graphColumns = append(graphColumns, col)
			graphColumnNames[col.Name] = true
		}
		graphRows = append(graphRows, snapshot.Panels[edge].Data.Rows...)
	}
	for _, node := range nodeNames {
		for _, col := range snapshot.Panels[node].Data.Columns {
			if graphColumnNames[col.Name] {
				// Already added
				continue
			}
			graphColumns = append(graphColumns, col)
			graphColumnNames[col.Name] = true
		}
		graphRows = append(graphRows, snapshot.Panels[node].Data.Rows...)
	}
	executionPanel.Data.Columns = graphColumns
	executionPanel.Data.Rows = graphRows

	tablePanel.Data.Rows = tableRows

	executionPanel.Properties["nodes"] = nodeNames
	executionPanel.Properties["edges"] = edgeNames

	// Finalize
	snapshot.EndTime = time.Now().UTC().Format(time.RFC3339)
	snapshot.Panels[pe.ID] = dashboardPanel
	snapshot.Panels["execution_tree"] = executionPanel
	snapshot.Panels["step_definitions"] = tablePanel

	return snapshot, nil
}

func Category(category string) map[string]interface{} {
	switch category {
	case "pipeline":
		return map[string]interface{}{
			"name":  "pipeline",
			"title": "Pipeline",
			"color": "green",
			"icon":  "valve",
		}
	case "sleep":
		return map[string]interface{}{
			"name":  "sleep",
			"title": "Sleep",
			"color": "grey",
			"icon":  "timer",
		}
	case "exec":
		return map[string]interface{}{
			"name":  "exec",
			"title": "Exec",
			"color": "red",
			"icon":  "terminal",
		}
	case "http_request":
		return map[string]interface{}{
			"name":  "http_request",
			"title": "HTTP Request",
			"color": "purple",
			"icon":  "http",
		}
	case "query":
		return map[string]interface{}{
			"name":  "query",
			"title": "Query",
			"color": "blue",
			"icon":  "table",
		}
	default:
		return map[string]interface{}{
			"name":  category,
			"title": category,
			"color": "red",
			"icon":  "info",
		}
	}
}
