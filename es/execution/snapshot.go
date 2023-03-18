package execution

import (
	"fmt"
	"strconv"
	"time"

	"github.com/turbot/steampipe-pipelines/pipeline"
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
		Properties: map[string]interface{}{
			"name":      "execution",
			"direction": "TD",
		},
	}

	// Cacluate the rows
	var edgeName string
	nodeNames := []string{}
	edgeNames := []string{}

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
						"Args":         pe.Args,
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
						"Output":       pe.Output,
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

		stepPanels, err := ex.StepExecutionSnapshotPanels(pe.ID, sd.Name)
		if err != nil {
			return nil, err
		}

		for panelName, panel := range stepPanels {
			// Add the panels to the dashboard
			snapshot.Panels[panelName] = panel
			// Gather panel names for the graph details
			if panel.PanelType == "node" {
				nodeNames = append(nodeNames, panelName)
			} else if panel.PanelType == "edge" {
				edgeNames = append(edgeNames, panelName)
			}
		}

		stepToID := "step_" + sd.Name
		if len(stepPanels) > 1 {
			stepToID = "stepstart_" + sd.Name
		}

		if len(sd.DependsOn) > 0 {

			// Build edges from dependencies to this step

			for _, dep := range sd.DependsOn {
				dependedOn[dep] = true

				edgeName = "edge_" + dep + "_to_" + stepToID
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
								"from_id": "step_" + dep,
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
			edgeName = "edge_" + "start_" + pe.ID + "_to_" + stepToID
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

	}

	// Add edges from steps that are not depended on to the end
	for _, sd := range pd.Steps {
		if dependedOn[sd.Name] {
			continue
		}

		// Edge from pipeline start to step
		edgeName = "edge_" + "step_" + sd.Name + "_to_" + "end_" + pe.ID
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
						"from_id": "step_" + sd.Name,
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

	executionPanel.Properties["nodes"] = nodeNames
	executionPanel.Properties["edges"] = edgeNames

	// Finalize
	snapshot.EndTime = time.Now().UTC().Format(time.RFC3339)
	snapshot.Panels[pe.ID] = dashboardPanel
	snapshot.Panels["execution_tree"] = executionPanel

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
			"icon":  "priority_high",
		}
	}
}

// StepExecutionSnapshotPanels will build and return a set of panels to represent
// the step execution in a dashboard. The panels includes both nodes and edges,
// depending on the exection of the step - for example, if the step has a for loop
// then it will be a start node, fan out to loop items and collapse back to an end node.
func (ex *Execution) StepExecutionSnapshotPanels(pipelineExecutionID string, stepName string) (map[string]SnapshotPanel, error) {

	// Get the pipeline execution for this step
	pe := ex.PipelineExecutions[pipelineExecutionID]

	// Get the pipeline definition that contains the step
	pd, err := ex.PipelineDefinition(pe.ID)
	if err != nil {
		return nil, err
	}

	sd, ok := pd.Steps[stepName]
	if !ok {
		return nil, fmt.Errorf("step %s not found in pipeline %s", stepName, pd.Name)
	}

	panels := map[string]SnapshotPanel{}

	stepExecutions := ex.PipelineStepExecutions(pe.ID, sd.Name)

	if len(stepExecutions) <= 0 {

		// The step has no actual executions, so should just show a "skipped"
		// style of node as a placeholder.

		// Single node, so panel name should match the step name with standard
		// prefix.
		panelName := "step_" + sd.Name

		panels[panelName] = SnapshotPanel{
			Dashboard: pe.ID,
			Name:      panelName,
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
						"id":    panelName,
						"title": sd.Name,
						"properties": map[string]interface{}{
							"Type":   sd.Type,
							"Status": "Skipped - no executions",
						},
					},
				},
			},
			Properties: map[string]interface{}{
				"name":     panelName,
				"category": Category(sd.Type),
			},
		}

		return panels, nil

	}

	if len(stepExecutions) == 1 {

		// The step has just a single execution (i.e. no for loop). We return
		// a single node in this case.

		// Convenience
		se := stepExecutions[0]

		// Single node, so panel name should match the step name with standard
		// prefix.
		panelName := "step_" + sd.Name

		panels[panelName] = SnapshotPanel{
			Dashboard: pe.ID,
			Name:      panelName,
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
					ex.StepExecutionNodeRow(panelName, sd, se),
				},
			},
			Properties: map[string]interface{}{
				"name":     panelName,
				"category": Category(sd.Type),
			},
		}

		return panels, nil

	}

	// Multiple executions, e.g. a for loop. Add a start node for the step,
	// execution nodes in the middle, all converging back into an end node.

	// Start of step
	startNodeName := "stepstart_" + sd.Name
	panels[startNodeName] = SnapshotPanel{
		Dashboard: pe.ID,
		Name:      startNodeName,
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
					"id":    startNodeName,
					"title": "Start: " + sd.Name,
					"properties": map[string]interface{}{
						"Type": sd.Type,
					},
				},
			},
		},
		Properties: map[string]interface{}{
			"name":     startNodeName,
			"category": Category(sd.Type),
		},
	}

	// End of step
	endNodeName := "step_" + sd.Name
	panels[endNodeName] = SnapshotPanel{
		Dashboard: pe.ID,
		Name:      endNodeName,
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
					"id":    endNodeName,
					"title": "End: " + sd.Name,
					"properties": map[string]interface{}{
						"Type": sd.Type,
					},
				},
			},
		},
		Properties: map[string]interface{}{
			"name":     endNodeName,
			"category": Category(sd.Type),
		},
	}

	// Node for the step execution
	nodeName := "exec_" + sd.Name
	nodePanel := SnapshotPanel{
		Dashboard: pe.ID,
		Name:      nodeName,
		PanelType: "node",
		Status:    "complete",
		Title:     sd.Name,
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "id", DataType: "TEXT"},
				{Name: "title", DataType: "TEXT"},
				{Name: "properties", DataType: "JSONB"},
			},
			Rows: []SnapshotPanelDataRow{},
		},
		Properties: map[string]interface{}{
			"name":     nodeName,
			"category": Category(sd.Type),
		},
	}

	// Edge from step start to execution
	startEdgeName := "edge_" + startNodeName + "_to_" + nodeName
	startEdgePanel := SnapshotPanel{
		Dashboard: pe.ID,
		Name:      startEdgeName,
		PanelType: "edge",
		Status:    "complete",
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "from_id", DataType: "TEXT"},
				{Name: "to_id", DataType: "TEXT"},
			},
			Rows: []SnapshotPanelDataRow{},
		},
		Properties: map[string]interface{}{
			"name": startEdgeName,
		},
	}

	// Edge from step start to execution
	endEdgeName := "edge_" + nodeName + "_to_" + endNodeName
	endEdgePanel := SnapshotPanel{
		Dashboard: pe.ID,
		Name:      endEdgeName,
		PanelType: "edge",
		Status:    "complete",
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "from_id", DataType: "TEXT"},
				{Name: "to_id", DataType: "TEXT"},
			},
			Rows: []SnapshotPanelDataRow{},
		},
		Properties: map[string]interface{}{
			"name": endEdgeName,
		},
	}

	// Add row data for each execution to the execution node and its edges
	for _, se := range stepExecutions {
		nodePanel.Data.Rows = append(nodePanel.Data.Rows, ex.StepExecutionNodeRow(se.ID, sd, se))
		startEdgePanel.Data.Rows = append(startEdgePanel.Data.Rows, SnapshotPanelDataRow{
			"from_id": "stepstart_" + sd.Name,
			"to_id":   se.ID,
		})
		endEdgePanel.Data.Rows = append(endEdgePanel.Data.Rows, SnapshotPanelDataRow{
			"from_id": se.ID,
			"to_id":   "step_" + sd.Name,
		})
	}

	panels[nodeName] = nodePanel
	panels[startEdgeName] = startEdgePanel
	panels[endEdgeName] = endEdgePanel

	return panels, nil
}

