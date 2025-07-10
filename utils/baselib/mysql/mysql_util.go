package mysql

import (
	"database/sql"
	"strconv"

	"github.com/aiden2048/pkg/frame/logs"
)

type DbConf struct {
	UserName string
	Password string
	Host     string
	Database string
	Table    string
	Port     int32
	Charset  string
}

func (dbCfg *DbConf) setMySQLConf() string {
	port := strconv.Itoa(int(dbCfg.Port))
	if dbCfg.Charset == "" {
		dbCfg.Charset = "utf8"
	}
	return dbCfg.UserName + ":" + dbCfg.Password + "@tcp(" + dbCfg.Host + ":" + port + ")/" + dbCfg.Database + "?charset=" + dbCfg.Charset
}

func (dbCfg *DbConf) Query(mSQL string) ([]map[string]sql.RawBytes, error) {
	mDbConfString := dbCfg.setMySQLConf()
	db, err := sql.Open("mysql", mDbConfString)
	if err != nil {
		logs.LogError(err.Error())
		return nil, err
	}
	defer db.Close()

	//1、query
	rows, err := db.Query(mSQL)
	if err != nil {
		logs.LogError(err.Error())
		return nil, err
	}
	defer rows.Close()

	//2、prepare data space
	columns, err := rows.Columns()
	if err != nil {
		logs.LogError(err.Error())
		return nil, err
	}

	//3、processing data
	totalRecord := make([]map[string]sql.RawBytes, 0)
	for rows.Next() {
		values := make([]sql.RawBytes, len(columns))
		scanArgus := make([]interface{}, len(values))
		for i := range values {
			scanArgus[i] = &values[i]
		}
		err = rows.Scan(scanArgus...)
		if err != nil {
			logs.LogError("Scan mysql data error")
			break
		}
		rowRecord := make(map[string]sql.RawBytes)
		for i, col := range values {
			newcol := sql.RawBytes(make([]byte, len([]byte(col))))
			copy([]byte(newcol), []byte(col))
			rowRecord[columns[i]] = newcol
		}
		totalRecord = append(totalRecord, rowRecord)
	}
	return totalRecord, nil
}
