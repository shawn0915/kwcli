package perf

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shawn0915/kwcli/pkg/config"
)

// Snapshot represents a performance snapshot
type Snapshot struct {
	Timestamp    string      `json:"timestamp"`
	QPS          QPSInfo     `json:"qps"`
	TPS          TPSInfo     `json:"tps"`
	Connections  ConnInfo    `json:"connections"`
	SlowQueries  []QueryInfo `json:"slow_queries"`
	Storage      StorageInfo `json:"storage"`
	CacheHitRate string      `json:"cache_hit_rate"`
}

// QPSInfo represents queries per second
type QPSInfo struct {
	Value     float64 `json:"value"`
	Available bool    `json:"available"`
}

// TPSInfo represents transactions per second
type TPSInfo struct {
	Value     float64 `json:"value"`
	Available bool    `json:"available"`
}

// ConnInfo represents connection information
type ConnInfo struct {
	Active    int  `json:"active"`
	Available bool `json:"available"`
}

// QueryInfo represents a running query
type QueryInfo struct {
	ID       string `json:"id"`
	Query    string `json:"query"`
	Duration string `json:"duration"`
}

// StorageInfo represents storage usage information
type StorageInfo struct {
	DataDir   string `json:"data_dir,omitempty"`
	Size      string `json:"size,omitempty"`
	Available bool   `json:"available"`
}

// TakeSnapshot collects a performance snapshot
func TakeSnapshot() (*Snapshot, error) {
	snapshot := &Snapshot{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Collect QPS/TPS from metrics
	collectQPSTPS(snapshot)

	// Collect connections
	collectConnections(snapshot)

	// Collect slow/running queries
	collectSlowQueries(snapshot)

	// Collect storage info
	collectStorage(snapshot)

	return snapshot, nil
}

func collectQPSTPS(snapshot *Snapshot) {
	// Try to get metrics from KWDB admin endpoint
	cfg, err := config.LoadKWDBConfig()
	if err != nil {
		snapshot.QPS = QPSInfo{Available: false}
		snapshot.TPS = TPSInfo{Available: false}
		return
	}

	httpAddr := cfg.HTTPAddr
	if httpAddr == "" {
		httpAddr = "0.0.0.0:8080"
	}

	// Try to fetch metrics from the _status/var endpoint
	cmd := exec.Command("curl", "-s", fmt.Sprintf("http://%s/_status/var", httpAddr))
	output, err := cmd.Output()
	if err != nil {
		snapshot.QPS = QPSInfo{Available: false}
		snapshot.TPS = TPSInfo{Available: false}
		return
	}

	metricsStr := string(output)

	// Parse sql_query_count metric for QPS
	for _, line := range strings.Split(metricsStr, "\n") {
		if strings.HasPrefix(line, "sql_query_count") && !strings.HasPrefix(line, "#") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				var val float64
				if _, err := fmt.Sscanf(parts[len(parts)-1], "%f", &val); err == nil {
					snapshot.QPS = QPSInfo{Value: val, Available: true}
				}
			}
		}
		if strings.HasPrefix(line, "txn_commits_count") && !strings.HasPrefix(line, "#") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				var val float64
				if _, err := fmt.Sscanf(parts[len(parts)-1], "%f", &val); err == nil {
					snapshot.TPS = TPSInfo{Value: val, Available: true}
				}
			}
		}
	}

	if !snapshot.QPS.Available {
		snapshot.QPS = QPSInfo{Available: false}
	}
	if !snapshot.TPS.Available {
		snapshot.TPS = TPSInfo{Available: false}
	}
}

func collectConnections(snapshot *Snapshot) {
	output, err := executePerfSQL("SHOW SESSIONS")
	if err != nil {
		snapshot.Connections = ConnInfo{Available: false}
		return
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "active_connections") && !strings.HasPrefix(trimmed, "(") {
			count++
		}
	}

	snapshot.Connections = ConnInfo{Active: count, Available: true}
}

func collectSlowQueries(snapshot *Snapshot) {
	output, err := executePerfSQL("SHOW QUERIES")
	if err != nil {
		snapshot.SlowQueries = []QueryInfo{}
		return
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	queries := []QueryInfo{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "query_id") && !strings.HasPrefix(trimmed, "(") {
			parts := strings.Split(trimmed, "|")
			if len(parts) >= 3 {
				query := QueryInfo{
					ID:       strings.TrimSpace(parts[0]),
					Duration: strings.TrimSpace(parts[1]),
				}
				// Get the query text (usually the last meaningful column)
				if len(parts) >= 5 {
					query.Query = strings.TrimSpace(parts[len(parts)-2])
				} else if len(parts) >= 2 {
					query.Query = strings.TrimSpace(parts[len(parts)-1])
				}
				// Truncate long queries for display
				if len(query.Query) > 200 {
					query.Query = query.Query[:200] + "..."
				}
				queries = append(queries, query)
			}
		}
	}

	// Keep top 10
	if len(queries) > 10 {
		queries = queries[:10]
	}

	snapshot.SlowQueries = queries
}

