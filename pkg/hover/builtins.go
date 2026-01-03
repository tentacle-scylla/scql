package hover

import "strings"

// KeywordInfo contains hover documentation for a CQL keyword.
type KeywordInfo struct {
	Name        string
	Description string
	Syntax      string
}

// FunctionInfo contains hover documentation for a CQL function.
type FunctionInfo struct {
	Name        string
	Signature   string
	ReturnType  string
	Description string
}

// TypeInfo contains hover documentation for a CQL type.
type TypeInfo struct {
	Name        string
	Description string
	Size        string
}

// Keywords is a map of CQL keywords to their documentation.
var Keywords = map[string]*KeywordInfo{
	// Statement keywords
	"SELECT":   {Name: "SELECT", Description: "Retrieves data from one or more columns", Syntax: "SELECT columns FROM table [WHERE conditions]"},
	"INSERT":   {Name: "INSERT", Description: "Inserts a new row into a table", Syntax: "INSERT INTO table (columns) VALUES (values)"},
	"UPDATE":   {Name: "UPDATE", Description: "Modifies existing rows in a table", Syntax: "UPDATE table SET column = value WHERE conditions"},
	"DELETE":   {Name: "DELETE", Description: "Removes rows from a table", Syntax: "DELETE FROM table WHERE conditions"},
	"CREATE":   {Name: "CREATE", Description: "Creates a new schema object (table, keyspace, index, etc.)", Syntax: "CREATE TABLE|KEYSPACE|INDEX ..."},
	"ALTER":    {Name: "ALTER", Description: "Modifies a schema object", Syntax: "ALTER TABLE|KEYSPACE ..."},
	"DROP":     {Name: "DROP", Description: "Removes a schema object", Syntax: "DROP TABLE|KEYSPACE|INDEX ..."},
	"TRUNCATE": {Name: "TRUNCATE", Description: "Removes all data from a table", Syntax: "TRUNCATE TABLE table_name"},
	"USE":      {Name: "USE", Description: "Sets the current keyspace for the session", Syntax: "USE keyspace_name"},
	"DESCRIBE": {Name: "DESCRIBE", Description: "Displays schema information", Syntax: "DESCRIBE TABLE|KEYSPACE name"},
	"DESC":     {Name: "DESC", Description: "Displays schema information (alias for DESCRIBE); also descending sort order", Syntax: "DESC TABLE|KEYSPACE name | ORDER BY column DESC"},
	"BEGIN":    {Name: "BEGIN", Description: "Starts a batch statement", Syntax: "BEGIN [UNLOGGED|COUNTER] BATCH ... APPLY BATCH"},
	"APPLY":    {Name: "APPLY", Description: "Ends and executes a batch statement", Syntax: "APPLY BATCH"},
	"BATCH":    {Name: "BATCH", Description: "Groups multiple statements for atomic execution", Syntax: "BEGIN BATCH ... APPLY BATCH"},
	"GRANT":    {Name: "GRANT", Description: "Grants permissions to a role", Syntax: "GRANT permission ON resource TO role"},
	"REVOKE":   {Name: "REVOKE", Description: "Revokes permissions from a role", Syntax: "REVOKE permission ON resource FROM role"},
	"LIST":     {Name: "LIST", Description: "Lists roles or permissions; also ordered collection type", Syntax: "LIST ROLES|PERMISSIONS | list<type>"},

	// Clause keywords
	"FROM":            {Name: "FROM", Description: "Specifies the source table", Syntax: "FROM [keyspace.]table"},
	"WHERE":           {Name: "WHERE", Description: "Filters rows based on conditions", Syntax: "WHERE condition [AND condition ...]"},
	"AND":             {Name: "AND", Description: "Combines conditions (all must be true)", Syntax: "condition AND condition"},
	"OR":              {Name: "OR", Description: "Combines conditions (any can be true)", Syntax: "condition OR condition"},
	"IN":              {Name: "IN", Description: "Matches any value in a list", Syntax: "column IN (value1, value2, ...)"},
	"ORDER":           {Name: "ORDER", Description: "Used with BY to sort results", Syntax: "ORDER BY column [ASC|DESC]"},
	"BY":              {Name: "BY", Description: "Used with ORDER or GROUP to specify columns", Syntax: "ORDER BY column | GROUP BY column"},
	"GROUP":           {Name: "GROUP", Description: "Groups rows by column values", Syntax: "GROUP BY column"},
	"LIMIT":           {Name: "LIMIT", Description: "Limits the number of returned rows", Syntax: "LIMIT count"},
	"ALLOW":           {Name: "ALLOW", Description: "Used with FILTERING to enable inefficient queries", Syntax: "ALLOW FILTERING"},
	"FILTERING":       {Name: "FILTERING", Description: "Enables queries that may scan many rows", Syntax: "ALLOW FILTERING"},
	"SET":             {Name: "SET", Description: "Specifies columns to update", Syntax: "SET column = value [, column = value ...]"},
	"INTO":            {Name: "INTO", Description: "Specifies the target table for INSERT", Syntax: "INSERT INTO table"},
	"VALUES":          {Name: "VALUES", Description: "Specifies values to insert", Syntax: "VALUES (value1, value2, ...)"},
	"IF":              {Name: "IF", Description: "Conditional execution (lightweight transaction)", Syntax: "IF [NOT] EXISTS | IF condition"},
	"EXISTS":          {Name: "EXISTS", Description: "Checks if row exists", Syntax: "IF EXISTS | IF NOT EXISTS"},
	"NOT":             {Name: "NOT", Description: "Negation operator", Syntax: "IF NOT EXISTS | IS NOT NULL"},
	"USING":           {Name: "USING", Description: "Specifies TTL or timestamp", Syntax: "USING TTL seconds | USING TIMESTAMP microseconds"},
	"TTL":             {Name: "TTL", Description: "Time-to-live in seconds", Syntax: "USING TTL seconds"},
	"TIMESTAMP":       {Name: "TIMESTAMP", Description: "Write timestamp in microseconds", Syntax: "USING TIMESTAMP microseconds"},
	"ASC":             {Name: "ASC", Description: "Ascending sort order", Syntax: "ORDER BY column ASC"},
	"CONTAINS":        {Name: "CONTAINS", Description: "Checks if collection contains a value", Syntax: "column CONTAINS value | column CONTAINS KEY key"},
	"KEY":             {Name: "KEY", Description: "Used with CONTAINS to check map keys", Syntax: "CONTAINS KEY key_value"},
	"NULL":            {Name: "NULL", Description: "Represents an absent value", Syntax: "column IS NULL | column IS NOT NULL"},
	"PRIMARY":         {Name: "PRIMARY", Description: "Defines the primary key", Syntax: "PRIMARY KEY ((partition_key), clustering_key)"},
	"PARTITION":       {Name: "PARTITION", Description: "Part of primary key that determines data distribution", Syntax: "PRIMARY KEY ((partition_key), ...)"},
	"CLUSTERING":      {Name: "CLUSTERING", Description: "Part of primary key that determines row ordering", Syntax: "PRIMARY KEY ((pk), clustering_key)"},
	"STATIC":          {Name: "STATIC", Description: "Column shared by all rows with same partition key", Syntax: "column_name type STATIC"},
	"WITH":            {Name: "WITH", Description: "Specifies table or keyspace options", Syntax: "WITH option = value [AND option = value ...]"},
	"REPLICATION":     {Name: "REPLICATION", Description: "Specifies keyspace replication strategy", Syntax: "WITH replication = {'class': 'strategy', ...}"},
	"COMPACT":         {Name: "COMPACT", Description: "Deprecated storage format", Syntax: "WITH COMPACT STORAGE"},
	"STORAGE":         {Name: "STORAGE", Description: "Used with COMPACT (deprecated)", Syntax: "WITH COMPACT STORAGE"},
	"INDEX":           {Name: "INDEX", Description: "Secondary index on a column", Syntax: "CREATE INDEX ON table (column)"},
	"MATERIALIZED":    {Name: "MATERIALIZED", Description: "Used with VIEW to create materialized views", Syntax: "CREATE MATERIALIZED VIEW"},
	"VIEW":            {Name: "VIEW", Description: "A view that stores query results", Syntax: "CREATE MATERIALIZED VIEW name AS SELECT ..."},
	"TABLE":           {Name: "TABLE", Description: "A collection of rows organized by primary key", Syntax: "CREATE TABLE name (columns, PRIMARY KEY (...))"},
	"KEYSPACE":        {Name: "KEYSPACE", Description: "A namespace for tables (similar to database)", Syntax: "CREATE KEYSPACE name WITH replication = {...}"},
	"TYPE":            {Name: "TYPE", Description: "User-defined type (UDT)", Syntax: "CREATE TYPE name (field1 type1, ...)"},
	"FUNCTION":        {Name: "FUNCTION", Description: "User-defined function (UDF)", Syntax: "CREATE FUNCTION name (params) ..."},
	"AGGREGATE":       {Name: "AGGREGATE", Description: "User-defined aggregate function (UDA)", Syntax: "CREATE AGGREGATE name (type) ..."},
	"ROLE":            {Name: "ROLE", Description: "A named collection of permissions", Syntax: "CREATE ROLE name"},
	"USER":            {Name: "USER", Description: "A user (deprecated, use ROLE)", Syntax: "CREATE USER name"},
	"JSON":            {Name: "JSON", Description: "JSON format for INSERT or SELECT", Syntax: "INSERT JSON '{}' | SELECT JSON *"},
	"DISTINCT":        {Name: "DISTINCT", Description: "Returns unique partition keys only", Syntax: "SELECT DISTINCT partition_key FROM table"},
	"COUNT":           {Name: "COUNT", Description: "Aggregate function to count rows", Syntax: "SELECT COUNT(*) FROM table"},
	"AS":              {Name: "AS", Description: "Alias for column or table", Syntax: "SELECT column AS alias"},
	"TOKEN":           {Name: "TOKEN", Description: "Partition token function", Syntax: "WHERE TOKEN(pk) > TOKEN(value)"},
	"WRITETIME":       {Name: "WRITETIME", Description: "Returns write timestamp of a column", Syntax: "SELECT WRITETIME(column) FROM table"},
	"PER":             {Name: "PER", Description: "Used with PARTITION LIMIT", Syntax: "PER PARTITION LIMIT n"},
	"UNLOGGED":        {Name: "UNLOGGED", Description: "Batch without atomicity guarantee", Syntax: "BEGIN UNLOGGED BATCH"},
	"COUNTER":         {Name: "COUNTER", Description: "Batch for counter updates", Syntax: "BEGIN COUNTER BATCH"},
	"FROZEN":          {Name: "FROZEN", Description: "Immutable collection or UDT", Syntax: "frozen<collection_type>"},
	"TUPLE":           {Name: "TUPLE", Description: "Fixed-length sequence of typed values", Syntax: "tuple<type1, type2, ...>"},
	"MAP":             {Name: "MAP", Description: "Key-value collection", Syntax: "map<key_type, value_type>"},
}

