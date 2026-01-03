package format

import (
	"strings"
	"testing"

	"github.com/tentacle-scylla/scql/pkg/parse"
)

func TestCompactString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "select with whitespace",
			input: "SELECT   *   FROM   users  ;",
		},
		{
			name:  "multiline select",
			input: "SELECT *\nFROM users\nWHERE id = 1;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := CompactString(tt.input)
			if err != nil {
				t.Fatalf("CompactString() error: %v", err)
			}

			if strings.Contains(output, "\n") {
				t.Errorf("Compact output contains newlines: %q", output)
			}

			if !parse.IsValid(output+";") && !parse.IsValid(output) {
				t.Errorf("Compact output is not valid CQL: %q", output)
			}
		})
	}
}

func TestPrettyString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple select",
			input: "SELECT * FROM users WHERE id = 1;",
		},
		{
			name:  "select with multiple clauses",
			input: "SELECT id, name FROM users WHERE id = 1 ORDER BY name LIMIT 10;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := PrettyString(tt.input)
			if err != nil {
				t.Fatalf("PrettyString() error: %v", err)
			}

			t.Logf("Pretty output:\n%s", output)

			lines := strings.Split(output, "\n")
			if len(lines) < 2 {
				t.Log("Pretty output is single line (acceptable for simple queries)")
			}
		})
	}
}

func TestOptions(t *testing.T) {
	input := "select * from users;"

	// Test uppercase keywords
	opts := Options{
		Style:             Compact,
		UppercaseKeywords: true,
	}
	result := parse.Parse(input)
	output := Format(result, opts)
	if !strings.Contains(output, "SELECT") {
		t.Errorf("Expected uppercase SELECT, got: %s", output)
	}

	// Test lowercase keywords
	opts.UppercaseKeywords = false
	output = Format(result, opts)
	if !strings.Contains(output, "select") {
		t.Errorf("Expected lowercase select, got: %s", output)
	}
}

func TestFormatInsert(t *testing.T) {
	input := "INSERT INTO users (id, name) VALUES (1, 'test');"

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("Pretty INSERT:\n%s", output)

	if !strings.Contains(output, "INSERT") {
		t.Error("Output should contain INSERT")
	}
	if !strings.Contains(output, "VALUES") {
		t.Error("Output should contain VALUES")
	}
}

func TestFormatUpdate(t *testing.T) {
	input := "UPDATE users SET name = 'test' WHERE id = 1;"

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("Pretty UPDATE:\n%s", output)

	if !strings.Contains(output, "UPDATE") {
		t.Error("Output should contain UPDATE")
	}
	if !strings.Contains(output, "SET") {
		t.Error("Output should contain SET")
	}
}

func TestFormatCreateTable(t *testing.T) {
	input := `CREATE TABLE users (id int PRIMARY KEY, name text) WITH comment = 'test';`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("Pretty CREATE TABLE:\n%s", output)

	if !strings.Contains(output, "CREATE TABLE") {
		t.Error("Output should contain CREATE TABLE")
	}
}

func TestFormatBatch(t *testing.T) {
	input := `BEGIN BATCH INSERT INTO users (id) VALUES (1); UPDATE users SET name = 'test' WHERE id = 1; APPLY BATCH;`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("Pretty BATCH:\n%s", output)

	if !strings.Contains(output, "BEGIN BATCH") {
		t.Error("Output should contain BEGIN BATCH")
	}
	if !strings.Contains(output, "APPLY BATCH") {
		t.Error("Output should contain APPLY BATCH")
	}
}

func TestFormatInvalidCQL(t *testing.T) {
	input := "SELECT * FORM users;"

	_, err := PrettyString(input)
	if err == nil {
		t.Error("Expected error for invalid CQL")
	}
}

