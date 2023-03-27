package benchmark

import (
	"fmt"
	"runtime"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rockwell-uk/go-utils/timeutils"
)

const benchTableName string = "benchmark"

type BenchmarkSuite struct {
	WarmUp     func(*sqlx.DB) error
	PrintStats bool
}

type BenchmarkResult struct {
	Duration time.Duration
	Queries  int
	Allocs   uint64
	Bytes    uint64
	Err      error
}

type Benchmark struct {
	Name string
	N    int
	Bm   func(*sqlx.DB, int) error
}

func (r BenchmarkResult) String() string {
	return fmt.Sprintf("\t\t\t\t"+"Duration: %v"+"\n"+
		"\t\t\t\t"+"Queries: %v"+"\n"+
		"\t\t\t\t"+"Allocs: %v"+"\n"+
		"\t\t\t\t"+"Bytes: %v"+"\n"+
		"\t\t\t\t"+"Err: %v",
		timeutils.FormatDuration(r.Duration, 2),
		r.Queries,
		r.Allocs,
		r.Bytes,
		r.Err,
	)
}

func (res *BenchmarkResult) QueriesPerSecond() float64 {
	return float64(res.Queries) / res.Duration.Seconds()
}

func (res *BenchmarkResult) AllocsPerQuery() int {
	return int(res.Allocs) / res.Queries
}

func (res *BenchmarkResult) BytesPerQuery() int {
	return int(res.Bytes) / res.Queries
}

func (b *Benchmark) Run(db *sqlx.DB) BenchmarkResult {
	runtime.GC()

	var memStats runtime.MemStats
	var startMallocs uint64 = memStats.Mallocs
	var startTotalAlloc uint64 = memStats.TotalAlloc
	var startTime time.Time = time.Now()

	runtime.ReadMemStats(&memStats)

	err := b.Bm(db, b.N)

	endTime := time.Now()
	runtime.ReadMemStats(&memStats)

	return BenchmarkResult{
		Err:      err,
		Queries:  b.N,
		Duration: endTime.Sub(startTime),
		Allocs:   memStats.Mallocs - startMallocs,
		Bytes:    memStats.TotalAlloc - startTotalAlloc,
	}
}

func Warmup(db *sqlx.DB) error {
	db.SetMaxIdleConns(16)

	for i := 0; i < 10000; i++ {
		rows, err := db.Query("SELECT 5 * 3")
		if err != nil {
			return err
		}

		if err = rows.Close(); err != nil {
			return err
		}
	}

	return nil
}
