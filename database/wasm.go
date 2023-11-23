package database

import (
	"fmt"

	dbtypes "github.com/forbole/bdjuno/v4/database/types"
	"github.com/forbole/bdjuno/v4/types"
)

// SaveWasmParams allows to store the wasm params
func (db *Db) SaveWasmParams(params types.WasmParams) error {
	stmt := `
INSERT INTO wasm_params(code_upload_access, instantiate_default_permission, height) 
VALUES ($1, $2, $3) 
ON CONFLICT (one_row_id) DO UPDATE 
	SET code_upload_access = excluded.code_upload_access, 
		instantiate_default_permission = excluded.instantiate_default_permission, 
WHERE wasm_params.height <= excluded.height
`
	accessConfig := dbtypes.NewDbAccessConfig(params.CodeUploadAccess)
	cfgValue, _ := accessConfig.Value()

	_, err := db.SQL.Exec(stmt,
		cfgValue, params.InstantiateDefaultPermission, params.Height,
	)
	if err != nil {
		return fmt.Errorf("error while saving wasm params: %s", err)
	}

	return nil
}
