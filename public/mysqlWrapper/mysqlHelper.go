package mysqlWrapper

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/stat"
)

func getTableErr(tbl string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("Tbl:%s, %s", tbl, err.Error())
}
func InsertOrUpdateDB(table string, data map[string]interface{}, where string) (int64, error) {
	db := g_MysqlDB
	if db == nil {
		return 0, errors.New("No Mysql DB")
	}
	if where == "" {
		return InsertDB(table, data)
	}
	id, err := UpdateDB(table, data, where) //更新不成功 那就插入
	if err != nil || id <= 0 {
		id, err = InsertDB(table, data)
		if err != nil || id <= 0 { //有问题再来一次更新
			id, err = UpdateDB(table, data, where)
		}
	}
	return id, err

	/*Id := int64(0)
	db := g_MysqlDB
	if db == nil {
		return Id, errors.New("No Mysql DB")
	}
	count := len(data)
	if count <= 0 {
		return Id, errors.New("No Update item")
	}
	arrKeys := make([]string, 0, count)
	arrValues := make([]interface{}, 0, count)
	for k, v := range data {
		k = k + "=?"
		arrKeys = append(arrKeys, k)
		arrValues = append(arrValues, v)
	}
	sqlStr := fmt.Sprintf("INSERT INTO %s set %s", table, strings.Join(arrKeys, ","))
	sqlStr = sqlStr + fmt.Sprintf("ON DUPLICATE KEY UPDATE %s", strings.Join(arrKeys, ","))

	start := time.Now()
	sqlStr = sqlStr + " WHERE " + where
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		logs.LogError("db.Prepare %s failed: %+v", sqlStr, err)
		return Id, err
	}
	defer stmt.Close()
	arrValues1 := append(arrValues, arrValues...)
	res, err := stmt.Exec(arrValues1...)

	logs.LogDebug("exec sql: %s,arrValues: %+v , result:%+v, err:+%v", sqlStr, arrValues, res, err)
	if err != nil {
		logs.LogError("InsertOrUpdateDB %s failed: %s", sqlStr, err.Error())
		stat.ReportStat("rp:mysql.InsertOrUpdateDB."+table, -1, time.Now().Sub(start))
		return Id, err
	}
	stat.ReportStat("rp:mysql.InsertOrUpdateDB."+table, 0, time.Now().Sub(start))
	Id, err = res.RowsAffected()
	if err != nil {
		return Id, err
	}
	return Id, nil*/
}

