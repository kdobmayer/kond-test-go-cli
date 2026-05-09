package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Formatter handles output formatting
type Formatter struct {
	Format string
	Writer io.Writer
}

// NewFormatter creates a new formatter
func NewFormatter(format string, w io.Writer) *Formatter {
	return &Formatter{Format: format, Writer: w}
}

// TableRow represents a row in table output
type TableRow struct {
	Columns []string
}

// Table outputs data as a formatted table
func (f *Formatter) Table(headers []string, rows []TableRow) error {
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Repeat("-\t", len(headers)))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row.Columns, "\t"))
	}
	return w.Flush()
}

// JSON outputs data as JSON
func (f *Formatter) JSON(data interface{}) error {
	enc := json.NewEncoder(f.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// YAML outputs data as YAML
func (f *Formatter) YAML(data interface{}) error {
	enc := yaml.NewEncoder(f.Writer)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(data)
}

// Render outputs data in the configured format
func (f *Formatter) Render(headers []string, rows []TableRow, data interface{}) error {
	switch f.Format {
	case "json":
		return f.JSON(data)
	case "yaml":
		return f.YAML(data)
	default:
		return f.Table(headers, rows)
	}
}

// RenderMessage outputs a simple message (for non-data responses)
func (f *Formatter) RenderMessage(msg string) {
	fmt.Fprintln(f.Writer, msg)
}

// RenderError outputs an error message
func (f *Formatter) RenderError(err error) {
	fmt.Fprintf(f.Writer, "Error: %v\n", err)
}
