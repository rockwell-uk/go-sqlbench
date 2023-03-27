package benchmark

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rockwell-uk/go-progress/progress"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
	"github.com/twpayne/go-geom/encoding/wkbcommon"
)

type BmWKBExecJob struct{}

func (j *BmWKBExecJob) Setup(jobName string, input interface{}) (*progress.Job, error) {
	if n, ok := input.(int); ok {
		var tasks []*progress.Task
		for i := 0; i <= n; i++ {
			tasks = append(tasks, &progress.Task{
				ID:        strconv.Itoa(i),
				Magnitude: 1,
			})
		}

		job := progress.SetupJob(jobName, tasks)

		return job, nil
	}

	return nil, fmt.Errorf("expected int got %T", input)
}

func (j *BmWKBExecJob) Run(job *progress.Job, input interface{}) (interface{}, error) {
	if db, ok := input.(*sqlx.DB); ok {
		const gridRef = 99

		var opts []wkbcommon.WKBOption

		g := geom.NewPolygon(geom.XY).MustSetCoords([][]geom.Coord{{{1, 2}, {3, 4}, {5, 6}, {1, 2}}})

		wkb, err := wkb.Marshal(g, wkb.NDR, opts...)
		if err != nil {
			return struct{}{}, err
		}

		var tpl string
		switch db.Driver().(type) {
		case *pq.Driver:
			tpl = "INSERT INTO %s (ID, GRIDREF, ogc_geom) VALUES (%v, %v, ST_GeomFromWKB('\\x%v', 4236))"
		default:
			tpl = "INSERT INTO %s (ID, GRIDREF, ogc_geom) VALUES (%v, %v, ST_GeomFromWKB(X'%v', 4236))"
		}

		for name, task := range job.Tasks {
			task.Start()

			insert := fmt.Sprintf(tpl,
				benchTableName,
				name,
				gridRef,
				hex.EncodeToString(wkb),
			)

			if _, err := db.Exec(insert); err != nil {
				panic(err)
			}

			task.End()
			job.IncrBar()
		}
	}

	return struct{}{}, nil
}
