package complete

import "fmt"

// Group ID prefixes
const (
	GroupPrefixKeyspace = "ks:"  // Keyspace group, e.g., "ks:myapp"
	GroupPrefixCategory = "cat:" // Category group, e.g., "cat:aggregate"
	GroupPrefixKeyType  = "key:" // Key type group, e.g., "key:partition"
	GroupPrefixSource   = "src:" // Source table group, e.g., "src:users"
)

// Predefined key type group IDs
const (
	GroupKeyPartition  = "key:partition"
	GroupKeyClustering = "key:clustering"
	GroupKeyRegular    = "key:regular"
	GroupKeyStatic     = "key:static"
)

// Predefined function category group IDs
const (
	GroupCatAggregate  = "cat:aggregate"
	GroupCatTime       = "cat:time"
	GroupCatUUID       = "cat:uuid"
	GroupCatConversion = "cat:conversion"
	GroupCatCollection = "cat:collection"
	GroupCatBlob       = "cat:blob"
	GroupCatMath       = "cat:math"
	GroupCatString     = "cat:string"
	GroupCatOther      = "cat:other"
)

// Predefined schema category group IDs
const (
	GroupCatKeyspaces = "cat:keyspaces"
	GroupCatTables    = "cat:tables"
	GroupCatColumns   = "cat:columns"
)

// Predefined type category group IDs
const (
	GroupCatNumeric    = "cat:numeric"
	GroupCatText       = "cat:text"
	GroupCatTemporal   = "cat:temporal"
	GroupCatCollType   = "cat:collection"
	GroupCatOtherTypes = "cat:other"
)

// KeyTypeGroups contains predefined groups for column key types
var KeyTypeGroups = map[string]CompletionGroup{
	GroupKeyPartition: {
		ID:       GroupKeyPartition,
		Kind:     GroupKindKeyType,
		Label:    "Partition Key",
		Icon:     "★",
		Priority: 1,
	},
	GroupKeyClustering: {
		ID:       GroupKeyClustering,
		Kind:     GroupKindKeyType,
		Label:    "Clustering Key",
		Icon:     "↓",
		Priority: 2,
	},
	GroupKeyStatic: {
		ID:       GroupKeyStatic,
		Kind:     GroupKindKeyType,
		Label:    "Static",
		Icon:     "◆",
		Priority: 3,
	},
	GroupKeyRegular: {
		ID:       GroupKeyRegular,
		Kind:     GroupKindKeyType,
		Label:    "Regular",
		Icon:     "",
		Priority: 4,
	},
}

// FunctionCategoryGroups contains predefined groups for function categories
var FunctionCategoryGroups = map[string]CompletionGroup{
	GroupCatAggregate: {
		ID:       GroupCatAggregate,
		Kind:     GroupKindCategory,
		Label:    "Aggregate",
		Icon:     "Σ",
		Priority: 1,
	},
	GroupCatTime: {
		ID:       GroupCatTime,
		Kind:     GroupKindCategory,
		Label:    "Time/Date",
		Icon:     "⏱",
		Priority: 2,
	},
	GroupCatUUID: {
		ID:       GroupCatUUID,
		Kind:     GroupKindCategory,
		Label:    "UUID",
		Icon:     "#",
		Priority: 3,
	},
	GroupCatConversion: {
		ID:       GroupCatConversion,
		Kind:     GroupKindCategory,
		Label:    "Conversion",
		Icon:     "→",
		Priority: 4,
	},
	GroupCatCollection: {
		ID:       GroupCatCollection,
		Kind:     GroupKindCategory,
		Label:    "Collection",
		Icon:     "[]",
		Priority: 5,
	},
	GroupCatBlob: {
		ID:       GroupCatBlob,
		Kind:     GroupKindCategory,
		Label:    "Blob",
		Icon:     "◫",
		Priority: 6,
	},
	GroupCatMath: {
		ID:       GroupCatMath,
		Kind:     GroupKindCategory,
		Label:    "Math",
		Icon:     "±",
		Priority: 7,
	},
	GroupCatString: {
		ID:       GroupCatString,
		Kind:     GroupKindCategory,
		Label:    "String",
		Icon:     "\"",
		Priority: 8,
	},
	GroupCatOther: {
		ID:       GroupCatOther,
		Kind:     GroupKindCategory,
		Label:    "Other",
		Icon:     "",
		Priority: 99,
	},
}

