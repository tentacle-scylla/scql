package schema

import (
	"encoding/json"
	"testing"
)

func TestNewSchema(t *testing.T) {
	s := NewSchema()
	if s == nil {
		t.Fatal("NewSchema returned nil")
	}
	if s.Keyspaces == nil {
		t.Error("Keyspaces map should be initialized")
	}
}

func TestAddKeyspace(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")

	if ks == nil {
		t.Fatal("AddKeyspace returned nil")
	}
	if ks.Name != "test_ks" {
		t.Errorf("Name = %q, want %q", ks.Name, "test_ks")
	}
	if ks.Tables == nil {
		t.Error("Tables map should be initialized")
	}

	// Adding same keyspace should return existing
	ks2 := s.AddKeyspace("test_ks")
	if ks2 != ks {
		t.Error("Adding same keyspace should return existing one")
	}
}

func TestKeyspaceReplication(t *testing.T) {
	s := NewSchema()

	// SimpleStrategy
	ks1 := s.AddKeyspace("simple_ks").WithSimpleStrategy(3)
	if ks1.ReplicationClass != "SimpleStrategy" {
		t.Errorf("ReplicationClass = %q, want SimpleStrategy", ks1.ReplicationClass)
	}
	if ks1.ReplicationFactor["replication_factor"] != 3 {
		t.Errorf("replication_factor = %d, want 3", ks1.ReplicationFactor["replication_factor"])
	}

	// NetworkTopologyStrategy
	ks2 := s.AddKeyspace("nts_ks").WithNetworkTopology(map[string]int{"dc1": 3, "dc2": 2})
	if ks2.ReplicationClass != "NetworkTopologyStrategy" {
		t.Errorf("ReplicationClass = %q, want NetworkTopologyStrategy", ks2.ReplicationClass)
	}
	if ks2.ReplicationFactor["dc1"] != 3 {
		t.Errorf("dc1 = %d, want 3", ks2.ReplicationFactor["dc1"])
	}
	if ks2.ReplicationFactor["dc2"] != 2 {
		t.Errorf("dc2 = %d, want 2", ks2.ReplicationFactor["dc2"])
	}
}

func TestAddTable(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")
	tbl := ks.AddTable("users")

	if tbl == nil {
		t.Fatal("AddTable returned nil")
	}
	if tbl.Name != "users" {
		t.Errorf("Name = %q, want %q", tbl.Name, "users")
	}
	if tbl.Keyspace != "test_ks" {
		t.Errorf("Keyspace = %q, want %q", tbl.Keyspace, "test_ks")
	}

	// Adding same table should return existing
	tbl2 := ks.AddTable("users")
	if tbl2 != tbl {
		t.Error("Adding same table should return existing one")
	}
}

func TestTableColumns(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")
	tbl := ks.AddTable("users").
		AddColumn("id", "uuid").
		AddColumn("name", "text").
		AddColumn("email", "text").
		AddStaticColumn("settings", "map<text, text>")

	// Check columns exist
	if len(tbl.Columns) != 4 {
		t.Errorf("len(Columns) = %d, want 4", len(tbl.Columns))
	}

	// Check column order
	expectedOrder := []string{"id", "name", "email", "settings"}
	if len(tbl.ColumnOrder) != len(expectedOrder) {
		t.Errorf("len(ColumnOrder) = %d, want %d", len(tbl.ColumnOrder), len(expectedOrder))
	}
	for i, name := range expectedOrder {
		if tbl.ColumnOrder[i] != name {
			t.Errorf("ColumnOrder[%d] = %q, want %q", i, tbl.ColumnOrder[i], name)
		}
	}

	// Check static column
	settings := tbl.GetColumn("settings")
	if settings == nil {
		t.Fatal("settings column not found")
	}
	if !settings.IsStatic {
		t.Error("settings should be static")
	}
}