// Functions is a map of CQL functions to their documentation.
var Functions = map[string]*FunctionInfo{
	// UUID and time functions
	"uuid":             {Name: "uuid", Signature: "uuid()", ReturnType: "uuid", Description: "Generates a random Type 4 UUID"},
	"now":              {Name: "now", Signature: "now()", ReturnType: "timeuuid", Description: "Returns a new unique timeuuid (Type 1 UUID) based on current time"},
	"timeuuid":         {Name: "timeuuid", Signature: "timeuuid()", ReturnType: "timeuuid", Description: "Creates a Type 1 UUID from a timestamp"},
	"currenttimestamp": {Name: "currentTimestamp", Signature: "currentTimestamp()", ReturnType: "timestamp", Description: "Returns the current timestamp"},
	"currentdate":      {Name: "currentDate", Signature: "currentDate()", ReturnType: "date", Description: "Returns the current date"},
	"currenttime":      {Name: "currentTime", Signature: "currentTime()", ReturnType: "time", Description: "Returns the current time of day"},
	"currenttimeuuid":  {Name: "currentTimeUUID", Signature: "currentTimeUUID()", ReturnType: "timeuuid", Description: "Returns a timeuuid for the current time"},

	// Token function
	"token": {Name: "token", Signature: "token(partition_key)", ReturnType: "bigint", Description: "Returns the token value for a partition key, used for token range queries"},

	// Time conversion functions
	"todate":           {Name: "toDate", Signature: "toDate(timeuuid|timestamp)", ReturnType: "date", Description: "Converts a timeuuid or timestamp to a date"},
	"totimestamp":      {Name: "toTimestamp", Signature: "toTimestamp(timeuuid|date)", ReturnType: "timestamp", Description: "Converts a timeuuid or date to a timestamp"},
	"tounixtimestamp":  {Name: "toUnixTimestamp", Signature: "toUnixTimestamp(timeuuid|timestamp|date)", ReturnType: "bigint", Description: "Converts to Unix timestamp in milliseconds"},
	"dateof":           {Name: "dateOf", Signature: "dateOf(timeuuid)", ReturnType: "timestamp", Description: "Extracts the timestamp from a timeuuid (deprecated, use toTimestamp)"},
	"unixtimestampof":  {Name: "unixTimestampOf", Signature: "unixTimestampOf(timeuuid)", ReturnType: "bigint", Description: "Extracts Unix timestamp from timeuuid (deprecated)"},
	"mintimeuuid":      {Name: "minTimeuuid", Signature: "minTimeuuid(timestamp)", ReturnType: "timeuuid", Description: "Returns the smallest possible timeuuid for a given timestamp"},
	"maxtimeuuid":      {Name: "maxTimeuuid", Signature: "maxTimeuuid(timestamp)", ReturnType: "timeuuid", Description: "Returns the largest possible timeuuid for a given timestamp"},

	// Blob conversion functions
	"blobastext":    {Name: "blobAsText", Signature: "blobAsText(blob)", ReturnType: "text", Description: "Converts a blob to UTF-8 text"},
	"textasblob":    {Name: "textAsBlob", Signature: "textAsBlob(text)", ReturnType: "blob", Description: "Converts text to a blob"},
	"blobasint":     {Name: "blobAsInt", Signature: "blobAsInt(blob)", ReturnType: "int", Description: "Converts a blob to a 32-bit integer"},
	"intasblob":     {Name: "intAsBlob", Signature: "intAsBlob(int)", ReturnType: "blob", Description: "Converts a 32-bit integer to a blob"},
	"blobasbigint":  {Name: "blobAsBigint", Signature: "blobAsBigint(blob)", ReturnType: "bigint", Description: "Converts a blob to a 64-bit integer"},
	"bigintasblob":  {Name: "bigintAsBlob", Signature: "bigintAsBlob(bigint)", ReturnType: "blob", Description: "Converts a 64-bit integer to a blob"},
	"blobasascii":   {Name: "blobAsAscii", Signature: "blobAsAscii(blob)", ReturnType: "ascii", Description: "Converts a blob to ASCII text"},
	"asciiasblob":   {Name: "asciiAsBlob", Signature: "asciiAsBlob(ascii)", ReturnType: "blob", Description: "Converts ASCII text to a blob"},
	"blobasboolean": {Name: "blobAsBoolean", Signature: "blobAsBoolean(blob)", ReturnType: "boolean", Description: "Converts a blob to a boolean"},
	"booleanasblob": {Name: "booleanAsBlob", Signature: "booleanAsBlob(boolean)", ReturnType: "blob", Description: "Converts a boolean to a blob"},
	"blobasdouble":  {Name: "blobAsDouble", Signature: "blobAsDouble(blob)", ReturnType: "double", Description: "Converts a blob to a double"},
	"doubleasblob":  {Name: "doubleAsBlob", Signature: "doubleAsBlob(double)", ReturnType: "blob", Description: "Converts a double to a blob"},
	"blobasfloat":   {Name: "blobAsFloat", Signature: "blobAsFloat(blob)", ReturnType: "float", Description: "Converts a blob to a float"},
	"floatasblob":   {Name: "floatAsBlob", Signature: "floatAsBlob(float)", ReturnType: "blob", Description: "Converts a float to a blob"},
	"blobasinet":    {Name: "blobAsInet", Signature: "blobAsInet(blob)", ReturnType: "inet", Description: "Converts a blob to an inet address"},
	"inetasblob":    {Name: "inetAsBlob", Signature: "inetAsBlob(inet)", ReturnType: "blob", Description: "Converts an inet address to a blob"},
	"blobasuuid":    {Name: "blobAsUuid", Signature: "blobAsUuid(blob)", ReturnType: "uuid", Description: "Converts a blob to a UUID"},
	"uuidasblob":    {Name: "uuidAsBlob", Signature: "uuidAsBlob(uuid)", ReturnType: "blob", Description: "Converts a UUID to a blob"},
	"blobastimeuuid": {Name: "blobAsTimeuuid", Signature: "blobAsTimeuuid(blob)", ReturnType: "timeuuid", Description: "Converts a blob to a timeuuid"},
	"timeuuidasblob": {Name: "timeuuidAsBlob", Signature: "timeuuidAsBlob(timeuuid)", ReturnType: "blob", Description: "Converts a timeuuid to a blob"},
	"blobasvarint":  {Name: "blobAsVarint", Signature: "blobAsVarint(blob)", ReturnType: "varint", Description: "Converts a blob to a varint"},
	"varintasblob":  {Name: "varintAsBlob", Signature: "varintAsBlob(varint)", ReturnType: "blob", Description: "Converts a varint to a blob"},

	// Aggregate functions
	"count": {Name: "count", Signature: "count(*) | count(column)", ReturnType: "bigint", Description: "Counts the number of rows or non-null values"},
	"sum":   {Name: "sum", Signature: "sum(column)", ReturnType: "varies", Description: "Returns the sum of numeric values"},
	"avg":   {Name: "avg", Signature: "avg(column)", ReturnType: "varies", Description: "Returns the average of numeric values"},
	"min":   {Name: "min", Signature: "min(column)", ReturnType: "varies", Description: "Returns the minimum value"},
	"max":   {Name: "max", Signature: "max(column)", ReturnType: "varies", Description: "Returns the maximum value"},

	// Cell metadata functions
	"writetime": {Name: "writetime", Signature: "writetime(column)", ReturnType: "bigint", Description: "Returns the write timestamp of a column in microseconds"},
	"ttl":       {Name: "ttl", Signature: "ttl(column)", ReturnType: "int", Description: "Returns the remaining TTL (time-to-live) of a column in seconds"},

	// JSON functions
	"tojson":   {Name: "toJson", Signature: "toJson(value)", ReturnType: "text", Description: "Converts any CQL value to its JSON representation"},
	"fromjson": {Name: "fromJson", Signature: "fromJson(text)", ReturnType: "varies", Description: "Parses a JSON string to a CQL value"},

	// Type casting
	"cast": {Name: "cast", Signature: "cast(value AS type)", ReturnType: "varies", Description: "Casts a value to another compatible type"},

	// Collection functions
	"collection_count":  {Name: "collection_count", Signature: "collection_count(collection)", ReturnType: "int", Description: "Returns the number of elements in a collection"},
	"collection_max":    {Name: "collection_max", Signature: "collection_max(collection)", ReturnType: "varies", Description: "Returns the maximum element in a collection"},
	"collection_min":    {Name: "collection_min", Signature: "collection_min(collection)", ReturnType: "varies", Description: "Returns the minimum element in a collection"},
}

