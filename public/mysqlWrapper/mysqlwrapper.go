package mysqlWrapper

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"

	"github.com/go-sql-driver/mysql"
)

var g_MysqlDB *sql.DB

var g_MysqlOnlyReadDB []*sql.DB
var g_MysqlOnlyReadDB_Len int

func StartMysql(database string) error {

	return StartMysqlEx(frame.GetMysqlConfig(), database)
}
func StartMysqlEx(cfg *frame.MysqlCfg, database string) error {

	err, db := InitMysql("Real", &cfg.Real, database)
	if err != nil {
		return err
	}
	g_MysqlDB = db
	g_MysqlOnlyReadDB = make([]*sql.DB, 0)
	for n, cfg := range cfg.ReadOnly {
		key := fmt.Sprintf("ReadOnly_%d", n)
		//err = InitOnlyReadsql(host, port, user, pwd, db)
		err, db := InitMysql(key, cfg, database)
		if err != nil {
			continue
		}
		g_MysqlOnlyReadDB = append(g_MysqlOnlyReadDB, db)
	}
	if err != nil {
		return err
	}
	g_MysqlOnlyReadDB_Len = len(g_MysqlOnlyReadDB)
	return err
}
func InitMysql(key string, cfg *frame.MysqlSvrCfg, database string) (error, *sql.DB) {
	mysqlCfg := mysql.Config{Net: "tcp", Addr: "127.0.0.1:3306", DBName: "dbname", Collation: "utf8_general_ci", Loc: time.UTC}
	if cfg.Port == 0 {
		cfg.Port = 3306
	}
	mysqlCfg.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mysqlCfg.User = cfg.User
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.DBName = database
	mysqlCfg.Timeout = 5 * time.Second
	if len(cfg.Param) > 0 {
		mysqlCfg.Params = cfg.Param
	}
	mysqlCfg.AllowNativePasswords = true
	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		logs.Errorf("ConnMysql %s: %+v, db:%s failed:%s", key, cfg, database, err.Error())
		return err, nil
	}
	maxConnCount := 2000 //runtime.GOMAXPROCS(0) * 16
	db.SetMaxOpenConns(maxConnCount)
	db.SetMaxIdleConns(32)

	db.SetConnMaxLifetime(time.Hour * 4)

	db.Ping()

	logs.Infof("Conn to Mysql")
	return nil, db
}

func GetRealDB() *sql.DB {
	return g_MysqlDB
}

func GetOnlyReadDB() *sql.DB {
	if g_MysqlOnlyReadDB_Len == 0 {
		return nil
	}
	n := rand.Int31n(int32(g_MysqlOnlyReadDB_Len))
	return g_MysqlOnlyReadDB[n]
}
