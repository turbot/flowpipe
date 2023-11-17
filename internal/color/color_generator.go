package color

import "fmt"

const minColumn = 16
const maxColumn = 51
const minRow = 0
const maxRow = 5

type DynamicColorGenerator struct {
	registry       map[string]uint8
	startingRow    uint8
	startingColumn uint8

	// state
	allocatedColorCodes []uint8
	forbiddenColumns    map[uint8]bool
	currentRow          uint8
	currentColumn       uint8
}

// NewDynamicColorGenerator creates a new DynamicColorGenerator
func NewDynamicColorGenerator(startingRow, startingColumn uint8) (*DynamicColorGenerator, error) {
	forbiddenColumns := map[uint8]bool{
		16: true, 17: true, 18: true, 19: true, 20: true, // red
		22: true, 23: true, 27: true, 28: true, 29: true, // orange
		34: true, 35: true, 36: true, 40: true, 41: true, 42: true, // green/orange
		46: true, 47: true, 48: true, 49: true, // green
	}
	if startingColumn < minColumn || startingColumn > maxColumn {
		return nil, fmt.Errorf("starting column must be between 16 and 51")
	}
	if startingRow < minRow || startingRow > maxRow {
		return nil, fmt.Errorf("starting row must be between 0 and 5")
	}

	g := &DynamicColorGenerator{
		registry:         make(map[string]uint8),
		startingRow:      startingRow,
		startingColumn:   startingColumn,
		forbiddenColumns: forbiddenColumns,
	}
	g.reset()
	return g, nil
}

func (g *DynamicColorGenerator) GetColorForElement(key string) uint8 {
	if value, exists := g.registry[key]; exists {
		return value
	} else {
		c := g.getNextColor()
		g.registry[key] = c
		return c
	}
}

func (g *DynamicColorGenerator) reset() {
	// create the state map
	g.currentRow = g.startingRow
	g.currentColumn = g.startingColumn
	// clear allocated colors
	g.allocatedColorCodes = nil
}

func (g *DynamicColorGenerator) incrementCurrentRow(increment uint8) {
	g.currentRow += increment
	if g.currentRow > maxRow {
		g.currentRow -= maxRow
	}
}

func (g *DynamicColorGenerator) incrementCurrentColumn(increment uint8) {
	g.currentColumn += increment
	if g.currentColumn > maxColumn {
		// reset to 16
		g.currentColumn -= maxColumn - minColumn + 1
	}
	for ; g.forbiddenColumns[g.currentColumn]; g.currentColumn++ {
	}
}

func (g *DynamicColorGenerator) getCurrentColor() uint8 {
	return g.currentColumn + g.currentRow*36
}

func (g *DynamicColorGenerator) getNextColor() uint8 {
	g.incrementCurrentColumn(2)
	g.incrementCurrentRow(2)

	// does this color clash, or is it forbidden
	color := g.getCurrentColor()
	origColor := color
	for g.colorClashes(color) {
		g.incrementCurrentColumn(1)
		g.incrementCurrentRow(1)
		color = g.getCurrentColor()
		if color == origColor {
			// we have tried them all reset and start from the first color
			g.reset()
			return g.getNextColor()
		}
	}

	// store this color code
	g.allocatedColorCodes = append(g.allocatedColorCodes, color)
	return color
}

func (g *DynamicColorGenerator) colorClashes(color uint8) bool {
	for _, a := range g.allocatedColorCodes {
		if a == color {
			return true
		}
	}

	return false
}