// 插入数据库通用函数 不可以使用bool 可以使用 int int32 int64 string  where => "uid=1 and cid=2"
func UpdateDB(table string, data map[string]interface{}, where string) (int64, error) {
	Id := int64(0)
	db := g_MysqlDB
	if db == nil {
		return Id, errors.New("No Mysql DB")
	}
	count := len(data)
	if count <= 0 {
		return Id, errors.New("No Update item")
	}
	arrKeys := make([]string, 0, count)
	arrValues := make([]interface{}, 0, count)
	for k, v := range data {
		k = k + "=?"
		arrKeys = append(arrKeys, k)
		tmp, ok := v.(string)
		if ok {
			//tmp = strings.Replace(tmp, "?", "？", -1)
			v = tmp
		}
		arrValues = append(arrValues, v)
	}
	sqlStr := fmt.Sprintf("update %s set %s", table, strings.Join(arrKeys, ","))

	if where == "" {
		return Id, errors.New("not set where")
	}
	start := time.Now()

	sqlStr = sqlStr + " WHERE " + where
	stmt, err := db.Prepare(sqlStr)
	err = getTableErr(table, err)
	if err != nil {

		logs.LogError("db.Prepare %s failed: %+v", sqlStr, err)
		stat.ReportStat("rp:mysql.UpdateDB."+table, -1, time.Now().Sub(start))
		return Id, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(arrValues...)
	err = getTableErr(table, err)
	logs.LogDebug("exec sql: %s,arrValues: %+v , result:%+v, err:+%v", sqlStr, arrValues, res, err)
	if err != nil {

		logs.LogError("UpdateDB %s failed: %s", sqlStr, err.Error())
		stat.ReportStat("rp:mysql.UpdateDB."+table, -2, time.Now().Sub(start))
		return Id, err
	}
	stat.ReportStat("rp:mysql.UpdateDB."+table, 0, time.Now().Sub(start))
	Id, err = res.RowsAffected()
	err = getTableErr(table, err)
	if err != nil {
		logs.LogError("getTableErr sql: %s failed: %s table:%s", sqlStr, err.Error())
		return Id, err
	}
	return Id, nil
}

// 插入数据库通用函数 不可以使用bool 可以使用 int int32 int64 string
func InsertDB(table string, data map[string]interface{}) (int64, error) {
	Id := int64(0)
	db := GetRealDB()
	if db == nil {
		return Id, errors.New("No Mysql DB")
	}
	count := len(data)
	if count <= 0 {
		return Id, errors.New("No Insert DB")
	}

	arrKeys := make([]string, 0, count)
	arrVValues := make([]string, 0, count)
	arrValues := make([]interface{}, 0, count)
	for k, v := range data {
		arrKeys = append(arrKeys, k)
		arrVValues = append(arrVValues, "?")
		tmp, ok := v.(string)
		if ok {
			// tmp = strings.Replace(tmp, "?", "'?", -1)
			v = tmp
		}
		arrValues = append(arrValues, v)
	}
	sqlStr := fmt.Sprintf("Insert into %s (%s) values(%s)", table, strings.Join(arrKeys, ","), strings.Join(arrVValues, ","))
	start := time.Now()
	stmt, err := db.Prepare(sqlStr)
	err = getTableErr(table, err)
	if err != nil {
		logs.LogError("db.Prepare  %s failed: %+v", sqlStr, err)
		stat.ReportStat("rp:mysql.InsertDB."+table, -1, time.Now().Sub(start))
		return Id, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(arrValues...)
	err = getTableErr(table, err)
	logs.LogDebug("exec sql: %s,arrValues = %+v, result:%+v, err:+%v", sqlStr, arrValues, res, err)
	if err != nil {
		stat.ReportStat("rp:mysql.InsertDB."+table, -2, time.Now().Sub(start))
		return Id, err
	}
	stat.ReportStat("rp:mysql.InsertDB."+table, 0, time.Now().Sub(start))
	Id, err = res.LastInsertId()
	err = getTableErr(table, err)
	if err != nil {

		return Id, err
	}
	return Id, nil
}

// 对参数做防注入过滤
func SafeExec(format string, args ...interface{}) (sql.Result, error) {
	ret, err, _ := SafeExecSql(format, args...)
	return ret, err
}
func SafeExecSql(format string, args ...interface{}) (sql.Result, error, string) {
	db := GetRealDB()
	if db == nil {
		return nil, errors.New("No Mysql DB"), ""
	}
	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case string:
			{
				args[i] = strings.Replace(args[i].(string), "'", "''", -1)
				args[i] = strings.Replace(args[i].(string), "\\", "\\\\", -1)
			}
		default:

		}
	}
	sql := fmt.Sprintf(format, args...)
	logs.LogDebug(sql)
	start := time.Now()
	ret, err := db.Exec(sql)
	res := 0
	if err != nil {
		logs.LogError("SafeExec %s failed: %s", sql, err.Error())
		res = -1
	}
	tmpStr := sql
	ind := strings.Index(tmpStr, "where")
	if ind >= 0 {
		tmpStr = tmpStr[:ind]
	}
	stat.ReportStat("rp:mysql.SafeExec."+tmpStr, res, time.Now().Sub(start))
	return ret, err, sql
}

// 对参数做防注入过滤
func RealSafeQuery(format string, args ...interface{}) (*sql.Rows, error) {
	ret, err, _ := RealSafeQuerySql(format, args...)
	return ret, err
}
func RealSafeQuerySql(format string, args ...interface{}) (*sql.Rows, error, string) {
	db := GetRealDB()
	if db == nil {
		return nil, errors.New("No Mysql DB"), ""
	}
	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case string:
			{
				args[i] = strings.Replace(args[i].(string), "'", "''", -1)
				args[i] = strings.Replace(args[i].(string), "\\", "\\\\", -1)
			}
		default:

		}
	}
	sql := fmt.Sprintf(format, args...)
	//logs.LogDebug(sql)
	start := time.Now()
	ret, err := db.Query(sql)

	//get tableName
	tmpStr := strings.ToLower(sql)
	ind := strings.Index(tmpStr, "from ") + 5
	if ind >= 0 {
		tmpStr = tmpStr[ind:]
	}
	tmpStr = strings.TrimLeft(tmpStr, " ")
	ind = strings.Index(tmpStr, " ")
	if ind > 0 {
		tmpStr = tmpStr[:ind]
	}
	res := 0
	if err != nil {
		logs.LogError("SafeQuery %s failed: %s", sql, err.Error())
		res = -1
	}
	end := time.Now()
	n := end.Sub(start)
	stat.ReportStat("rp:mysql.SafeQuery."+tmpStr, res, n)
	if n > 5*time.Second { //5秒以上 告警
		logs.LogError("查询时间 超过 5s sql:%+v,time = %d", sql, n/time.Millisecond)
	}
	return ret, err, sql
}

