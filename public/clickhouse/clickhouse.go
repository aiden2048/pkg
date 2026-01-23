package clickhouse

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/aiden2048/pkg/frame"
)

var dbcs *sync.Map = &sync.Map{} // map[int32]*sql.DB

var ckUrl string // clickhouse uri

func GetCkUri(platId int32) string {
	// Assuming a similar config retrieval mechanism exists or will exist
	// For now, returning global or default if not set
	return ckUrl
}

func StartClickHouse(dsn string) error {
	return StartPlatClickHouse(frame.GetPlatformId(), dsn)
}

func StartPlatClickHouse(platId int32, dsn string) error {
	_, ok := dbcs.Load(platId)
	if !ok {
		db, err := startCk(dsn)
		if err != nil {
			return err
		}
		dbcs.Store(platId, db)
	}
	return nil
}

func startCk(dsn string) (*sql.DB, error) {
	// Parsing DSN or configuring directly options could be done here.
	// Using sql.Open with clickhouse driver
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("clickhouse connect error: %v", err)
	}

	if err := db.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			return nil, fmt.Errorf("clickhouse ping error: [%d] %s \n%s", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, fmt.Errorf("clickhouse ping error: %v", err)
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func GetCkDb(platId int32) *sql.DB {
	v, ok := dbcs.Load(platId)
	if ok {
		return v.(*sql.DB)
	}
	// Fallback or retry logic could go here
	return nil
}