func collectStorage(snapshot *Snapshot) {
	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if data, err := os.ReadFile(dockerMarker); err == nil {
		runtimeCmd := strings.TrimSpace(string(data))
		if runtimeCmd == "" {
			runtimeCmd = "docker"
		}
		cmd := exec.Command(runtimeCmd, "exec", "kwdb", "du", "-sh", "/kwbase")
		output, err := cmd.Output()
		if err == nil {
			snapshot.Storage = StorageInfo{
				DataDir:   "/kwbase (container)",
				Size:      strings.TrimSpace(string(output)),
				Available: true,
			}
			return
		}
	}

	// Binary mode
	cfg, err := config.LoadKWDBConfig()
	if err != nil || cfg.DataDir == "" {
		snapshot.Storage = StorageInfo{Available: false}
		return
	}

	cmd := exec.Command("du", "-sh", cfg.DataDir)
	output, err := cmd.Output()
	if err != nil {
		snapshot.Storage = StorageInfo{Available: false}
		return
	}

	snapshot.Storage = StorageInfo{
		DataDir:   cfg.DataDir,
		Size:      strings.TrimSpace(string(output)),
		Available: true,
	}
}

func executePerfSQL(sql string) (string, error) {
	installDir := filepath.Join(config.GetComponentsDir(), "kwdb")

	// Check Docker mode
	dockerMarker := installDir + "/.docker_mode"
	if data, err := os.ReadFile(dockerMarker); err == nil {
		runtimeCmd := strings.TrimSpace(string(data))
		if runtimeCmd == "" {
			runtimeCmd = "docker"
		}
		cmd := exec.Command(runtimeCmd, "exec", "kwdb", "./kwbase", "sql", "--insecure", "-e", sql)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	// Binary mode
	binaryMarker := installDir + "/.binary_mode"
	if _, err := os.Stat(binaryMarker); err == nil {
		kwbasePath := findKwbaseBinaryPerf(installDir)
		if kwbasePath != "" {
			cmd := exec.Command(kwbasePath, "sql", "--insecure", "-e", sql)
			output, err := cmd.CombinedOutput()
			return string(output), err
		}
	}

	return "", fmt.Errorf("KWDB is not installed")
}

func findKwbaseBinaryPerf(searchDir string) string {
	var kwbasePath string
	filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && (info.Name() == "kwbase" || info.Name() == "kwbase.exe") {
			kwbasePath = path
		}
		return nil
	})
	return kwbasePath
}

// PrintSnapshot prints the snapshot in human-readable format
func PrintSnapshot(snapshot *Snapshot) {
	fmt.Println("=== KWDB Performance Snapshot ===")
	fmt.Printf("Timestamp: %s\n\n", snapshot.Timestamp)

	// QPS
	if snapshot.QPS.Available {
		fmt.Printf("QPS: %.2f\n", snapshot.QPS.Value)
	} else {
		fmt.Println("QPS: N/A")
	}

	// TPS
	if snapshot.TPS.Available {
		fmt.Printf("TPS: %.2f\n", snapshot.TPS.Value)
	} else {
		fmt.Println("TPS: N/A")
	}

	// Connections
	if snapshot.Connections.Available {
		fmt.Printf("Active Connections: %d\n", snapshot.Connections.Active)
	} else {
		fmt.Println("Active Connections: N/A")
	}

	// Slow Queries
	fmt.Printf("\nRunning Queries (TOP %d):\n", len(snapshot.SlowQueries))
	if len(snapshot.SlowQueries) == 0 {
		fmt.Println("  No running queries")
	} else {
		for i, q := range snapshot.SlowQueries {
			fmt.Printf("  [%d] ID: %s | Duration: %s | Query: %s\n", i+1, q.ID, q.Duration, q.Query)
		}
	}

	// Storage
	if snapshot.Storage.Available {
		fmt.Printf("\nStorage: %s (%s)\n", snapshot.Storage.Size, snapshot.Storage.DataDir)
	} else {
		fmt.Println("\nStorage: N/A")
	}
}

// SnapshotToJSON converts snapshot to JSON string
func SnapshotToJSON(snapshot *Snapshot) (string, error) {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteSnapshot writes the snapshot to a file
func WriteSnapshot(snapshot *Snapshot, format string, outputFile string) error {
	switch format {
	case "json":
		content, err := SnapshotToJSON(snapshot)
		if err != nil {
			return err
		}
		if outputFile != "" {
			return os.WriteFile(outputFile, []byte(content), 0644)
		}
		fmt.Println(content)
	default:
		return fmt.Errorf("unsupported format: %s (supported: json)", format)
	}
	return nil
}
