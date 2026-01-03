// Package schema provides types for representing CQL schemas.
// These types can be populated from various sources (live DB, JSON, CQL DDL)
// and used for query validation, auto-completion, and type hints.
package schema

// Schema represents a complete CQL schema with all keyspaces.
type Schema struct {
	Keyspaces map[string]*Keyspace
}

// Keyspace represents a CQL keyspace with its tables, types, and functions.
type Keyspace struct {
	Name              string
	ReplicationClass  string            // SimpleStrategy, NetworkTopologyStrategy, etc.
	ReplicationFactor map[string]int    // DC name -> replication factor (or "replication_factor" -> n)
	DurableWrites     bool
	Tables            map[string]*Table
	Types             map[string]*UserType   // User-defined types (UDTs)
	Functions         map[string]*Function   // User-defined functions (UDFs)
	Aggregates        map[string]*Aggregate  // User-defined aggregates (UDAs)
}

// Table represents a CQL table with its columns and keys.
type Table struct {
	Name              string
	Keyspace          string
	Columns           map[string]*Column
	ColumnOrder       []string          // Preserves column definition order
	PartitionKey      []string          // Column names in partition key order
	ClusteringKey     []string          // Column names in clustering key order
	ClusteringOrder   map[string]Order  // Column -> ASC/DESC
	Indexes           map[string]*Index
	MaterializedViews map[string]*MaterializedView
	Comment           string
	// Table options
	GCGraceSeconds      int
	BloomFilterFPChance float64
	Compaction          map[string]string
	Compression         map[string]string
	Caching             map[string]string
}

// Column represents a column in a table.
type Column struct {
	Name            string
	Type            string // CQL type string (e.g., "text", "int", "map<text, int>")
	IsStatic        bool
	IsPartitionKey  bool
	IsClusteringKey bool
	Position        int // Position in primary key (0-based)
}

// Order represents clustering order direction.
type Order string

const (
	OrderAsc  Order = "ASC"
	OrderDesc Order = "DESC"
)

// Index represents a secondary index on a table.
type Index struct {
	Name         string
	Table        string
	TargetColumn string
	Kind         string            // COMPOSITES, CUSTOM, etc.
	Options      map[string]string // For custom index options
	ClassName    string            // For custom indexes
}

// MaterializedView represents a materialized view.
type MaterializedView struct {
	Name           string
	Keyspace       string
	BaseTable      string
	Columns        map[string]*Column
	ColumnOrder    []string
	PartitionKey   []string
	ClusteringKey  []string
	ClusteringOrder map[string]Order
	WhereClause    string // The WHERE clause from the view definition
}

// UserType represents a user-defined type (UDT).
type UserType struct {
	Name       string
	Keyspace   string
	Fields     map[string]string // Field name -> CQL type
	FieldOrder []string          // Preserves field definition order
}

// Function represents a user-defined function (UDF).
type Function struct {
	Name           string
	Keyspace       string
	Parameters     []FunctionParam
	ReturnType     string
	Language       string // java, javascript, lua, etc.
	Body           string
	CalledOnNull   bool   // true = CALLED ON NULL INPUT, false = RETURNS NULL ON NULL INPUT
	Deterministic  bool
}

// FunctionParam represents a function parameter.
type FunctionParam struct {
	Name string
	Type string
}

// Aggregate represents a user-defined aggregate (UDA).
type Aggregate struct {
	Name       string
	Keyspace   string
	StateFunc  string
	StateType  string
	FinalFunc  string
	InitCond   string
	Parameters []string // Parameter types
	ReturnType string
}

// Lookup methods

// GetKeyspace returns a keyspace by name, or nil if not found.
func (s *Schema) GetKeyspace(name string) *Keyspace {
	if s == nil || s.Keyspaces == nil {
		return nil
	}
	return s.Keyspaces[name]
}

// GetTable returns a table by name, or nil if not found.
func (ks *Keyspace) GetTable(name string) *Table {
	if ks == nil || ks.Tables == nil {
		return nil
	}
	return ks.Tables[name]
}

// GetColumn returns a column by name, or nil if not found.
func (t *Table) GetColumn(name string) *Column {
	if t == nil || t.Columns == nil {
		return nil
	}
	return t.Columns[name]
}