func (ex *Execution) StepExecutionNodeRow(panelName string, sd *pipeline.PipelineStep, se *StepExecution) SnapshotPanelDataRow {

	var row SnapshotPanelDataRow

	var title string
	if se.ForEach != nil {
		// TODO - we can do better than this I think?
		switch k := (*se.ForEach)["key"].(type) {
		case int:
			// Don't include integer keys in title
		case string:
			title = k + " = "
		}
		switch v := (*se.ForEach)["value"].(type) {
		case string:
			title += v
		case int:
			title += strconv.Itoa(v)
		}
	}
	if title == "" {
		title = sd.Name + " [" + se.ID[len(se.ID)-4:] + "]"
	}

	switch sd.Type {

	case "sleep":
		row = SnapshotPanelDataRow{
			"id":    panelName,
			"title": se.Input["duration"],
			"properties": map[string]interface{}{
				"Execution ID": se.ID,
				"Status":       se.Status,
				"Duration":     se.Input["duration"],
				"Started At":   se.Output.Get("started_at"),
				"Finished At":  se.Output.Get("finished_at"),
				"For Each":     se.ForEach,
			},
		}

	case "http_request":
		row = SnapshotPanelDataRow{
			"id":    panelName,
			"title": se.Input["url"],
			"properties": map[string]interface{}{
				"Execution ID":         se.ID,
				"Status":               se.Status,
				"URL":                  se.Input["url"],
				"Response Status Code": se.Output.Get("status_code"),
				"Started At":           se.Output.Get("started_at"),
				"Finished At":          se.Output.Get("finished_at"),
				"For Each":             se.ForEach,
			},
		}

	case "query":
		row = SnapshotPanelDataRow{
			"id":    panelName,
			"title": title,
			"properties": map[string]interface{}{
				"Execution ID": se.ID,
				"Status":       se.Status,
				"Row Count":    len(se.Output.Get("rows").([]interface{})),
				"Started At":   se.Output.Get("started_at"),
				"Finished At":  se.Output.Get("finished_at"),
				"For Each":     se.ForEach,
			},
		}

	default:
		row = SnapshotPanelDataRow{
			"id":    panelName,
			"title": title,
			"properties": map[string]interface{}{
				"Execution ID": se.ID,
				"Status":       se.Status,
				"Input":        se.Input,
				"For Each":     se.ForEach,
				"Output":       se.Output,
			},
		}
	}

	return row
}