func TestTablePrimaryKey(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")
	tbl := ks.AddTable("events").
		AddColumn("user_id", "uuid").
		AddColumn("event_time", "timestamp").
		AddColumn("event_type", "text").
		AddColumn("data", "text").
		SetPartitionKey("user_id").
		SetClusteringKey("event_time", "event_type").
		SetClusteringOrder("event_time", OrderDesc)

	// Check partition key
	if len(tbl.PartitionKey) != 1 || tbl.PartitionKey[0] != "user_id" {
		t.Errorf("PartitionKey = %v, want [user_id]", tbl.PartitionKey)
	}

	userIdCol := tbl.GetColumn("user_id")
	if !userIdCol.IsPartitionKey {
		t.Error("user_id should be partition key")
	}
	if userIdCol.Position != 0 {
		t.Errorf("user_id.Position = %d, want 0", userIdCol.Position)
	}

	// Check clustering key
	if len(tbl.ClusteringKey) != 2 {
		t.Errorf("len(ClusteringKey) = %d, want 2", len(tbl.ClusteringKey))
	}

	eventTimeCol := tbl.GetColumn("event_time")
	if !eventTimeCol.IsClusteringKey {
		t.Error("event_time should be clustering key")
	}
	if eventTimeCol.Position != 1 {
		t.Errorf("event_time.Position = %d, want 1", eventTimeCol.Position)
	}

	// Check clustering order
	if tbl.ClusteringOrder["event_time"] != OrderDesc {
		t.Errorf("event_time order = %q, want DESC", tbl.ClusteringOrder["event_time"])
	}
	if tbl.ClusteringOrder["event_type"] != OrderAsc {
		t.Errorf("event_type order = %q, want ASC", tbl.ClusteringOrder["event_type"])
	}
}

func TestTableKeyColumns(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")
	tbl := ks.AddTable("events").
		AddColumn("user_id", "uuid").
		AddColumn("bucket", "int").
		AddColumn("event_time", "timestamp").
		AddColumn("data", "text").
		SetPartitionKey("user_id", "bucket").
		SetClusteringKey("event_time")

	// Test PartitionKeyColumns
	pkCols := tbl.PartitionKeyColumns()
	if len(pkCols) != 2 {
		t.Errorf("len(PartitionKeyColumns) = %d, want 2", len(pkCols))
	}

	// Test ClusteringKeyColumns
	ckCols := tbl.ClusteringKeyColumns()
	if len(ckCols) != 1 {
		t.Errorf("len(ClusteringKeyColumns) = %d, want 1", len(ckCols))
	}

	// Test PrimaryKeyColumns
	pkAllCols := tbl.PrimaryKeyColumns()
	if len(pkAllCols) != 3 {
		t.Errorf("len(PrimaryKeyColumns) = %d, want 3", len(pkAllCols))
	}

	// Test RegularColumns
	regularCols := tbl.RegularColumns()
	if len(regularCols) != 1 {
		t.Errorf("len(RegularColumns) = %d, want 1", len(regularCols))
	}
	if regularCols[0].Name != "data" {
		t.Errorf("RegularColumns[0].Name = %q, want data", regularCols[0].Name)
	}
}

func TestAddIndex(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")
	tbl := ks.AddTable("users").
		AddColumn("id", "uuid").
		AddColumn("email", "text").
		SetPartitionKey("id")

	idx := tbl.AddIndex("users_email_idx", "email")
	if idx == nil {
		t.Fatal("AddIndex returned nil")
	}
	if idx.Name != "users_email_idx" {
		t.Errorf("Name = %q, want users_email_idx", idx.Name)
	}
	if idx.TargetColumn != "email" {
		t.Errorf("TargetColumn = %q, want email", idx.TargetColumn)
	}
	if idx.Table != "users" {
		t.Errorf("Table = %q, want users", idx.Table)
	}

	// Test GetIndex
	idx2 := tbl.GetIndex("users_email_idx")
	if idx2 != idx {
		t.Error("GetIndex should return same index")
	}
}

func TestAddUserType(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")
	udt := ks.AddType("address").
		AddField("street", "text").
		AddField("city", "text").
		AddField("zip", "text")

	if udt == nil {
		t.Fatal("AddType returned nil")
	}
	if udt.Name != "address" {
		t.Errorf("Name = %q, want address", udt.Name)
	}
	if udt.Keyspace != "test_ks" {
		t.Errorf("Keyspace = %q, want test_ks", udt.Keyspace)
	}
	if len(udt.Fields) != 3 {
		t.Errorf("len(Fields) = %d, want 3", len(udt.Fields))
	}
	if udt.Fields["city"] != "text" {
		t.Errorf("Fields[city] = %q, want text", udt.Fields["city"])
	}

	// Check field order
	if len(udt.FieldOrder) != 3 {
		t.Errorf("len(FieldOrder) = %d, want 3", len(udt.FieldOrder))
	}
}