func TestFormatComplexCreateTable(t *testing.T) {
	input := `CREATE TABLE orchestrator_multi_dc.builder_build_job_history(build_hour TIMESTAMP, deployment_id UUID, buildkit_ref TEXT, build_data TEXT, build_status TEXT, build_time BIGINT, builder_vm_id TEXT, environment_id UUID, image_size BIGINT, project_id UUID, publish_time BIGINT, service_id UUID, snapshot_id UUID, total_time BIGINT, PRIMARY KEY (build_hour, deployment_id, buildkit_ref));`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("Complex CREATE TABLE:\n%s", output)

	// Check structure
	if !strings.Contains(output, "CREATE TABLE orchestrator_multi_dc.builder_build_job_history (") {
		t.Error("Should contain properly formatted CREATE TABLE header")
	}
	if !strings.Contains(output, "  build_hour TIMESTAMP,") {
		t.Error("Should contain indented column definitions")
	}
	if !strings.Contains(output, "  PRIMARY KEY (build_hour, deployment_id, buildkit_ref)") {
		t.Error("Should contain indented PRIMARY KEY")
	}
	if strings.Contains(output, "PRIMARY KEY\n") {
		t.Error("PRIMARY KEY should not be on a line by itself")
	}
}

func TestFormatCreateTableWithClusteringOrder(t *testing.T) {
	input := `CREATE TABLE users (id int, cluster_id int, ts timestamp, name text, PRIMARY KEY ((id, cluster_id), ts)) WITH CLUSTERING ORDER BY (ts DESC) AND comment = 'Users table' AND gc_grace_seconds = 86400;`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("CREATE TABLE with CLUSTERING ORDER:\n%s", output)

	// Check WITH clause structure
	if !strings.Contains(output, "WITH CLUSTERING ORDER BY") {
		t.Error("Should contain CLUSTERING ORDER BY")
	}
	if !strings.Contains(output, "AND comment = 'Users table'") {
		t.Error("Should contain comment option")
	}
	if !strings.Contains(output, "AND gc_grace_seconds = 86400") {
		t.Error("Should contain gc_grace_seconds option")
	}
}

func TestFormatCreateTablePreservesColumnNames(t *testing.T) {
	// Test that keyword-like column names are preserved lowercase (cqlsh style)
	input := `CREATE TABLE data (id uuid PRIMARY KEY, value blob, key text, type int);`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("CREATE TABLE with keyword column names:\n%s", output)

	// Column names and types should be lowercase (cqlsh style)
	if !strings.Contains(output, "value blob") {
		t.Error("Column name and type should be lowercase: 'value blob'")
	}
	if !strings.Contains(output, "key text") {
		t.Error("Column name and type should be lowercase: 'key text'")
	}
	if !strings.Contains(output, "type int") {
		t.Error("Column name and type should be lowercase: 'type int'")
	}
}

func TestFormatCreateType(t *testing.T) {
	input := `CREATE TYPE address (street text, city text, zip text, country text);`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("CREATE TYPE:\n%s", output)

	// Check structure (cqlsh style: 4-space indentation, lowercase types)
	if !strings.Contains(output, "CREATE TYPE address (") {
		t.Error("Should contain CREATE TYPE header")
	}
	if !strings.Contains(output, "    street text,") {
		t.Error("Should contain 4-space indented field definitions with lowercase types")
	}
	if !strings.Contains(output, "    country text\n)") {
		t.Error("Last field should not have trailing comma")
	}
}

func TestFormatCreateTableCompoundPartitionKey(t *testing.T) {
	input := `CREATE TABLE events (tenant_id uuid, event_date date, event_id timeuuid, data text, PRIMARY KEY ((tenant_id, event_date), event_id));`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("CREATE TABLE compound partition key:\n%s", output)

	// Check compound partition key is preserved
	if !strings.Contains(output, "PRIMARY KEY ((tenant_id, event_date), event_id)") {
		t.Error("Should preserve compound partition key format")
	}
}

func TestFormatCreateMaterializedView(t *testing.T) {
	input := `CREATE MATERIALIZED VIEW orchestrator_multi_dc.volume_instance_assignment_zfs_id_idx_index AS SELECT zfs_id, idx_token, volume_instance_id FROM orchestrator_multi_dc.volume_instance_assignment WHERE zfs_id IS NOT NULL PRIMARY KEY (zfs_id, idx_token, volume_instance_id);`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("CREATE MATERIALIZED VIEW:\n%s", output)

	// Check structure
	if !strings.Contains(output, "CREATE MATERIALIZED VIEW") {
		t.Error("Should contain CREATE MATERIALIZED VIEW")
	}
	if !strings.Contains(output, "AS\n") {
		t.Error("Should have AS on its own line")
	}
	if !strings.Contains(output, "SELECT zfs_id, idx_token, volume_instance_id") {
		t.Error("Should contain SELECT clause")
	}
	if !strings.Contains(output, "FROM orchestrator_multi_dc.volume_instance_assignment") {
		t.Error("Should contain FROM clause")
	}
	if !strings.Contains(output, "WHERE zfs_id IS NOT NULL") {
		t.Error("Should contain WHERE clause")
	}
	if !strings.Contains(output, "PRIMARY KEY") {
		t.Error("Should contain PRIMARY KEY")
	}
}

