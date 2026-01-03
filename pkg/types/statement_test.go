package types

import "testing"

func TestStatementTypeString(t *testing.T) {
	tests := []struct {
		stmtType StatementType
		want     string
	}{
		{StatementSelect, "SELECT"},
		{StatementInsert, "INSERT"},
		{StatementUpdate, "UPDATE"},
		{StatementDelete, "DELETE"},
		{StatementBatch, "BATCH"},
		{StatementCreateTable, "CREATE TABLE"},
		{StatementDropTable, "DROP TABLE"},
		{StatementCreateKeyspace, "CREATE KEYSPACE"},
		{StatementUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.stmtType.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatementTypeCategories(t *testing.T) {
	tests := []struct {
		stmtType StatementType
		isDML    bool
		isDDL    bool
		isDCL    bool
	}{
		{StatementSelect, true, false, false},
		{StatementInsert, true, false, false},
		{StatementUpdate, true, false, false},
		{StatementDelete, true, false, false},
		{StatementBatch, true, false, false},
		{StatementCreateTable, false, true, false},
		{StatementAlterTable, false, true, false},
		{StatementDropTable, false, true, false},
		{StatementTruncate, false, true, false},
		{StatementCreateKeyspace, false, true, false},
		{StatementCreateIndex, false, true, false},
		{StatementCreateMaterializedView, false, true, false},
		{StatementCreateType, false, true, false},
		{StatementCreateFunction, false, true, false},
		{StatementCreateAggregate, false, true, false},
		{StatementCreateTrigger, false, true, false},
		{StatementCreateRole, false, false, true},
		{StatementAlterRole, false, false, true},
		{StatementDropRole, false, false, true},
		{StatementCreateUser, false, false, true},
		{StatementGrant, false, false, true},
		{StatementRevoke, false, false, true},
		{StatementListRoles, false, false, true},
		{StatementListPermissions, false, false, true},
		{StatementUnknown, false, false, false},
		{StatementUse, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.stmtType.String(), func(t *testing.T) {
			if got := tt.stmtType.IsDML(); got != tt.isDML {
				t.Errorf("IsDML() = %v, want %v", got, tt.isDML)
			}
			if got := tt.stmtType.IsDDL(); got != tt.isDDL {
				t.Errorf("IsDDL() = %v, want %v", got, tt.isDDL)
			}
			if got := tt.stmtType.IsDCL(); got != tt.isDCL {
				t.Errorf("IsDCL() = %v, want %v", got, tt.isDCL)
			}
		})
	}
}