// Types is a map of CQL types to their documentation.
var Types = map[string]*TypeInfo{
	// Numeric types
	"int":      {Name: "int", Description: "32-bit signed integer", Size: "4 bytes"},
	"bigint":   {Name: "bigint", Description: "64-bit signed integer (long)", Size: "8 bytes"},
	"smallint": {Name: "smallint", Description: "16-bit signed integer", Size: "2 bytes"},
	"tinyint":  {Name: "tinyint", Description: "8-bit signed integer", Size: "1 byte"},
	"varint":   {Name: "varint", Description: "Arbitrary-precision integer", Size: "variable"},
	"float":    {Name: "float", Description: "32-bit IEEE-754 floating point", Size: "4 bytes"},
	"double":   {Name: "double", Description: "64-bit IEEE-754 floating point", Size: "8 bytes"},
	"decimal":  {Name: "decimal", Description: "Arbitrary-precision decimal", Size: "variable"},
	"counter":  {Name: "counter", Description: "Distributed counter (64-bit)", Size: "8 bytes"},

	// String types
	"text":    {Name: "text", Description: "UTF-8 encoded string", Size: "variable"},
	"varchar": {Name: "varchar", Description: "UTF-8 encoded string (alias for text)", Size: "variable"},
	"ascii":   {Name: "ascii", Description: "ASCII encoded string", Size: "variable"},

	// UUID types
	"uuid":     {Name: "uuid", Description: "Type 4 UUID (random)", Size: "16 bytes"},
	"timeuuid": {Name: "timeuuid", Description: "Type 1 UUID (time-based, sortable)", Size: "16 bytes"},

	// Time types
	"timestamp": {Name: "timestamp", Description: "Date and time (millisecond precision)", Size: "8 bytes"},
	"date":      {Name: "date", Description: "Date without time component", Size: "4 bytes"},
	"time":      {Name: "time", Description: "Time without date (nanosecond precision)", Size: "8 bytes"},
	"duration":  {Name: "duration", Description: "Time duration (months, days, nanoseconds)", Size: "variable"},

	// Binary types
	"blob":    {Name: "blob", Description: "Arbitrary bytes (Binary Large OBject)", Size: "variable"},
	"boolean": {Name: "boolean", Description: "Boolean true or false", Size: "1 byte"},
	"inet":    {Name: "inet", Description: "IPv4 or IPv6 address", Size: "4 or 16 bytes"},

	// Collection types
	"list":   {Name: "list", Description: "Ordered collection of elements", Size: "variable"},
	"set":    {Name: "set", Description: "Unordered collection of unique elements", Size: "variable"},
	"map":    {Name: "map", Description: "Collection of key-value pairs", Size: "variable"},
	"tuple":  {Name: "tuple", Description: "Fixed-length sequence of typed values", Size: "variable"},
	"frozen": {Name: "frozen", Description: "Immutable collection or UDT (serialized as a blob)", Size: "variable"},
}

// GetKeywordInfo returns hover info for a keyword.
func GetKeywordInfo(name string) *KeywordInfo {
	return Keywords[strings.ToUpper(name)]
}

// GetFunctionInfo returns hover info for a function.
func GetFunctionInfo(name string) *FunctionInfo {
	return Functions[strings.ToLower(name)]
}

// GetTypeInfo returns hover info for a type.
func GetTypeInfo(name string) *TypeInfo {
	return Types[strings.ToLower(name)]
}

// IsKeyword checks if the name is a known CQL keyword.
func IsKeyword(name string) bool {
	_, ok := Keywords[strings.ToUpper(name)]
	return ok
}

// IsFunction checks if the name is a known CQL function.
func IsFunction(name string) bool {
	_, ok := Functions[strings.ToLower(name)]
	return ok
}

// IsType checks if the name is a known CQL type.
func IsType(name string) bool {
	_, ok := Types[strings.ToLower(name)]
	return ok
}
