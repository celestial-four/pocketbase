package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.SystemMigrations.Add(&core.Migration{
		Up: func(txApp core.App) error {
			logsSQL := `
				CREATE TABLE IF NOT EXISTS {{_logs}} (
					[[id]]      TEXT PRIMARY KEY DEFAULT ('r'||lower(hex(randomblob(7)))) NOT NULL,
					[[level]]   INTEGER DEFAULT 0 NOT NULL,
					[[message]] TEXT DEFAULT "" NOT NULL,
					[[data]]    JSON DEFAULT "{}" NOT NULL,
					[[created]] TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL
				);

				CREATE INDEX IF NOT EXISTS idx_logs_level on {{_logs}} ([[level]]);
				CREATE INDEX IF NOT EXISTS idx_logs_message on {{_logs}} ([[message]]);
			`
			// Translate SQL for the current database type
			logsSQL = core.TranslateSQLForDB(logsSQL, txApp.DBType())
			
			_, execErr := txApp.AuxDB().NewQuery(logsSQL).Execute()
			if execErr != nil {
				return execErr
			}
			
			// Create the hour index separately since it needs special handling for PostgreSQL
			var indexSQL string
			if txApp.DBType() == core.DatabaseTypePostgreSQL {
				// For PostgreSQL, use date_trunc with proper column reference
				indexSQL = `CREATE INDEX IF NOT EXISTS idx_logs_created_hour on {{_logs}} (date_trunc('hour', "created"::timestamp))`
			} else {
				indexSQL = `CREATE INDEX IF NOT EXISTS idx_logs_created_hour on {{_logs}} (strftime('%Y-%m-%d %H:00:00', [[created]]))`
			}
			_, execErr = txApp.AuxDB().NewQuery(indexSQL).Execute()

			return execErr
		},
		Down: func(txApp core.App) error {
			_, err := txApp.AuxDB().DropTable("_logs").Execute()
			return err
		},
		ReapplyCondition: func(txApp core.App, runner *core.MigrationsRunner, fileName string) (bool, error) {
			// reapply only if the _logs table doesn't exist
			exists := txApp.AuxHasTable("_logs")
			return !exists, nil
		},
	})
}