// TypeCategoryGroups contains predefined groups for CQL type categories
var TypeCategoryGroups = map[string]CompletionGroup{
	GroupCatNumeric: {
		ID:       GroupCatNumeric,
		Kind:     GroupKindCategory,
		Label:    "Numeric",
		Icon:     "#",
		Priority: 1,
	},
	GroupCatText: {
		ID:       GroupCatText,
		Kind:     GroupKindCategory,
		Label:    "Text",
		Icon:     "\"",
		Priority: 2,
	},
	GroupCatTemporal: {
		ID:       GroupCatTemporal,
		Kind:     GroupKindCategory,
		Label:    "Temporal",
		Icon:     "⏱",
		Priority: 3,
	},
	GroupCatCollType: {
		ID:       GroupCatCollType,
		Kind:     GroupKindCategory,
		Label:    "Collection",
		Icon:     "[]",
		Priority: 4,
	},
	GroupCatOtherTypes: {
		ID:       GroupCatOtherTypes,
		Kind:     GroupKindCategory,
		Label:    "Other",
		Icon:     "",
		Priority: 99,
	},
}

// SchemaCategoryGroups contains predefined groups for schema object categories
// Note: Columns has priority 0 to appear before function categories (which start at 1)
var SchemaCategoryGroups = map[string]CompletionGroup{
	GroupCatKeyspaces: {
		ID:       GroupCatKeyspaces,
		Kind:     GroupKindCategory,
		Label:    "Keyspaces",
		Icon:     "⬡",
		Priority: -2,
	},
	GroupCatTables: {
		ID:       GroupCatTables,
		Kind:     GroupKindCategory,
		Label:    "Tables",
		Icon:     "T",
		Priority: -1,
	},
	GroupCatColumns: {
		ID:       GroupCatColumns,
		Kind:     GroupKindCategory,
		Label:    "Columns",
		Icon:     "C",
		Priority: 0,
	},
}

// FunctionToCategory maps function names to their category group ID
var FunctionToCategory = map[string]string{
	// Aggregate functions
	"count": GroupCatAggregate, "sum": GroupCatAggregate, "avg": GroupCatAggregate,
	"min": GroupCatAggregate, "max": GroupCatAggregate,

	// Time/Date functions
	"now": GroupCatTime, "currenttimestamp": GroupCatTime, "currentdate": GroupCatTime,
	"currenttime": GroupCatTime, "currenttimeuuid": GroupCatTime,
	"dateof": GroupCatTime, "unixtimestampof": GroupCatTime,
	"todate": GroupCatTime, "totimestamp": GroupCatTime, "tounixtime": GroupCatTime,
	"mintimeuuid": GroupCatTime, "maxtimeuuid": GroupCatTime,

	// UUID functions
	"uuid": GroupCatUUID, "timeuuid": GroupCatUUID,

	// Conversion functions
	"cast": GroupCatConversion, "typeof": GroupCatConversion,
	"tojson": GroupCatConversion, "fromjson": GroupCatConversion,
	"asciiblobas": GroupCatConversion, "bigintasblog": GroupCatConversion,
	"blobasascii": GroupCatConversion, "blobasbigint": GroupCatConversion,
	"blobasboolean": GroupCatConversion, "blobascounter": GroupCatConversion,
	"blobasdecimal": GroupCatConversion, "blobasdouble": GroupCatConversion,
	"blobasfloat": GroupCatConversion, "blobasint": GroupCatConversion,
	"blobastext": GroupCatConversion, "blobastimestamp": GroupCatConversion,
	"blobasuuid": GroupCatConversion, "blobasvarchar": GroupCatConversion,
	"blobasvarint": GroupCatConversion,

	// Collection functions
	"ttl": GroupCatCollection, "writetime": GroupCatCollection,

	// Blob functions
	"blobastype": GroupCatBlob, "typeasblob": GroupCatBlob, "textasblob": GroupCatBlob,

	// Token (special)
	"token": GroupCatOther,
}

// TypeToCategory maps CQL types to their category group ID
var TypeToCategory = map[string]string{
	// Numeric
	"int": GroupCatNumeric, "bigint": GroupCatNumeric, "smallint": GroupCatNumeric,
	"tinyint": GroupCatNumeric, "varint": GroupCatNumeric, "decimal": GroupCatNumeric,
	"float": GroupCatNumeric, "double": GroupCatNumeric, "counter": GroupCatNumeric,

	// Text
	"text": GroupCatText, "varchar": GroupCatText, "ascii": GroupCatText,

	// Temporal
	"timestamp": GroupCatTemporal, "date": GroupCatTemporal, "time": GroupCatTemporal,
	"timeuuid": GroupCatTemporal, "duration": GroupCatTemporal,

	// Collections
	"list": GroupCatCollType, "set": GroupCatCollType, "map": GroupCatCollType,
	"tuple": GroupCatCollType, "frozen": GroupCatCollType,

	// Other
	"uuid": GroupCatOtherTypes, "boolean": GroupCatOtherTypes, "blob": GroupCatOtherTypes,
	"inet": GroupCatOtherTypes,
}

