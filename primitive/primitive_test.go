package primitive

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrimitiveInputOutput(t *testing.T) {
	fmt.Println("TestPrimitiveInputOutput")
	assert := assert.New(t)
	p := primitive{}
	input := Input(map[string]interface{}{"foo": "bar"})
	err := p.SetInput(input)
	if err != nil {
		t.Errorf("SetInput: %v", err)
	}
	output := p.Input()
	assert.Equal(input, output, "Input() should return value set by SetInput()")
}