func TestAddFunction(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")
	fn := ks.AddFunction("double_it").
		AddParameter("input", "int").
		WithReturnType("int").
		WithLanguage("java").
		WithBody("return input * 2;").
		CalledOnNullInput()

	if fn == nil {
		t.Fatal("AddFunction returned nil")
	}
	if fn.Name != "double_it" {
		t.Errorf("Name = %q, want double_it", fn.Name)
	}
	if len(fn.Parameters) != 1 {
		t.Errorf("len(Parameters) = %d, want 1", len(fn.Parameters))
	}
	if fn.ReturnType != "int" {
		t.Errorf("ReturnType = %q, want int", fn.ReturnType)
	}
	if fn.Language != "java" {
		t.Errorf("Language = %q, want java", fn.Language)
	}
	if !fn.CalledOnNull {
		t.Error("CalledOnNull should be true")
	}
}

func TestSchemaLookups(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")
	tbl := ks.AddTable("users").
		AddColumn("id", "uuid").
		SetPartitionKey("id")

	// Test GetKeyspace
	if s.GetKeyspace("test_ks") != ks {
		t.Error("GetKeyspace failed")
	}
	if s.GetKeyspace("nonexistent") != nil {
		t.Error("GetKeyspace should return nil for nonexistent")
	}

	// Test GetTable
	if ks.GetTable("users") != tbl {
		t.Error("GetTable failed")
	}
	if ks.GetTable("nonexistent") != nil {
		t.Error("GetTable should return nil for nonexistent")
	}

	// Test GetColumn
	if tbl.GetColumn("id") == nil {
		t.Error("GetColumn failed")
	}
	if tbl.GetColumn("nonexistent") != nil {
		t.Error("GetColumn should return nil for nonexistent")
	}

	// Test KeyspaceNames
	names := s.KeyspaceNames()
	if len(names) != 1 || names[0] != "test_ks" {
		t.Errorf("KeyspaceNames = %v, want [test_ks]", names)
	}

	// Test TableNames
	tableNames := ks.TableNames()
	if len(tableNames) != 1 || tableNames[0] != "users" {
		t.Errorf("TableNames = %v, want [users]", tableNames)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	// Build a schema
	s := NewSchema()
	ks := s.AddKeyspace("test_ks").WithSimpleStrategy(3)
	ks.AddTable("users").
		AddColumn("id", "uuid").
		AddColumn("name", "text").
		AddColumn("email", "text").
		SetPartitionKey("id").
		WithComment("Users table")

	ks.AddType("address").
		AddField("street", "text").
		AddField("city", "text")

	// Serialize to JSON
	data, err := s.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Parse back
	s2, err := ParseJSON(data)
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	// Verify
	ks2 := s2.GetKeyspace("test_ks")
	if ks2 == nil {
		t.Fatal("Keyspace not found after round-trip")
	}
	if ks2.ReplicationClass != "SimpleStrategy" {
		t.Errorf("ReplicationClass = %q, want SimpleStrategy", ks2.ReplicationClass)
	}

	tbl2 := ks2.GetTable("users")
	if tbl2 == nil {
		t.Fatal("Table not found after round-trip")
	}
	if tbl2.Comment != "Users table" {
		t.Errorf("Comment = %q, want 'Users table'", tbl2.Comment)
	}

	idCol := tbl2.GetColumn("id")
	if idCol == nil {
		t.Fatal("Column id not found")
	}
	if !idCol.IsPartitionKey {
		t.Error("id should be partition key after round-trip")
	}

	udt2 := ks2.GetType("address")
	if udt2 == nil {
		t.Fatal("UDT not found after round-trip")
	}
	if udt2.Fields["city"] != "text" {
		t.Errorf("UDT field city = %q, want text", udt2.Fields["city"])
	}
}

func TestMaterializedView(t *testing.T) {
	s := NewSchema()
	ks := s.AddKeyspace("test_ks")
	tbl := ks.AddTable("users").
		AddColumn("id", "uuid").
		AddColumn("name", "text").
		AddColumn("email", "text").
		SetPartitionKey("id")

	mv := tbl.AddMaterializedView("users_by_email").
		AddColumn("email", "text").
		AddColumn("id", "uuid").
		AddColumn("name", "text").
		SetPartitionKey("email").
		SetClusteringKey("id").
		WithWhereClause("email IS NOT NULL AND id IS NOT NULL")

	if mv == nil {
		t.Fatal("AddMaterializedView returned nil")
	}
	if mv.Name != "users_by_email" {
		t.Errorf("Name = %q, want users_by_email", mv.Name)
	}
	if mv.BaseTable != "users" {
		t.Errorf("BaseTable = %q, want users", mv.BaseTable)
	}
	if mv.Keyspace != "test_ks" {
		t.Errorf("Keyspace = %q, want test_ks", mv.Keyspace)
	}
	if len(mv.PartitionKey) != 1 || mv.PartitionKey[0] != "email" {
		t.Errorf("PartitionKey = %v, want [email]", mv.PartitionKey)
	}

	// Test GetMaterializedView
	mv2 := tbl.GetMaterializedView("users_by_email")
	if mv2 != mv {
		t.Error("GetMaterializedView should return same view")
	}
}

func TestNilSafety(t *testing.T) {
	var s *Schema
	if s.GetKeyspace("test") != nil {
		t.Error("nil Schema.GetKeyspace should return nil")
	}
	if s.KeyspaceNames() != nil {
		t.Error("nil Schema.KeyspaceNames should return nil")
	}

	var ks *Keyspace
	if ks.GetTable("test") != nil {
		t.Error("nil Keyspace.GetTable should return nil")
	}
	if ks.TableNames() != nil {
		t.Error("nil Keyspace.TableNames should return nil")
	}

	var tbl *Table
	if tbl.GetColumn("test") != nil {
		t.Error("nil Table.GetColumn should return nil")
	}
	if tbl.PartitionKeyColumns() != nil {
		t.Error("nil Table.PartitionKeyColumns should return nil")
	}
	if tbl.AllColumns() != nil {
		t.Error("nil Table.AllColumns should return nil")
	}
}

func TestComplexSchema(t *testing.T) {
	// Build a more realistic schema
	s := NewSchema()

	// System keyspaces (like what you'd see in real ScyllaDB)
	sys := s.AddKeyspace("system").WithSimpleStrategy(1)
	sys.AddTable("local").
		AddColumn("key", "text").
		AddColumn("bootstrapped", "text").
		AddColumn("cluster_name", "text").
		SetPartitionKey("key")

	// Application keyspace
	app := s.AddKeyspace("myapp").WithNetworkTopology(map[string]int{"dc1": 3, "dc2": 2})

	// Users table with composite key
	app.AddTable("users").
		AddColumn("tenant_id", "uuid").
		AddColumn("user_id", "uuid").
		AddColumn("email", "text").
		AddColumn("name", "text").
		AddColumn("created_at", "timestamp").
		SetPartitionKey("tenant_id").
		SetClusteringKey("user_id")

	// Events table with compound partition key
	app.AddTable("events").
		AddColumn("tenant_id", "uuid").
		AddColumn("day", "date").
		AddColumn("event_time", "timestamp").
		AddColumn("event_type", "text").
		AddColumn("data", "text").
		SetPartitionKey("tenant_id", "day").
		SetClusteringKey("event_time").
		SetClusteringOrder("event_time", OrderDesc)

	// UDT
	app.AddType("address").
		AddField("street", "text").
		AddField("city", "text").
		AddField("country", "text").
		AddField("postal_code", "text")

	// Verify structure
	if len(s.Keyspaces) != 2 {
		t.Errorf("len(Keyspaces) = %d, want 2", len(s.Keyspaces))
	}

	events := app.GetTable("events")
	if len(events.PartitionKey) != 2 {
		t.Errorf("events partition key size = %d, want 2", len(events.PartitionKey))
	}

	// JSON round-trip
	data, err := s.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Invalid JSON produced: %v", err)
	}

	// Parse back and verify
	s2, err := ParseJSON(data)
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	events2 := s2.GetKeyspace("myapp").GetTable("events")
	if events2.ClusteringOrder["event_time"] != OrderDesc {
		t.Error("Clustering order not preserved in round-trip")
	}
}
