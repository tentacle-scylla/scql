package schema

// NewSchema creates a new empty schema.
func NewSchema() *Schema {
	return &Schema{
		Keyspaces: make(map[string]*Keyspace),
	}
}

// AddKeyspace adds a new keyspace to the schema and returns it.
// If a keyspace with the same name already exists, it returns the existing one.
func (s *Schema) AddKeyspace(name string) *Keyspace {
	if s.Keyspaces == nil {
		s.Keyspaces = make(map[string]*Keyspace)
	}
	if ks, exists := s.Keyspaces[name]; exists {
		return ks
	}
	ks := &Keyspace{
		Name:              name,
		ReplicationFactor: make(map[string]int),
		Tables:            make(map[string]*Table),
		Types:             make(map[string]*UserType),
		Functions:         make(map[string]*Function),
		Aggregates:        make(map[string]*Aggregate),
		DurableWrites:     true,
	}
	s.Keyspaces[name] = ks
	return ks
}

// WithReplication sets the replication strategy for the keyspace.
func (ks *Keyspace) WithReplication(class string, factors map[string]int) *Keyspace {
	ks.ReplicationClass = class
	ks.ReplicationFactor = factors
	return ks
}

// WithSimpleStrategy sets SimpleStrategy replication for the keyspace.
func (ks *Keyspace) WithSimpleStrategy(replicationFactor int) *Keyspace {
	ks.ReplicationClass = "SimpleStrategy"
	ks.ReplicationFactor = map[string]int{"replication_factor": replicationFactor}
	return ks
}

// WithNetworkTopology sets NetworkTopologyStrategy replication for the keyspace.
func (ks *Keyspace) WithNetworkTopology(dcFactors map[string]int) *Keyspace {
	ks.ReplicationClass = "NetworkTopologyStrategy"
	ks.ReplicationFactor = dcFactors
	return ks
}

// WithDurableWrites sets durable_writes for the keyspace.
func (ks *Keyspace) WithDurableWrites(durable bool) *Keyspace {
	ks.DurableWrites = durable
	return ks
}

// AddTable adds a new table to the keyspace and returns it.
// If a table with the same name already exists, it returns the existing one.
func (ks *Keyspace) AddTable(name string) *Table {
	if ks.Tables == nil {
		ks.Tables = make(map[string]*Table)
	}
	if t, exists := ks.Tables[name]; exists {
		return t
	}
	t := &Table{
		Name:              name,
		Keyspace:          ks.Name,
		Columns:           make(map[string]*Column),
		ColumnOrder:       make([]string, 0),
		PartitionKey:      make([]string, 0),
		ClusteringKey:     make([]string, 0),
		ClusteringOrder:   make(map[string]Order),
		Indexes:           make(map[string]*Index),
		MaterializedViews: make(map[string]*MaterializedView),
		Compaction:        make(map[string]string),
		Compression:       make(map[string]string),
		Caching:           make(map[string]string),
	}
	ks.Tables[name] = t
	return t
}

// AddColumn adds a column to the table and returns the table for chaining.
func (t *Table) AddColumn(name, cqlType string) *Table {
	if t.Columns == nil {
		t.Columns = make(map[string]*Column)
	}
	col := &Column{
		Name: name,
		Type: cqlType,
	}
	t.Columns[name] = col
	t.ColumnOrder = append(t.ColumnOrder, name)
	return t
}

// AddStaticColumn adds a static column to the table.
func (t *Table) AddStaticColumn(name, cqlType string) *Table {
	t.AddColumn(name, cqlType)
	if col := t.GetColumn(name); col != nil {
		col.IsStatic = true
	}
	return t
}

// SetPartitionKey sets the partition key columns.
// Columns must already exist in the table.
func (t *Table) SetPartitionKey(columns ...string) *Table {
	t.PartitionKey = columns
	for i, name := range columns {
		if col := t.GetColumn(name); col != nil {
			col.IsPartitionKey = true
			col.Position = i
		}
	}
	return t
}

// SetClusteringKey sets the clustering key columns with default ASC order.
// Columns must already exist in the table.
func (t *Table) SetClusteringKey(columns ...string) *Table {
	t.ClusteringKey = columns
	for i, name := range columns {
		if col := t.GetColumn(name); col != nil {
			col.IsClusteringKey = true
			col.Position = len(t.PartitionKey) + i
		}
		if t.ClusteringOrder == nil {
			t.ClusteringOrder = make(map[string]Order)
		}
		t.ClusteringOrder[name] = OrderAsc
	}
	return t
}

// SetClusteringOrder sets the clustering order for a column.
func (t *Table) SetClusteringOrder(column string, order Order) *Table {
	if t.ClusteringOrder == nil {
		t.ClusteringOrder = make(map[string]Order)
	}
	t.ClusteringOrder[column] = order
	return t
}

// WithComment sets the table comment.
func (t *Table) WithComment(comment string) *Table {
	t.Comment = comment
	return t
}

// WithGCGraceSeconds sets gc_grace_seconds for the table.
func (t *Table) WithGCGraceSeconds(seconds int) *Table {
	t.GCGraceSeconds = seconds
	return t
}

