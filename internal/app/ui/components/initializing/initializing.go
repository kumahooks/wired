// Package initializing implements the initialization view, a simple renderer that shows initialization steps in case
// there is an error, or libraries are missing.
package initializing

import (
	"fmt"
	"strings"
)

// Model holds the init view state: the log lines rendered in the view.
type Model struct {
	logLines []string
}

// New returns an empty Model.
func New() *Model {
	return &Model{}
}

// AppendLog adds a line to the log buffer shown in the view.
func (model *Model) AppendLog(line string) {
	model.logLines = append(model.logLines, line)
}

// LogLines returns the accumulated log lines.
func (model *Model) LogLines() []string {
	return model.logLines
}

// Render returns the initialization screen view.
// TODO: this will be a better dialog eventually.
func (model *Model) Render() string {
	var builder strings.Builder

	builder.WriteString("wire(d) is starting...\n\n")

	if len(model.logLines) == 0 {
		builder.WriteString("loading configuration...\n")
	} else {
		for _, line := range model.logLines {
			builder.WriteString(fmt.Sprintf("  %s\n", line))
		}
	}

	return builder.String()
}
