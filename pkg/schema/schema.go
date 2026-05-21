package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/shawn0915/kwcli/pkg/sampledb"
)

// SchemaInfo represents the schema of a database
type SchemaInfo struct {
	Database string      `json:"database"`
	Tables   []TableInfo `json:"tables"`
	DDL      string      `json:"ddl,omitempty"`
}

// TableInfo represents a table's schema information
type TableInfo struct {
	Name    string       `json:"name"`
	Schema  string       `json:"schema"`
	Columns []ColumnInfo `json:"columns"`
	Indexes []IndexInfo  `json:"indexes,omitempty"`
	Tags    []TagInfo    `json:"tags,omitempty"`
}

// ColumnInfo represents a column in a table
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default,omitempty"`
}

// IndexInfo represents an index on a table
type IndexInfo struct {
	Name    string `json:"name"`
	Columns string `json:"columns"`
	Unique  bool   `json:"unique"`
}

// TagInfo represents a tag on a time-series table
type TagInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// DumpSchema generates schema information for a given database
func DumpSchema(dbName string) (*SchemaInfo, error) {
	switch strings.ToLower(dbName) {
	case "iot_db", "rdb":
		return dumpRDBSchema()
	case "tsdb":
		return dumpTSDBSchema()
	case "all":
		rdb, err := dumpRDBSchema()
		if err != nil {
			return nil, err
		}
		tsdb, err := dumpTSDBSchema()
		if err != nil {
			return nil, err
		}
		// Combine both
		rdb.Tables = append(rdb.Tables, tsdb.Tables...)
		rdb.DDL = sampledb.RDBSchema + "\n" + sampledb.TSDBSchema
		return rdb, nil
	default:
		return nil, fmt.Errorf("unsupported database: %s (supported: rdb, tsdb, all)", dbName)
	}
}

func dumpRDBSchema() (*SchemaInfo, error) {
	info := &SchemaInfo{
		Database: "rdb",
		DDL:      sampledb.RDBSchema,
		Tables: []TableInfo{
			{
				Name:   "meter_info",
				Schema: "rdb",
				Columns: []ColumnInfo{
					{Name: "meter_id", Type: "VARCHAR(50)", Nullable: false},
					{Name: "install_date", Type: "DATE", Nullable: true},
					{Name: "voltage_level", Type: "VARCHAR(20)", Nullable: true},
					{Name: "manufacturer", Type: "VARCHAR(50)", Nullable: true},
					{Name: "status", Type: "VARCHAR(20)", Nullable: true},
					{Name: "area_id", Type: "VARCHAR(20)", Nullable: true},
					{Name: "user_id", Type: "VARCHAR(50)", Nullable: true},
				},
			},
			{
				Name:   "user_info",
				Schema: "rdb",
				Columns: []ColumnInfo{
					{Name: "user_id", Type: "VARCHAR(50)", Nullable: false},
					{Name: "user_name", Type: "VARCHAR(100)", Nullable: true},
					{Name: "address", Type: "VARCHAR(200)", Nullable: true},
					{Name: "contact", Type: "VARCHAR(20)", Nullable: true},
				},
			},
			{
				Name:   "area_info",
				Schema: "rdb",
				Columns: []ColumnInfo{
					{Name: "area_id", Type: "VARCHAR(20)", Nullable: false},
					{Name: "area_name", Type: "VARCHAR(100)", Nullable: true},
					{Name: "manager", Type: "VARCHAR(50)", Nullable: true},
					{Name: "region", Type: "VARCHAR(50)", Nullable: true},
				},
			},
			{
				Name:   "alarm_rules",
				Schema: "rdb",
				Columns: []ColumnInfo{
					{Name: "rule_id", Type: "SERIAL", Nullable: false},
					{Name: "rule_name", Type: "VARCHAR(100)", Nullable: true},
					{Name: "metric", Type: "VARCHAR(50)", Nullable: true},
					{Name: "operator", Type: "VARCHAR(10)", Nullable: true},
					{Name: "threshold", Type: "FLOAT8", Nullable: true},
					{Name: "severity", Type: "VARCHAR(20)", Nullable: true},
					{Name: "notify_method", Type: "VARCHAR(50)", Nullable: true},
				},
			},
		},
	}
	return info, nil
}

func dumpTSDBSchema() (*SchemaInfo, error) {
	info := &SchemaInfo{
		Database: "tsdb",
		DDL:      sampledb.TSDBSchema,
		Tables: []TableInfo{
			{
				Name:   "meter_data",
				Schema: "tsdb",
				Columns: []ColumnInfo{
					{Name: "ts", Type: "TIMESTAMPTZ(3)", Nullable: false},
					{Name: "voltage", Type: "FLOAT8", Nullable: true},
					{Name: "current", Type: "FLOAT8", Nullable: true},
					{Name: "power", Type: "FLOAT8", Nullable: true},
					{Name: "energy", Type: "FLOAT8", Nullable: true},
				},
				Tags: []TagInfo{
					{Name: "meter_id", Type: "VARCHAR(50)"},
				},
			},
		},
	}
	return info, nil
}

// WriteSchemaDDL writes the DDL to a file
func WriteSchemaDDL(info *SchemaInfo, outputFile string) error {
	return os.WriteFile(outputFile, []byte(info.DDL), 0644)
}

// WriteSchemaJSON writes the schema info as JSON to a file or stdout
func WriteSchemaJSON(info *SchemaInfo, outputFile string) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	if outputFile != "" {
		return os.WriteFile(outputFile, data, 0644)
	}
	fmt.Println(string(data))
	return nil
}