// 对参数做防注入过滤
func SafeQuery(format string, args ...interface{}) (*sql.Rows, error) {
	ret, err, _ := SafeQuerySql(format, args...)
	return ret, err
}
func SafeQuerySql(format string, args ...interface{}) (*sql.Rows, error, string) {
	db := GetOnlyReadDB()
	if db == nil { //没有只读 那就获取 读写的
		db = GetRealDB()
		if db == nil {
			return nil, errors.New("No Mysql DB"), ""
		}
	}
	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case string:
			{
				args[i] = strings.Replace(args[i].(string), "'", "''", -1)
				args[i] = strings.Replace(args[i].(string), "\\", "\\\\", -1)
			}
		default:

		}
	}
	sql := fmt.Sprintf(format, args...)
	logs.LogDebug(sql)
	start := time.Now()
	ret, err := db.Query(sql)

	//get tableName
	tmpStr := strings.ToLower(sql)
	ind := strings.Index(tmpStr, "from ") + 5
	if ind >= 0 {
		tmpStr = tmpStr[ind:]
	}
	tmpStr = strings.TrimLeft(tmpStr, " ")
	ind = strings.Index(tmpStr, " ")
	if ind > 0 {
		tmpStr = tmpStr[:ind]
	}
	res := 0
	if err != nil {
		logs.LogError("SafeQuery %s failed: %s", sql, err.Error())
		res = -1
	}
	end := time.Now()
	n := end.Sub(start)
	stat.ReportStat("rp:mysql.SafeQuery."+tmpStr, res, n)
	if n > 6*time.Second { //5秒以上 告警
		logs.LogError("查询时间 超过 5s sql:%+v,time=%d", sql, n/time.Millisecond)
	}
	return ret, err, sql
}

// 直接查询 对于 有%% 才可以用 有闲在用
func SafeQueryByString(sql string) (*sql.Rows, error) {
	db := GetOnlyReadDB()
	if db == nil { //没有只读 那就获取 读写的
		db = GetRealDB()
		if db == nil {
			return nil, errors.New("No Mysql DB")
		}
	}
	logs.LogDebug(sql)
	start := time.Now()
	ret, err := db.Query(sql)
	//get tableName
	tmpStr := strings.ToLower(sql)
	ind := strings.Index(tmpStr, "from ") + 5
	if ind >= 0 {
		tmpStr = tmpStr[ind:]
	}
	tmpStr = strings.TrimLeft(tmpStr, " ")
	ind = strings.Index(tmpStr, " ")
	if ind > 0 {
		tmpStr = tmpStr[:ind]
	}
	res := 0
	if err != nil {
		logs.LogError("SafeQueryByString %s failed: %s", sql, err.Error())
		res = -1
	}
	end := time.Now()
	n := end.Sub(start)
	stat.ReportStat("rp:mysql.SafeQuery."+tmpStr, res, n)
	if n > 5*time.Second { //5秒以上 告警
		logs.LogError("查询时间 超过 5s sql:%+v,time=%d", sql, n/time.Millisecond)
	}
	return ret, err
}

func DelTableDB(table string, format string, args ...interface{}) (int64, error) {
	db := GetRealDB()
	if db == nil {
		return 0, errors.New("No Mysql DB")
	}
	row := int64(0)
	DelSql := fmt.Sprintf("DELETE FROM %s WHERE ", table)
	if len(args) > 0 {
		for i := 0; i < len(args); i++ {
			switch args[i].(type) {
			case string:
				{
					args[i] = strings.Replace(args[i].(string), "'", "''", -1)
					args[i] = strings.Replace(args[i].(string), "\\", "\\\\", -1)
				}
			default:
			}
		}

		DelSql = fmt.Sprintf(DelSql+" "+format, args...)
	} else {
		DelSql = DelSql + " " + format
	}
	start := time.Now()
	res, err := db.Exec(DelSql)
	err = getTableErr(table, err)
	logs.LogDebug("exec sql: %s, result:%+v, err:+%v", DelSql, res, err)
	if err != nil {
		return row, err
	}
	row, err = res.RowsAffected()
	err = getTableErr(table, err)
	if err != nil {
		logs.LogError("DelTableDB %s failed: %s", DelSql, err.Error())
		return row, err
	}
	end := time.Now()
	n := end.Sub(start)
	stat.ReportStat("rp:mysql.SafeQuery."+table, 0, n)
	return row, nil
}
