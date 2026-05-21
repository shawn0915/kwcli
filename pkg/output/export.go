package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ExportFormat represents the export format type
type ExportFormat string

const (
	ExportCSV  ExportFormat = "csv"
	ExportJSON ExportFormat = "json"
)

// ParseExportFormat converts a string to ExportFormat
func ParseExportFormat(s string) (ExportFormat, error) {
	switch strings.ToLower(s) {
	case "csv":
		return ExportCSV, nil
	case "json":
		return ExportJSON, nil
	default:
		return "", fmt.Errorf("unsupported export format: %s (supported: csv, json)", s)
	}
}

// QueryResult holds the result of a SQL query for export
type QueryResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

// Export writes query results to a file or stdout in the specified format
func Export(result *QueryResult, format ExportFormat, outputFile string) error {
	var content string
	var err error

	switch format {
	case ExportCSV:
		content, err = ToCSV(result)
	case ExportJSON:
		content, err = ToJSON(result)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		return err
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(content), 0644)
	}

	fmt.Print(content)
	return nil
}

// ToCSV converts query result to CSV format
func ToCSV(result *QueryResult) (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Write header
	if err := writer.Write(result.Columns); err != nil {
		return "", err
	}

	// Write rows
	for _, row := range result.Rows {
		record := make([]string, len(row))
		for i, val := range row {
			record[i] = fmt.Sprintf("%v", val)
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	writer.Flush()
	return sb.String(), writer.Error()
}

// ToJSON converts query result to JSON format
func ToJSON(result *QueryResult) (string, error) {
	// Build array of objects
	rows := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		entry := make(map[string]interface{})
		for j, col := range result.Columns {
			if j < len(row) {
				entry[col] = row[j]
			}
		}
		rows[i] = entry
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

// ParseTableOutput parses text-based table output from kwbase into a QueryResult
// It handles tab-separated output:
//
//	column1\tcolumn2\tcolumn3
//	val1\tval2\tval3
//
// and pipe-separated output:
//
//	column1 | column2 | column3
//	--------+---------+--------
//	val1    | val2    | val3
func ParseTableOutput(text string) *QueryResult {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) < 2 {
		return &QueryResult{Columns: []string{}, Rows: [][]interface{}{}}
	}

	// Detect separator: tab or pipe
	headerLine := strings.TrimSpace(lines[0])
	var columns []string
	var separator string
	if strings.Contains(headerLine, "\t") {
		separator = "\t"
		columns = strings.Split(headerLine, "\t")
		for i := range columns {
			columns[i] = strings.TrimSpace(columns[i])
		}
	} else if strings.Contains(headerLine, "|") {
		separator = "|"
		columns = splitTableRow(headerLine)
	} else {
		return &QueryResult{Columns: []string{}, Rows: [][]interface{}{}}
	}

	result := &QueryResult{
		Columns: columns,
		Rows:    [][]interface{}{},
	}

	startIdx := 1
	// Skip separator line for pipe format (--------+---------+--------)
	if separator == "|" && len(lines) > 1 && strings.HasPrefix(strings.TrimSpace(lines[1]), "-") {
		startIdx = 2
	}

	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		// Skip lines that look like row counts "(N rows)"
		if strings.HasPrefix(line, "(") && strings.HasSuffix(line, ")") {
			continue
		}
		var values []string
		if separator == "\t" {
			values = strings.Split(line, "\t")
		} else {
			values = splitTableRow(line)
		}
		row := make([]interface{}, len(values))
		for j, v := range values {
			row[j] = strings.TrimSpace(v)
		}
		result.Rows = append(result.Rows, row)
	}

	return result
}

// splitTableRow splits a table row by | separator
func splitTableRow(line string) []string {
	parts := strings.Split(line, "|")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimSpace(p)
	}
	return result
}