// KeyspaceGroupID returns the group ID for a keyspace
func KeyspaceGroupID(keyspace string) string {
	return GroupPrefixKeyspace + keyspace
}

// SourceGroupID returns the group ID for a source table
func SourceGroupID(table string) string {
	return GroupPrefixSource + table
}

// NewKeyspaceGroup creates a group for a keyspace
func NewKeyspaceGroup(keyspace string, priority int) CompletionGroup {
	return CompletionGroup{
		ID:       KeyspaceGroupID(keyspace),
		Kind:     GroupKindKeyspace,
		Label:    keyspace,
		Icon:     "⬡",
		Priority: priority,
	}
}

// NewSourceGroup creates a group for a source table
func NewSourceGroup(table string, priority int) CompletionGroup {
	return CompletionGroup{
		ID:       SourceGroupID(table),
		Kind:     GroupKindSource,
		Label:    table,
		Icon:     "T",
		Priority: priority,
	}
}

// GroupRegistry collects groups as completions are generated
type GroupRegistry struct {
	groups map[string]CompletionGroup
}

// NewGroupRegistry creates a new group registry
func NewGroupRegistry() *GroupRegistry {
	return &GroupRegistry{
		groups: make(map[string]CompletionGroup),
	}
}

// Register adds a group to the registry (idempotent)
func (r *GroupRegistry) Register(group CompletionGroup) {
	if _, exists := r.groups[group.ID]; !exists {
		r.groups[group.ID] = group
	}
}

// RegisterKeyspace registers a keyspace group
func (r *GroupRegistry) RegisterKeyspace(keyspace string, priority int) string {
	id := KeyspaceGroupID(keyspace)
	r.Register(NewKeyspaceGroup(keyspace, priority))
	return id
}

// RegisterSource registers a source table group
func (r *GroupRegistry) RegisterSource(table string, priority int) string {
	id := SourceGroupID(table)
	r.Register(NewSourceGroup(table, priority))
	return id
}

// RegisterKeyType registers a key type group and returns its ID
func (r *GroupRegistry) RegisterKeyType(keyType string) string {
	var id string
	switch keyType {
	case "partition":
		id = GroupKeyPartition
	case "clustering":
		id = GroupKeyClustering
	case "static":
		id = GroupKeyStatic
	default:
		id = GroupKeyRegular
	}
	if g, ok := KeyTypeGroups[id]; ok {
		r.Register(g)
	}
	return id
}

// RegisterFunctionCategory registers a function category group and returns its ID
func (r *GroupRegistry) RegisterFunctionCategory(funcName string) string {
	id, ok := FunctionToCategory[funcName]
	if !ok {
		id = GroupCatOther
	}
	if g, ok := FunctionCategoryGroups[id]; ok {
		r.Register(g)
	}
	return id
}

// RegisterTypeCategory registers a type category group and returns its ID
func (r *GroupRegistry) RegisterTypeCategory(typeName string) string {
	id, ok := TypeToCategory[typeName]
	if !ok {
		id = GroupCatOtherTypes
	}
	if g, ok := TypeCategoryGroups[id]; ok {
		r.Register(g)
	}
	return id
}

// RegisterSchemaCategory registers a schema category group (keyspaces, tables) and returns its ID
func (r *GroupRegistry) RegisterSchemaCategory(categoryID string) string {
	if g, ok := SchemaCategoryGroups[categoryID]; ok {
		r.Register(g)
	}
	return categoryID
}

// Groups returns all registered groups as a slice, sorted by priority
func (r *GroupRegistry) Groups() []CompletionGroup {
	result := make([]CompletionGroup, 0, len(r.groups))
	for _, g := range r.groups {
		result = append(result, g)
	}
	// Sort by priority (simple bubble sort for small lists)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Priority < result[i].Priority {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

// AddGroups is a helper to append group IDs to an item
func AddGroups(item *CompletionItem, groupIDs ...string) {
	for _, id := range groupIDs {
		// Avoid duplicates
		found := false
		for _, existing := range item.Groups {
			if existing == id {
				found = true
				break
			}
		}
		if !found {
			item.Groups = append(item.Groups, id)
		}
	}
}

// Debug helper
func (r *GroupRegistry) String() string {
	return fmt.Sprintf("GroupRegistry{%d groups}", len(r.groups))
}
