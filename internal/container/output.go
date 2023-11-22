package container

import (
	"encoding/binary"
	"io"
)

const (
	StdoutType = "stdout"
	StderrType = "stderr"
)

type Output struct {
	Lines []OutputLine
}

type OutputLine struct {
	Stream string `json:"stream"`
	Line   string `json:"line"`
}

func NewOutput() *Output {
	return &Output{Lines: []OutputLine{}}
}

// FromDockerLogsReader reads the output from a docker logs reader
// and populates the Output struct. Docker logs inject a control character
// at the start of each line to indicate if it's stdout or stderr.
//
// For example, the bytes output:
// [1 0 0 0 0 0 0 2 123 10 1 0 0 0 0 0 0 14 32 32 32 32 34 86 112 99 115 34 58 32 91 10
//
// which corresponds to the stdout:
// 1 0 0 0 0 0 0  2 {
// 1 0 0 0 0 0 0 14     "Vpcs": [
//
// Specifically it has "1 0 0 0 w x y z" as the start of each line, indicating
// stdout (1) or stderr (2) and with the last 4 bytes being the length of the
// payload.
//
// See https://github.com/moby/moby/issues/7375#issuecomment-51462963
//
// This function will read that input into our Output struct so we can choose
// the format we want later.
func (o *Output) FromDockerLogsReader(reader io.Reader) error {
	header := make([]byte, 8)

	for {
		_, err := reader.Read(header)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		streamType := header[0]
		payloadSize := binary.BigEndian.Uint32(header[4:8])
		payload := make([]byte, payloadSize)

		_, err = io.ReadFull(reader, payload)
		if err != nil {
			return err
		}

		switch streamType {
		case 2:
			o.Lines = append(o.Lines, OutputLine{Stream: StderrType, Line: string(payload)})
		default:
			o.Lines = append(o.Lines, OutputLine{Stream: StdoutType, Line: string(payload)})
		}
	}

	return nil

}

// Combined returns the combined stdout and stderr output as a single string.
func (o *Output) Combined() string {
	txt := ""
	for _, line := range o.Lines {
		txt += line.Line
	}
	return txt
}

// Stdout returns the stdout output as a single string.
func (o *Output) Stdout() string {
	txt := ""
	for _, line := range o.Lines {
		if line.Stream == StdoutType {
			txt += line.Line
		}
	}
	return txt
}

// Stderr returns the stderr output as a single string.
func (o *Output) Stderr() string {
	txt := ""
	for _, line := range o.Lines {
		if line.Stream == StderrType {
			txt += line.Line
		}
	}
	return txt
}
