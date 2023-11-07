package execution

import (
	"fmt"
	"time"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
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
		return nil, perr.BadRequestWithMessage(fmt.Sprintf("pipeline execution %s not found", pipelineExecutionID))
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
		Title:     "Start: " + pd.Name(),
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "id", DataType: "TEXT"},
				{Name: "title", DataType: "TEXT"},
				{Name: "properties", DataType: "JSONB"},
			},
			Rows: []SnapshotPanelDataRow{
				{
					"id":    "start_" + pe.ID,
					"title": "Start: " + pd.Name(),
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
		Title:     "End: " + pd.Name(),
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "id", DataType: "TEXT"},
				{Name: "title", DataType: "TEXT"},
				{Name: "properties", DataType: "JSONB"},
			},
			Rows: []SnapshotPanelDataRow{
				{
					"id":    "end_" + pe.ID,
					"title": "End: " + pd.Name(),
					"properties": map[string]interface{}{
						"Execution ID": pe.ID,
						"Output":       pe.PipelineOutput,
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

		stepPanels, err := ex.StepExecutionSnapshotPanels(pe.ID, sd.GetFullyQualifiedName())
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

		stepToID := "step_" + sd.GetFullyQualifiedName()
		if len(stepPanels) > 1 {
			stepToID = "stepstart_" + sd.GetFullyQualifiedName()
		}

		if len(sd.GetDependsOn()) > 0 {

			// Build edges from dependencies to this step

			for _, dep := range sd.GetDependsOn() {
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
		if dependedOn[sd.GetFullyQualifiedName()] {
			continue
		}

		// Edge from pipeline start to step
		edgeName = "edge_" + "step_" + sd.GetFullyQualifiedName() + "_to_" + "end_" + pe.ID
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
						"from_id": "step_" + sd.GetFullyQualifiedName(),
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
	case schema.BlockTypePipelineStepSleep:
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
	case schema.BlockTypePipelineStepHttp:
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
	case "echo":
		return map[string]interface{}{
			"name":  "text",
			"title": "Text",
			"color": "blue",
			"icon":  "description",
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
func (ex *Execution) StepExecutionSnapshotPanels(pipelineExecutionID string, stepFullyQualifiedName string) (map[string]SnapshotPanel, error) {

	// Get the pipeline execution for this step
	pe := ex.PipelineExecutions[pipelineExecutionID]

	// Get the pipeline definition that contains the step
	pd, err := ex.PipelineDefinition(pe.ID)
	if err != nil {
		return nil, err
	}

	sd := pd.GetStep(stepFullyQualifiedName)
	if sd == nil {
		return nil, perr.BadRequestWithMessage("step " + stepFullyQualifiedName + " not found in pipeline " + pd.Name())
	}

	panels := map[string]SnapshotPanel{}

	stepExecutions := ex.PipelineStepExecutions(pe.ID, sd.GetFullyQualifiedName())

	if len(stepExecutions) == 0 {

		// The step has no actual executions, so should just show a "skipped"
		// style of node as a placeholder.

		// Single node, so panel name should match the step name with standard
		// prefix.
		panelName := "step_" + sd.GetFullyQualifiedName()

		panels[panelName] = SnapshotPanel{
			Dashboard: pe.ID,
			Name:      panelName,
			PanelType: "node",
			Status:    "complete",
			Title:     sd.GetFullyQualifiedName(),
			Data: SnapshotPanelData{
				Columns: []SnapshotPanelDataColumn{
					{Name: "id", DataType: "TEXT"},
					{Name: "title", DataType: "TEXT"},
					{Name: "properties", DataType: "JSONB"},
				},
				Rows: []SnapshotPanelDataRow{
					{
						"id":    panelName,
						"title": sd.GetFullyQualifiedName(),
						"properties": map[string]interface{}{
							"Type":   sd.GetType(),
							"Status": "Skipped - no executions",
						},
					},
				},
			},
			Properties: map[string]interface{}{
				"name":     panelName,
				"category": Category(sd.GetType()),
			},
		}
		panels[panelName].Properties["category"].(map[string]interface{})["color"] = "grey"

		return panels, nil

	}

	if len(stepExecutions) == 1 {

		// The step has just a single execution (i.e. no for loop). We return
		// a single node in this case.

		// Convenience
		se := stepExecutions[0]

		// Single node, so panel name should match the step name with standard
		// prefix.
		panelName := "step_" + sd.GetFullyQualifiedName()

		panels[panelName] = SnapshotPanel{
			Dashboard: pe.ID,
			Name:      panelName,
			PanelType: "node",
			Status:    "complete",
			Title:     sd.GetFullyQualifiedName(),
			Data: SnapshotPanelData{
				Columns: []SnapshotPanelDataColumn{
					{Name: "id", DataType: "TEXT"},
					{Name: "title", DataType: "TEXT"},
					{Name: "properties", DataType: "JSONB"},
				},
				Rows: []SnapshotPanelDataRow{
					ex.StepExecutionNodeRow(panelName, sd, &se),
				},
			},
			Properties: map[string]interface{}{
				"name":     panelName,
				"category": Category(sd.GetType()),
			},
		}

		if se.Status == "skipped" {
			panels[panelName].Properties["category"].(map[string]interface{})["color"] = "grey"
		}

		return panels, nil

	}

	// Multiple executions, e.g. a for loop. Add a start node for the step,
	// execution nodes in the middle, all converging back into an end node.

	// Start of step
	startNodeName := "stepstart_" + sd.GetFullyQualifiedName()
	panels[startNodeName] = SnapshotPanel{
		Dashboard: pe.ID,
		Name:      startNodeName,
		PanelType: "node",
		Status:    "complete",
		Title:     "Start: " + sd.GetFullyQualifiedName(),
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "id", DataType: "TEXT"},
				{Name: "title", DataType: "TEXT"},
				{Name: "properties", DataType: "JSONB"},
			},
			Rows: []SnapshotPanelDataRow{
				{
					"id":    startNodeName,
					"title": "Start: " + sd.GetFullyQualifiedName(),
					"properties": map[string]interface{}{
						"Type": sd.GetType(),
					},
				},
			},
		},
		Properties: map[string]interface{}{
			"name":     startNodeName,
			"category": Category(sd.GetType()),
		},
	}

	// End of step
	endNodeName := "step_" + sd.GetFullyQualifiedName()
	panels[endNodeName] = SnapshotPanel{
		Dashboard: pe.ID,
		Name:      endNodeName,
		PanelType: "node",
		Status:    "complete",
		Title:     "End: " + sd.GetFullyQualifiedName(),
		Data: SnapshotPanelData{
			Columns: []SnapshotPanelDataColumn{
				{Name: "id", DataType: "TEXT"},
				{Name: "title", DataType: "TEXT"},
				{Name: "properties", DataType: "JSONB"},
			},
			Rows: []SnapshotPanelDataRow{
				{
					"id":    endNodeName,
					"title": "End: " + sd.GetFullyQualifiedName(),
					"properties": map[string]interface{}{
						"Type": sd.GetType(),
					},
				},
			},
		},
		Properties: map[string]interface{}{
			"name":     endNodeName,
			"category": Category(sd.GetType()),
		},
	}

	// Node for the step execution
	nodeName := "exec_" + sd.GetFullyQualifiedName()
	nodePanel := SnapshotPanel{
		Dashboard: pe.ID,
		Name:      nodeName,
		PanelType: "node",
		Status:    "complete",
		Title:     sd.GetFullyQualifiedName(),
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
			"category": Category(sd.GetType()),
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

	// TODO: I don't know how to do multiple executions with different colour (grey for skipped). Seems like there's only 1 nodePanel for all data rows??

	// Add row data for each execution to the execution node and its edges
	for i, se := range stepExecutions {
		nodePanel.Data.Rows = append(nodePanel.Data.Rows, ex.StepExecutionNodeRow(se.ID, sd, &stepExecutions[i]))

		startEdgePanel.Data.Rows = append(startEdgePanel.Data.Rows, SnapshotPanelDataRow{
			"from_id": "stepstart_" + sd.GetFullyQualifiedName(),
			"to_id":   se.ID,
		})
		endEdgePanel.Data.Rows = append(endEdgePanel.Data.Rows, SnapshotPanelDataRow{
			"from_id": se.ID,
			"to_id":   "step_" + sd.GetFullyQualifiedName(),
		})
	}

	panels[nodeName] = nodePanel
	panels[startEdgeName] = startEdgePanel
	panels[endEdgeName] = endEdgePanel

	return panels, nil
}

func (ex *Execution) StepExecutionNodeRow(panelName string, sd modconfig.PipelineStep, se *StepExecution) SnapshotPanelDataRow {

	var row SnapshotPanelDataRow

	var title string

	if se.StepForEach != nil {
		title = se.StepForEach.Key + " = "

		// TODO: this is a bit yuck
		// forEachOutput, ok := se.StepForEach.Output.Get(schema.AttributeTypeValue).(string)
		// if !ok {
		// 	title += sd.GetFullyQualifiedName()
		// } else {
		// 	title += forEachOutput
		// }

	}
	if title == "" {
		title = sd.GetFullyQualifiedName() + " [" + se.ID[len(se.ID)-4:] + "]"
	}

	if se.Status == "skipped" {
		title = "[skipped] " + title
	}

	switch sd.GetType() {

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
				// "For Each":     se.ForEach,
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
				// "For Each":             se.ForEach,
			},
		}

	case "query":

		rowCount := 0
		if se.Output.Get("rows") != nil {
			rows, ok := se.Output.Get("rows").([]interface{})
			if ok {
				rowCount = len(rows)
			}
		}

		row = SnapshotPanelDataRow{
			"id":    panelName,
			"title": title,
			"properties": map[string]interface{}{
				"Execution ID": se.ID,
				"Status":       se.Status,
				"Row Count":    rowCount,
				"Started At":   se.Output.Get("started_at"),
				"Finished At":  se.Output.Get("finished_at"),
				// "For Each":     se.ForEach,
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
				// "For Each":     se.ForEach,
				"Output": se.Output,
			},
		}
	}

	return row
}
