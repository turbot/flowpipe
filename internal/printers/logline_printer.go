package printers

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/types"
	"io"
)

type LogLinePrinter struct {
}

func (p LogLinePrinter) PrintResource(ctx context.Context, r types.PrintableResource, writer io.Writer) error {
	lines := r.GetItems().([]types.LogLine)

	for _, line := range lines {
		msg := buildLogLinePrefix(line)
		msg += line.Message
		msg += "\n"

		_, err := writer.Write([]byte(msg))
		if err != nil {
			return err
		}
	}

	return nil
}

func buildLogLinePrefix(l types.LogLine) string {
	out := fmt.Sprintf("[%s", l.Name)
	if l.StepName != nil {
		out += fmt.Sprintf(".%s", *l.StepName)
	}
	if l.ForEachKey != nil {
		out += fmt.Sprintf("[%s]", *l.ForEachKey)
	}
	if l.LoopIndex != nil {
		out += fmt.Sprintf("[%d]", *l.LoopIndex)
	}
	if l.RetryIndex != nil {
		out += fmt.Sprintf("#%d", *l.RetryIndex)
	}
	out += "] "
	return out
}