// GetType returns a user-defined type by name, or nil if not found.
func (ks *Keyspace) GetType(name string) *UserType {
	if ks == nil || ks.Types == nil {
		return nil
	}
	return ks.Types[name]
}

// GetFunction returns a user-defined function by name, or nil if not found.
func (ks *Keyspace) GetFunction(name string) *Function {
	if ks == nil || ks.Functions == nil {
		return nil
	}
	return ks.Functions[name]
}

// GetIndex returns an index by name, or nil if not found.
func (t *Table) GetIndex(name string) *Index {
	if t == nil || t.Indexes == nil {
		return nil
	}
	return t.Indexes[name]
}

// GetMaterializedView returns a materialized view by name, or nil if not found.
func (t *Table) GetMaterializedView(name string) *MaterializedView {
	if t == nil || t.MaterializedViews == nil {
		return nil
	}
	return t.MaterializedViews[name]
}

// Utility methods

// PartitionKeyColumns returns the partition key columns in order.
func (t *Table) PartitionKeyColumns() []*Column {
	if t == nil {
		return nil
	}
	cols := make([]*Column, 0, len(t.PartitionKey))
	for _, name := range t.PartitionKey {
		if col := t.GetColumn(name); col != nil {
			cols = append(cols, col)
		}
	}
	return cols
}

// ClusteringKeyColumns returns the clustering key columns in order.
func (t *Table) ClusteringKeyColumns() []*Column {
	if t == nil {
		return nil
	}
	cols := make([]*Column, 0, len(t.ClusteringKey))
	for _, name := range t.ClusteringKey {
		if col := t.GetColumn(name); col != nil {
			cols = append(cols, col)
		}
	}
	return cols
}

// PrimaryKeyColumns returns all primary key columns (partition + clustering) in order.
func (t *Table) PrimaryKeyColumns() []*Column {
	if t == nil {
		return nil
	}
	cols := make([]*Column, 0, len(t.PartitionKey)+len(t.ClusteringKey))
	cols = append(cols, t.PartitionKeyColumns()...)
	cols = append(cols, t.ClusteringKeyColumns()...)
	return cols
}

// RegularColumns returns all non-primary-key columns.
func (t *Table) RegularColumns() []*Column {
	if t == nil {
		return nil
	}
	cols := make([]*Column, 0)
	for _, name := range t.ColumnOrder {
		col := t.GetColumn(name)
		if col != nil && !col.IsPartitionKey && !col.IsClusteringKey {
			cols = append(cols, col)
		}
	}
	return cols
}

// AllColumns returns all columns in definition order.
func (t *Table) AllColumns() []*Column {
	if t == nil {
		return nil
	}
	cols := make([]*Column, 0, len(t.ColumnOrder))
	for _, name := range t.ColumnOrder {
		if col := t.GetColumn(name); col != nil {
			cols = append(cols, col)
		}
	}
	return cols
}

// TableNames returns all table names in the keyspace.
func (ks *Keyspace) TableNames() []string {
	if ks == nil || ks.Tables == nil {
		return nil
	}
	names := make([]string, 0, len(ks.Tables))
	for name := range ks.Tables {
		names = append(names, name)
	}
	return names
}

// MaterializedViewNames returns all materialized view names in the keyspace.
// MVs are collected from all tables in the keyspace.
func (ks *Keyspace) MaterializedViewNames() []string {
	if ks == nil || ks.Tables == nil {
		return nil
	}
	var names []string
	for _, tbl := range ks.Tables {
		if tbl.MaterializedViews != nil {
			for name := range tbl.MaterializedViews {
				names = append(names, name)
			}
		}
	}
	return names
}

// GetMaterializedView returns a materialized view by name from any table in the keyspace.
func (ks *Keyspace) GetMaterializedView(name string) *MaterializedView {
	if ks == nil || ks.Tables == nil {
		return nil
	}
	for _, tbl := range ks.Tables {
		if mv := tbl.GetMaterializedView(name); mv != nil {
			return mv
		}
	}
	return nil
}

// KeyspaceNames returns all keyspace names in the schema.
func (s *Schema) KeyspaceNames() []string {
	if s == nil || s.Keyspaces == nil {
		return nil
	}
	names := make([]string, 0, len(s.Keyspaces))
	for name := range s.Keyspaces {
		names = append(names, name)
	}
	return names
}
