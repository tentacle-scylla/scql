# scql

A Go library for parsing, linting, and formatting CQL (Cassandra Query Language) with full ScyllaDB support.

Built on ANTLR4 with grammar patches for ScyllaDB-specific syntax.

**CQL Reference**: https://docs.scylladb.com/stable/cql/

## Installation

```bash
go get github.com/tentacle-scylla/scql
```

## CLI

```bash
go install github.com/tentacle-scylla/scql/cmd/scql@latest
```

### Lint

```bash
echo "SELECT * FROM users;" | scql lint
scql lint -f queries.cql
scql lint -q -f queries.cql  # quiet, errors only
```

### Format

```bash
# Pretty (default)
echo "select * from users where id=1;" | scql format
# SELECT *
# FROM users
# WHERE id = 1;

# Compact
scql format --compact -f queries.cql

# Lowercase keywords
scql format --lowercase -f queries.cql

# In-place
scql format -w -f queries.cql
```

### Parse

```bash
echo "SELECT * FROM users; INSERT INTO users (id) VALUES (1);" | scql parse
# Statement 1:
#   Type:  SELECT
#   Valid: true
#
# Statement 2:
#   Type:  INSERT
#   Valid: true
```

## Library

```go
import "github.com/tentacle-scylla/scql"

// Parse
result := scql.Parse("SELECT * FROM users WHERE id = 1;")
fmt.Println(result.Type)       // SELECT
fmt.Println(result.IsValid())  // true

// Validate
if scql.IsValid("SELECT * FROM users;") {
    // ...
}

// Lint with suggestions
errors := scql.Lint("SELECT * FORM users;")
if errors.HasErrors() {
    err := errors.First()
    fmt.Println(err.Message)     // "no viable alternative..."
    fmt.Println(err.Suggestion)  // "Did you mean 'FROM'?"
}

// Format
pretty, _ := scql.Pretty("select * from users where id=1;")
compact, _ := scql.Compact("SELECT *\nFROM users;")

// Multiple statements
results := scql.ParseMultiple("SELECT * FROM a; SELECT * FROM b;")
errors := scql.LintMultiple(input)
```

### Custom formatting

```go
opts := scql.FormatOptions{
    Style:             scql.FormatPretty,
    IndentString:      "    ",
    UppercaseKeywords: true,
}
formatted := scql.Format(result, opts)
```

### Sub-packages

For more control:

```go
import (
    "github.com/tentacle-scylla/scql/pkg/parse"
    "github.com/tentacle-scylla/scql/pkg/format"
    "github.com/tentacle-scylla/scql/pkg/lint"
    "github.com/tentacle-scylla/scql/pkg/types"
)
```

## Statement types

All standard CQL statements are detected:

| Category | Statements |
|----------|------------|
| DML | SELECT, INSERT, UPDATE, DELETE, BATCH |
| DDL | CREATE/ALTER/DROP KEYSPACE, TABLE, INDEX, TYPE, FUNCTION, AGGREGATE, TRIGGER, MATERIALIZED VIEW |
| DCL | CREATE/ALTER/DROP ROLE, USER, GRANT, REVOKE, LIST ROLES, LIST PERMISSIONS, LIST USERS |
| QoS | CREATE/ALTER/DROP/ATTACH/DETACH/LIST SERVICE LEVEL |
| Other | USE, TRUNCATE, PRUNE MATERIALIZED VIEW, DESCRIBE |

```go
result := scql.Parse("CREATE TABLE users (...);")
result.Type.IsDDL()  // true
result.Type.IsDML()  // false
```

## ScyllaDB extensions

Full support for ScyllaDB-specific syntax (100% coverage, 1566 test queries):

### Query Extensions
- `BYPASS CACHE`, `USING TIMEOUT`
- `PER PARTITION LIMIT`, `GROUP BY`
- `LIKE` operator, `CAST` function
- `ORDER BY column ANN OF [vector]` (vector search)
- Duration literals (`500ms`, `1s`, `2h`)

### DDL Extensions
- `PRUNE MATERIALIZED VIEW` with `CONCURRENCY` option
- `CREATE CUSTOM INDEX ... USING 'StorageAttachedIndex'` (vector indexes)
- `VECTOR<float, N>`, `DURATION` data types
- `DESCRIBE` / `DESC` statements with `WITH INTERNALS`

### Auth Extensions
- `NOLOGIN`, `HASHED PASSWORD` for roles
- `LIST USERS`, `LIST ALL ROLES OF ... NORECURSIVE`
- `VECTOR_SEARCH_INDEXING` permission

### Service Levels (QoS)
- `CREATE/ALTER/DROP SERVICE LEVEL`
- `ATTACH/DETACH SERVICE LEVEL ... TO/FROM role`
- `LIST SERVICE LEVEL`, `LIST ALL SERVICE LEVELS`
- Properties: `timeout`, `shares`, `workload_type`

### Aggregates
- `REDUCEFUNC` for distributed aggregation
- Optional `FINALFUNC`, `INITCOND`

### Other
- `EMPTY` collection literal
- `writetime()`, `ttl()`, `token()` functions
- Map subscript assignments (`data['key'] = value`)
- Prepared statement placeholders (`?`, `:name`)
- Tablets keyspace option

## Contributing

See [codegen/README.md](codegen/README.md) for development setup and adding new CQL syntax.

## License

MIT
