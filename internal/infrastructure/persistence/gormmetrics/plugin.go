package gormmetrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
)

var (
	gormDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "medilog",
		Subsystem: "gorm",
		Name:      "query_duration_seconds",
		Help:      "GORM query duration in seconds bucketed by latency",
		Buckets:   prometheus.DefBuckets,
	}, []string{"table", "operation"})
)

// Plugin records GORM query durations as Prometheus histograms.
// Remove the db.Use(gormmetrics.Plugin{}) call to switch back to OTel.
type Plugin struct{}

func (Plugin) Name() string {
	return "gorm:metrics"
}

func (Plugin) Initialize(db *gorm.DB) error {
	for _, op := range []struct {
		name   string
		before func(string, func(*gorm.DB)) error
		after  func(string, func(*gorm.DB)) error
	}{
		{"query",
			db.Callback().Query().Before("gorm:query").Register,
			db.Callback().Query().After("gorm:query").Register},
		{"create",
			db.Callback().Create().Before("gorm:create").Register,
			db.Callback().Create().After("gorm:create").Register},
		{"update",
			db.Callback().Update().Before("gorm:update").Register,
			db.Callback().Update().After("gorm:update").Register},
		{"delete",
			db.Callback().Delete().Before("gorm:delete").Register,
			db.Callback().Delete().After("gorm:delete").Register},
		{"raw",
			db.Callback().Raw().Before("gorm:raw").Register,
			db.Callback().Raw().After("gorm:raw").Register},
		{"row_query",
			db.Callback().Row().Before("gorm:row_query").Register,
			db.Callback().Row().After("gorm:row_query").Register},
	} {
		name := op.name
		_ = op.before("metrics:before_"+name, func(db *gorm.DB) {
			db.InstanceSet("gorm:metrics:start", time.Now())
		})
		_ = op.after("metrics:after_"+name, func(db *gorm.DB) {
			v, ok := db.InstanceGet("gorm:metrics:start")
			if !ok {
				return
			}
			start, ok := v.(time.Time)
			if !ok {
				return
			}

			table := db.Statement.Table
			if table == "" && db.Statement.Schema != nil {
				table = db.Statement.Schema.Table
			}
			if table == "" {
				table = "unknown"
			}

			gormDuration.WithLabelValues(table, name).Observe(time.Since(start).Seconds())
		})
	}

	return nil
}
