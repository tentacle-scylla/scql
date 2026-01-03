# scql Code Generator

Generates two outputs from ScyllaDB sources:

1. **`gen/parser/`** - ANTLR4 CQL parser (patched for ScyllaDB extensions)
2. **`gen/cqldata/`** - CQL language data (keywords, functions, types) for IDE completion

## Quick Start

```bash
# Run library tests
go test ./pkg/...

# Run grammar tests (1566 queries)
cd gen/parser/tests && go test -v

# Regenerate everything (requires Java for ANTLR4)
go run ./cmd/bootstrap
```

## Pipeline Overview

```
                              PARSER GENERATION
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Download   │───▶│   Analyze   │───▶│   Patch     │───▶│  Generate   │
│  Grammars   │    │   Diff      │    │   Grammar   │    │  ANTLR4     │
└─────────────┘    └─────────────┘    └─────────────┘    └──────┬──────┘
                                                                │
                         COMPLETION DATA GENERATION             ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Sparse Clone│───▶│   Parse     │───▶│  Generate   │───▶│    Test     │
│  ScyllaDB   │    │   C++ Code  │    │  Go Files   │    │  (1566 CQL) │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

## Generated Outputs

### gen/parser/ - CQL Parser

ANTLR4-generated Go parser supporting all ScyllaDB CQL syntax.

**Source:** Cassandra ANTLR4 grammar + ScyllaDB patches
**Generated files:** `cql_lexer.go`, `cql_parser.go`, `cqlparser_base_listener.go`, etc.

### gen/cqldata/ - CQL Language Data

Extracted from ScyllaDB C++ source code for IDE completion support.

| File | Contents | Source |
|------|----------|--------|
| `gen_keywords.go` | All CQL keywords, categorized (reserved, unreserved, type keywords) | `cql3/Cql.g` |
| `gen_functions.go` | Built-in functions with signatures and return types | `cql3/functions/*.cc` |
| `gen_types.go` | Data types with aliases and compatibility mappings | `types/types.cc` |

The pipeline sparse-clones only the required directories from ScyllaDB (`cql3/` and `types/`) to minimize download size.

---

## Parser Generation

### Stage 1: Download Grammars

Downloads from GitHub:
- `CqlLexer.g4` - Base Cassandra lexer (ANTLR4)
- `CqlParser.g4` - Base Cassandra parser (ANTLR4)
- `scylla_Cql.g` - ScyllaDB grammar (ANTLR3, for reference)

### Stage 2: Analyze Diff

Compares ScyllaDB grammar against base Cassandra to identify:
- Missing keywords
- Different parser rules

Output: `gen/parser/build/analysis/missing_keywords.txt`

### Stage 3: Apply Patches

Patches defined in `codegen/patches/`:

| File | Purpose |
|------|---------|
| `scylla_lexer_keywords.txt` | New keywords (e.g., `K_BYPASS:BYPASS`) |
| `scylla_lexer_rules.txt` | New lexer rules (e.g., `DURATION_LITERAL`) |
| `scylla_lexer_replacements.txt` | Lexer rule replacements |
| `scylla_parser_patches.json` | Parser rule additions/replacements |

### Stage 4: Generate ANTLR4

Runs ANTLR4 to produce Go parser from patched `.g4` files.

---

## Completion Data Generation

### Stage 1: Sparse Clone ScyllaDB

Clones only required directories from ScyllaDB repo:
```
scylladb/
├── cql3/
│   ├── Cql.g           # Grammar (keywords)
│   └── functions/      # Built-in functions
└── types/
    └── types.cc        # Data type definitions
```

### Stage 2: Parse C++ Source

Extracts using regex patterns:
- **Keywords:** From grammar rules in `Cql.g`
- **Functions:** From `add_function()` calls in `functions/*.cc`
- **Types:** From type registration in `types/types.cc`

### Stage 3: Generate Go Files

Outputs to `gen/cqldata/`:
- `gen_keywords.go` - Keyword lists and categories
- `gen_functions.go` - Function definitions with `GenFunctionDef` structs
- `gen_types.go` - Type definitions with `GenTypeDef` structs

---

## Adding New ScyllaDB Features

### 1. Add Test Cases First

Create `gen/parser/tests/queries/extra_features/my_feature.cql`:

```sql
-- Description of the feature
NEW STATEMENT SYNTAX;
ANOTHER VARIANT;
```

### 2. Run Tests

```bash
cd gen/parser/tests && go test -v
```

Note which queries fail.

### 3. Add Lexer Keywords

Edit `codegen/patches/scylla_lexer_keywords.txt`:

```
# My new feature
K_NEWKEYWORD:NEWKEYWORD
```

### 4. Add Parser Patches

Edit `codegen/patches/scylla_parser_patches.json`:

```json
{
  "type": "add_keywords",
  "content": "kwNewKeyword: K_NEWKEYWORD;"
},
{
  "type": "add_rule",
  "after": "existingRule",
  "content": "newRule\n    : kwNewKeyword OBJECT_NAME\n    ;"
},
{
  "type": "add_to_cql_rule",
  "content": "| newRule"
}
```

### 5. Rebuild

```bash
go run ./cmd/bootstrap
```

### Patch Types

| Type | Description |
|------|-------------|
| `add_keywords` | Add keyword wrapper rules |
| `add_rule` | Add new parser rule after specified rule |
| `add_to_rule` | Append alternatives to existing rule |
| `add_to_cql_rule` | Add statement type to main `cql` rule |
| `replace_rule` | Replace entire rule definition |

### Keyword-Identifier Conflicts

If a keyword conflicts with table/column names:

```json
{
  "type": "add_to_rule",
  "rule": "reservedKeywordAsTable",
  "content": "| K_NEWKEYWORD"
}
```

---

## Patch File Formats

### scylla_lexer_keywords.txt

```
# Comments start with #
K_KEYWORD_NAME:LITERAL_VALUE
```

### scylla_lexer_rules.txt

```
# Comments start with #
RULE_NAME=ANTLR_PATTERN
```

Example: `DURATION_LITERAL=[0-9]+('ms'|'s'|'m'|'h'|'d')`

### scylla_parser_patches.json

```json
{
  "description": "ScyllaDB parser grammar patches",
  "version": "5.6",
  "patches": [
    {
      "type": "patch_type",
      "rule": "optional_rule_name",
      "after": "optional_insertion_point",
      "content": "ANTLR grammar content"
    }
  ]
}
```

---

## Directory Structure

```
scql/
├── codegen/                    # This pipeline
│   ├── pipeline/               # Pipeline stages
│   │   ├── download.go         # Download grammars + sparse clone ScyllaDB
│   │   ├── extract.go          # Extract keywords/rules from grammars
│   │   ├── diff.go             # Compare grammars
│   │   ├── patch.go            # Apply patches to grammar
│   │   ├── generate.go         # Run ANTLR4
│   │   ├── cqlextract.go          # Generate completion data
│   │   └── test.go             # Run grammar tests
│   ├── cqlextract/                # Completion data extractor
│   │   ├── grammar.go          # Parse Cql.g for keywords
│   │   ├── functions.go        # Parse functions/*.cc
│   │   ├── types.go            # Parse types/types.cc
│   │   └── output.go           # Generate Go files
│   ├── grammar/                # Grammar utilities
│   ├── patches/                # Patch definitions
│   └── util/                   # File utilities
│
├── gen/                        # Generated code (DO NOT EDIT)
│   ├── parser/                 # ANTLR-generated CQL parser
│   │   ├── build/              # Build artifacts (gitignored)
│   │   │   ├── grammars/       # Downloaded grammars
│   │   │   ├── patched/        # Patched .g4 files
│   │   │   ├── analysis/       # Diff analysis
│   │   │   └── scylladb/       # Sparse-cloned ScyllaDB source
│   │   ├── tests/              # Grammar tests (1566 queries)
│   │   └── *.go                # Generated Go parser
│   └── cqldata/                # Generated CQL language data
│       ├── gen_keywords.go     # Keywords from Cql.g
│       ├── gen_functions.go    # Functions from cql3/functions/
│       └── gen_types.go        # Types from types/types.cc
│
├── pkg/                        # Library (uses gen/)
└── cmd/
    └── bootstrap/              # Entry point: go run ./cmd/bootstrap
```

---

## Troubleshooting

### "rule X redefinition"

Keyword already exists in base lexer. Remove from `scylla_lexer_keywords.txt`.

### "no viable alternative at input 'X'"

Parser doesn't recognize syntax. Check:
1. Keyword added to lexer?
2. Parser rule added?
3. Rule added to `cql` rule?

### Keyword conflicts with table/column names

Add keyword to `reservedKeywordAsTable` or `reservedKeywordAsColumn`.

### "implicit definition of token K_X"

Parser uses undefined keyword. Either:
1. Add to `scylla_lexer_keywords.txt`, or
2. Check if base grammar already defines it

---

## Current Coverage

**1566 queries, 100% passing**

ScyllaDB-specific features: BYPASS CACHE, USING TIMEOUT, GROUP BY, PER PARTITION LIMIT, LIKE, CAST, DURATION, VECTOR, PRUNE MATERIALIZED VIEW, DESCRIBE, Service Levels, ANN/Vector Search, REDUCEFUNC, Auth Extensions, EMPTY collections, Tablets, and more.