func TestFormatCreateIndex(t *testing.T) {
	input := `CREATE INDEX user_email_idx ON users (email);`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("CREATE INDEX:\n%s", output)

	if !strings.Contains(output, "CREATE INDEX user_email_idx") {
		t.Error("Should contain CREATE INDEX with name")
	}
	if !strings.Contains(output, "ON users") {
		t.Error("Should contain ON table")
	}
	if !strings.Contains(output, "(email)") {
		t.Error("Should contain column specification")
	}
}

func TestFormatCreateIndexWithKeyspace(t *testing.T) {
	input := `CREATE INDEX IF NOT EXISTS idx_name ON my_keyspace.my_table (my_column);`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("CREATE INDEX with keyspace:\n%s", output)

	if !strings.Contains(output, "CREATE INDEX IF NOT EXISTS") {
		t.Error("Should contain IF NOT EXISTS")
	}
	if !strings.Contains(output, "ON my_keyspace.my_table") {
		t.Error("Should contain my_keyspace.my_table")
	}
}

func TestFormatCreateKeyspace(t *testing.T) {
	input := `CREATE KEYSPACE test_ks WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 3} AND durable_writes = true;`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("CREATE KEYSPACE:\n%s", output)

	if !strings.Contains(output, "CREATE KEYSPACE test_ks") {
		t.Error("Should contain CREATE KEYSPACE with name")
	}
	if !strings.Contains(output, "WITH replication =") {
		t.Error("Should contain WITH replication")
	}
	if !strings.Contains(output, "AND durable_writes") {
		t.Error("Should contain durable_writes option (lowercase, cqlsh style)")
	}
}

func TestFormatAlterKeyspace(t *testing.T) {
	input := `ALTER KEYSPACE test_ks WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': 3, 'dc2': 2} AND durable_writes = true;`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("ALTER KEYSPACE:\n%s", output)

	if !strings.Contains(output, "ALTER KEYSPACE test_ks") {
		t.Error("Should contain ALTER KEYSPACE with name")
	}
	if !strings.Contains(output, "WITH replication =") {
		t.Error("Should contain WITH replication")
	}
}

func TestFormatSelectWithAllClauses(t *testing.T) {
	input := `SELECT id, name, email FROM users WHERE status = 'active' AND created_at > '2024-01-01' ORDER BY created_at DESC LIMIT 100 ALLOW FILTERING;`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("Complex SELECT:\n%s", output)

	if !strings.Contains(output, "SELECT id, name, email") {
		t.Error("Should contain SELECT clause")
	}
	if !strings.Contains(output, "FROM users") {
		t.Error("Should contain FROM clause")
	}
	if !strings.Contains(output, "WHERE status") {
		t.Error("Should contain WHERE clause")
	}
	if !strings.Contains(output, "ORDER BY created_at DESC") {
		t.Error("Should contain ORDER BY clause")
	}
	if !strings.Contains(output, "LIMIT 100") {
		t.Error("Should contain LIMIT clause")
	}
	if !strings.Contains(output, "ALLOW FILTERING") {
		t.Error("Should contain ALLOW FILTERING clause")
	}
}

func TestFormatDelete(t *testing.T) {
	input := `DELETE FROM users WHERE id = 123;`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("DELETE:\n%s", output)

	if !strings.Contains(output, "DELETE") {
		t.Error("Should contain DELETE")
	}
	if !strings.Contains(output, "FROM users") {
		t.Error("Should contain FROM clause")
	}
	if !strings.Contains(output, "WHERE id = 123") {
		t.Error("Should contain WHERE clause")
	}
}

func TestFormatDeleteWithColumns(t *testing.T) {
	input := `DELETE email, phone FROM users WHERE id = 123;`

	output, err := PrettyString(input)
	if err != nil {
		t.Fatalf("PrettyString() error: %v", err)
	}

	t.Logf("DELETE with columns:\n%s", output)

	if !strings.Contains(output, "DELETE email, phone") {
		t.Error("Should contain DELETE with columns")
	}
}
