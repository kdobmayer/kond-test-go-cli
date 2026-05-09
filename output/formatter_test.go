package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestFormatter_Table(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("table", &buf)

	headers := []string{"NAME", "VALUE"}
	rows := []TableRow{
		{Columns: []string{"foo", "bar"}},
		{Columns: []string{"baz", "qux"}},
	}

	if err := f.Table(headers, rows); err != nil {
		t.Fatalf("Table() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Error("output should contain header NAME")
	}
	if !strings.Contains(out, "foo") {
		t.Error("output should contain 'foo'")
	}
	if !strings.Contains(out, "baz") {
		t.Error("output should contain 'baz'")
	}
}

func TestFormatter_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("json", &buf)

	data := map[string]string{"key": "value"}
	if err := f.JSON(data); err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("result[key] = %q, want %q", result["key"], "value")
	}
}

func TestFormatter_YAML(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("yaml", &buf)

	data := map[string]string{"key": "value"}
	if err := f.YAML(data); err != nil {
		t.Fatalf("YAML() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "key: value") {
		t.Errorf("output = %q, expected to contain 'key: value'", out)
	}
}

func TestFormatter_Render_Table(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("table", &buf)

	headers := []string{"A", "B"}
	rows := []TableRow{{Columns: []string{"1", "2"}}}
	data := []map[string]string{{"a": "1", "b": "2"}}

	if err := f.Render(headers, rows, data); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(buf.String(), "A") {
		t.Error("table render should contain header")
	}
}

func TestFormatter_Render_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("json", &buf)

	headers := []string{"A"}
	rows := []TableRow{{Columns: []string{"1"}}}
	data := []map[string]string{{"a": "1"}}

	if err := f.Render(headers, rows, data); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(buf.String(), `"a"`) {
		t.Error("json render should contain key")
	}
}

func TestFormatter_Render_YAML(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("yaml", &buf)

	headers := []string{"A"}
	rows := []TableRow{{Columns: []string{"1"}}}
	data := map[string]string{"a": "1"}

	if err := f.Render(headers, rows, data); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(buf.String(), "a: \"1\"") {
		t.Errorf("yaml render output = %q", buf.String())
	}
}

func TestFormatter_RenderMessage(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("table", &buf)
	f.RenderMessage("hello world")

	if !strings.Contains(buf.String(), "hello world") {
		t.Error("RenderMessage should output the message")
	}
}

func TestFormatter_RenderError(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("table", &buf)
	f.RenderError(fmt.Errorf("something broke"))

	if !strings.Contains(buf.String(), "something broke") {
		t.Error("RenderError should output the error")
	}
}