// AddIndex adds an index to the table.
func (t *Table) AddIndex(name, targetColumn string) *Index {
	if t.Indexes == nil {
		t.Indexes = make(map[string]*Index)
	}
	idx := &Index{
		Name:         name,
		Table:        t.Name,
		TargetColumn: targetColumn,
		Options:      make(map[string]string),
	}
	t.Indexes[name] = idx
	return idx
}

// WithKind sets the index kind.
func (idx *Index) WithKind(kind string) *Index {
	idx.Kind = kind
	return idx
}

// WithClassName sets the custom index class name.
func (idx *Index) WithClassName(className string) *Index {
	idx.ClassName = className
	return idx
}

// AddType adds a user-defined type to the keyspace.
func (ks *Keyspace) AddType(name string) *UserType {
	if ks.Types == nil {
		ks.Types = make(map[string]*UserType)
	}
	udt := &UserType{
		Name:       name,
		Keyspace:   ks.Name,
		Fields:     make(map[string]string),
		FieldOrder: make([]string, 0),
	}
	ks.Types[name] = udt
	return udt
}

// AddField adds a field to the user-defined type.
func (udt *UserType) AddField(name, cqlType string) *UserType {
	if udt.Fields == nil {
		udt.Fields = make(map[string]string)
	}
	udt.Fields[name] = cqlType
	udt.FieldOrder = append(udt.FieldOrder, name)
	return udt
}

// AddFunction adds a user-defined function to the keyspace.
func (ks *Keyspace) AddFunction(name string) *Function {
	if ks.Functions == nil {
		ks.Functions = make(map[string]*Function)
	}
	fn := &Function{
		Name:       name,
		Keyspace:   ks.Name,
		Parameters: make([]FunctionParam, 0),
	}
	ks.Functions[name] = fn
	return fn
}

// AddParameter adds a parameter to the function.
func (fn *Function) AddParameter(name, cqlType string) *Function {
	fn.Parameters = append(fn.Parameters, FunctionParam{Name: name, Type: cqlType})
	return fn
}

// WithReturnType sets the return type of the function.
func (fn *Function) WithReturnType(cqlType string) *Function {
	fn.ReturnType = cqlType
	return fn
}

// WithLanguage sets the language of the function.
func (fn *Function) WithLanguage(lang string) *Function {
	fn.Language = lang
	return fn
}

// WithBody sets the body of the function.
func (fn *Function) WithBody(body string) *Function {
	fn.Body = body
	return fn
}

// CalledOnNullInput sets the function to be called on null input.
func (fn *Function) CalledOnNullInput() *Function {
	fn.CalledOnNull = true
	return fn
}

// ReturnsNullOnNullInput sets the function to return null on null input.
func (fn *Function) ReturnsNullOnNullInput() *Function {
	fn.CalledOnNull = false
	return fn
}

// AddMaterializedView adds a materialized view to the table.
func (t *Table) AddMaterializedView(name string) *MaterializedView {
	if t.MaterializedViews == nil {
		t.MaterializedViews = make(map[string]*MaterializedView)
	}
	mv := &MaterializedView{
		Name:            name,
		Keyspace:        t.Keyspace,
		BaseTable:       t.Name,
		Columns:         make(map[string]*Column),
		ColumnOrder:     make([]string, 0),
		PartitionKey:    make([]string, 0),
		ClusteringKey:   make([]string, 0),
		ClusteringOrder: make(map[string]Order),
	}
	t.MaterializedViews[name] = mv
	return mv
}

// AddColumn adds a column to the materialized view.
func (mv *MaterializedView) AddColumn(name, cqlType string) *MaterializedView {
	if mv.Columns == nil {
		mv.Columns = make(map[string]*Column)
	}
	col := &Column{
		Name: name,
		Type: cqlType,
	}
	mv.Columns[name] = col
	mv.ColumnOrder = append(mv.ColumnOrder, name)
	return mv
}

// SetPartitionKey sets the partition key for the materialized view.
func (mv *MaterializedView) SetPartitionKey(columns ...string) *MaterializedView {
	mv.PartitionKey = columns
	for i, name := range columns {
		if col := mv.Columns[name]; col != nil {
			col.IsPartitionKey = true
			col.Position = i
		}
	}
	return mv
}

// SetClusteringKey sets the clustering key for the materialized view.
func (mv *MaterializedView) SetClusteringKey(columns ...string) *MaterializedView {
	mv.ClusteringKey = columns
	for i, name := range columns {
		if col := mv.Columns[name]; col != nil {
			col.IsClusteringKey = true
			col.Position = len(mv.PartitionKey) + i
		}
		if mv.ClusteringOrder == nil {
			mv.ClusteringOrder = make(map[string]Order)
		}
		mv.ClusteringOrder[name] = OrderAsc
	}
	return mv
}

// WithWhereClause sets the WHERE clause for the materialized view.
func (mv *MaterializedView) WithWhereClause(where string) *MaterializedView {
	mv.WhereClause = where
	return mv
}
