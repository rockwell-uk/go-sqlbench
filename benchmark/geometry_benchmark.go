package benchmark

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rockwell-uk/go-progress/progress"
)

func SetupGeomBench(db *sqlx.DB, tableParams string) error {
	// Delete the bench table if it exists
	TeardownGeomBench(db)

	// Create the bench table
	createTableSQL := getTableSQL(benchTableName, tableParams)

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func TeardownGeomBench(db *sqlx.DB) {
	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %v", benchTableName)
	db.MustExec(dropTableSQL)
}

func BmWKBExec(db *sqlx.DB, n int) error {
	var funcName string = "benchmark.BmWKBExec"
	var jobName string = "Geometry Benchmark [BmWKB]"

	var magnitude int = n

	// BmWKBExec Job
	var job progress.ProgressJob = &BmWKBExecJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, n, db)
}

func getTableSQL(fullTableName, tableParams string) string {
	var fields = map[string]string{
		"ID":      "varchar(36) NOT NULL",
		"GRIDREF": "smallint NOT NULL",
	}

	tableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", fullTableName)

	for f, t := range fields {
		tableSQL += fmt.Sprintf("%s %s,", f, t)
	}

	tableSQL += fmt.Sprintf("ogc_geom geometry DEFAULT NULL, PRIMARY KEY (ID, GRIDREF))%v;", tableParams)

	return tableSQL
}
