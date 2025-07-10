package mysql

import (
	"database/sql"

	"github.com/aiden2048/pkg/frame/logs"

	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlCfgManager struct {
	dbConf *DbConf
	mSQL   string
	//FirstQuery bool
	LastCheckTime int64
	Close         chan struct{}
}

func RegisterMYSQLHandler(dbconf *DbConf, mSQL string, dataRefreshCallBack func([]map[string]sql.RawBytes)) *MysqlCfgManager {
	m := &MysqlCfgManager{dbconf, mSQL, 0, make(chan struct{})}
	go m.fetchMySQLRoutine(dataRefreshCallBack)
	return m
}

func (m *MysqlCfgManager) GetDbConf() *DbConf {
	return m.dbConf
}

func (m *MysqlCfgManager) Stop() {
	close(m.Close)
}

func (m *MysqlCfgManager) fetchMySQLRoutine(dataRefreshCallBack func([]map[string]sql.RawBytes)) {
	for {
		select {
		case <-m.Close:
			return
		default:
		}
		if m.fetchLastUpdateTime() >= m.LastCheckTime {
			mData, err := m.fetchMYSQLData()
			if err != nil {
				logs.LogError("FetchMYSQLData Error:%s", err.Error())
			} else {
				dataRefreshCallBack(mData)
			}
		}
		m.LastCheckTime = time.Now().Unix()
		time.Sleep(time.Second * 30)
	}
}

func (m *MysqlCfgManager) fetchLastUpdateTime() int64 {
	s := "select TABLE_NAME, unix_timestamp(UPDATE_TIME) from information_schema.tables " +
		"where TABLE_NAME = '" + m.dbConf.Table + "' and TABLE_SCHEMA = '" + m.dbConf.Database + "'"
	data, err := m.dbConf.Query(s)
	if err != nil || len(data) != 1 {
		logs.LogError("FetchLastUpdateTime Error")
		return 0
	}
	lastUpdateTime, _ := strconv.Atoi(string((data)[0]["unix_timestamp(UPDATE_TIME)"]))
	return int64(lastUpdateTime)
}

func (m *MysqlCfgManager) fetchMYSQLData() ([]map[string]sql.RawBytes, error) {
	logs.LogDebug("sql:%s", m.mSQL)
	return m.dbConf.Query(m.mSQL)
}
