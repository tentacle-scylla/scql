package types

// StatementType represents the type of CQL statement
type StatementType int

const (
	StatementUnknown StatementType = iota
	StatementSelect
	StatementInsert
	StatementUpdate
	StatementDelete
	StatementBatch
	StatementCreateKeyspace
	StatementAlterKeyspace
	StatementDropKeyspace
	StatementCreateTable
	StatementAlterTable
	StatementDropTable
	StatementTruncate
	StatementCreateIndex
	StatementDropIndex
	StatementCreateMaterializedView
	StatementAlterMaterializedView
	StatementDropMaterializedView
	StatementCreateType
	StatementAlterType
	StatementDropType
	StatementCreateFunction
	StatementDropFunction
	StatementCreateAggregate
	StatementDropAggregate
	StatementCreateTrigger
	StatementDropTrigger
	StatementCreateRole
	StatementAlterRole
	StatementDropRole
	StatementCreateUser
	StatementAlterUser
	StatementDropUser
	StatementGrant
	StatementRevoke
	StatementListRoles
	StatementListPermissions
	StatementUse
	StatementPruneMaterializedView
)

// String returns the string representation of the statement type
func (s StatementType) String() string {
	switch s {
	case StatementSelect:
		return "SELECT"
	case StatementInsert:
		return "INSERT"
	case StatementUpdate:
		return "UPDATE"
	case StatementDelete:
		return "DELETE"
	case StatementBatch:
		return "BATCH"
	case StatementCreateKeyspace:
		return "CREATE KEYSPACE"
	case StatementAlterKeyspace:
		return "ALTER KEYSPACE"
	case StatementDropKeyspace:
		return "DROP KEYSPACE"
	case StatementCreateTable:
		return "CREATE TABLE"
	case StatementAlterTable:
		return "ALTER TABLE"
	case StatementDropTable:
		return "DROP TABLE"
	case StatementTruncate:
		return "TRUNCATE"
	case StatementCreateIndex:
		return "CREATE INDEX"
	case StatementDropIndex:
		return "DROP INDEX"
	case StatementCreateMaterializedView:
		return "CREATE MATERIALIZED VIEW"
	case StatementAlterMaterializedView:
		return "ALTER MATERIALIZED VIEW"
	case StatementDropMaterializedView:
		return "DROP MATERIALIZED VIEW"
	case StatementCreateType:
		return "CREATE TYPE"
	case StatementAlterType:
		return "ALTER TYPE"
	case StatementDropType:
		return "DROP TYPE"
	case StatementCreateFunction:
		return "CREATE FUNCTION"
	case StatementDropFunction:
		return "DROP FUNCTION"
	case StatementCreateAggregate:
		return "CREATE AGGREGATE"
	case StatementDropAggregate:
		return "DROP AGGREGATE"
	case StatementCreateTrigger:
		return "CREATE TRIGGER"
	case StatementDropTrigger:
		return "DROP TRIGGER"
	case StatementCreateRole:
		return "CREATE ROLE"
	case StatementAlterRole:
		return "ALTER ROLE"
	case StatementDropRole:
		return "DROP ROLE"
	case StatementCreateUser:
		return "CREATE USER"
	case StatementAlterUser:
		return "ALTER USER"
	case StatementDropUser:
		return "DROP USER"
	case StatementGrant:
		return "GRANT"
	case StatementRevoke:
		return "REVOKE"
	case StatementListRoles:
		return "LIST ROLES"
	case StatementListPermissions:
		return "LIST PERMISSIONS"
	case StatementUse:
		return "USE"
	case StatementPruneMaterializedView:
		return "PRUNE MATERIALIZED VIEW"
	default:
		return "UNKNOWN"
	}
}

// IsDML returns true if the statement is a Data Manipulation Language statement
func (s StatementType) IsDML() bool {
	switch s {
	case StatementSelect, StatementInsert, StatementUpdate, StatementDelete, StatementBatch:
		return true
	default:
		return false
	}
}

// IsDDL returns true if the statement is a Data Definition Language statement
func (s StatementType) IsDDL() bool {
	switch s {
	case StatementCreateKeyspace, StatementAlterKeyspace, StatementDropKeyspace,
		StatementCreateTable, StatementAlterTable, StatementDropTable, StatementTruncate,
		StatementCreateIndex, StatementDropIndex,
		StatementCreateMaterializedView, StatementAlterMaterializedView, StatementDropMaterializedView,
		StatementCreateType, StatementAlterType, StatementDropType,
		StatementCreateFunction, StatementDropFunction,
		StatementCreateAggregate, StatementDropAggregate,
		StatementCreateTrigger, StatementDropTrigger:
		return true
	default:
		return false
	}
}

// IsDCL returns true if the statement is a Data Control Language statement
func (s StatementType) IsDCL() bool {
	switch s {
	case StatementCreateRole, StatementAlterRole, StatementDropRole,
		StatementCreateUser, StatementAlterUser, StatementDropUser,
		StatementGrant, StatementRevoke, StatementListRoles, StatementListPermissions:
		return true
	default:
		return false
	}
}
