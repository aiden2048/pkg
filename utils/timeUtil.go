package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	TimeFormat_Full           = "2006-01-02 15:04:05.999"
	TimeFormat_Sec            = "2006-01-02 15:04:05"
	TimeFormat_MonthSec       = "0102150405"
	TimeFormat_Sec_12_PM      = "2006-01-02 03:04:05PM"
	TimeFormatSecZone         = "2006-01-02T15:04:05+07:00"
	TimeFormatSecE8Zone       = "2006-01-02T15:04:05+08:00"
	TimeFormat_TSec           = "2006-01-02T15:04:05"
	TimeFormat_TZ             = "2006-01-02T15:04:05Z"
	TimeFormat_TZSEC          = "2006-01-02T15:04:05.999Z"
	TimeFormat_TFull          = "2006-01-02T15:04:05.999"
	TimeFormat_YMD            = "20060102"
	TimeFormat_Sec_New        = "2006/01/02 15:04:05"
	TimeFormat_Sec_NoZero     = "2006-1-2 15:4:5"
	TimeFormat_Minute         = "2006-01-02 15:04"
	TimeFormat_Hour           = "2006-01-02 15"
	TimeFormat_Date           = "2006-01-02"
	TimeFormat_IDate          = "20060102"
	TimeFormat_IDate2         = "2006/01/02"
	TimeFormat_Month          = "200601"
	TimeFormat_LianTime       = "20060102150405"
	TimeFormat_LianTimeMs     = "20060102150405000"
	TimeFormat_HourTime       = "2006010215"
	TimeFormat_DateTime       = "01-02"
	TimeFormat_HourMinute     = "15:04"
	TimeFormat_HourMinuteTime = "15:04:05"
	TimeZeroPointsDff         = 72000 //与北京时间相差
	TimeRFC3339               = "2006-01-02T15:04:05-07:00"
	TimeFormatSecZone1        = "15:04 02.01.2006"
	TimeFormat_SecUTC         = "02-01-2006"
	TimeUTC_4                 = "2006-01-02T15:04:05-04:00"
	TimeFormat_Min            = "2006-01-02 15:04:00"
	TimeFormat_TimeHour       = "2006-01-02 15:00:00"
)

// 获取小时
func GetDateHour(t int64) int64 {
	return StrToInt64(time.Unix(t, 0).Format(TimeFormat_HourTime))
}

func GetUTCZERODateHour(t int64) int64 {
	return StrToInt64(time.Unix(t, 0).In(time.UTC).Format(TimeFormat_HourTime))
}

// idate 转时间戳
func ParseIdateToUnix(idate int64) int64 {
	local, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(TimeFormat_LianTime, fmt.Sprintf("%d000000", idate), local)
	return theTime.Unix()
}

/*
*
CST时间转到美东时间 utc-4
*/
func TimeToZoneUTC4(str string) (string, error) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return "", err
	}
	local, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(TimeFormat_Sec, str, local)
	return theTime.In(loc).Format(TimeFormat_Sec), nil
}

func TimeToZoneUTC4TZ(str string) (string, error) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return "", err
	}
	local, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(TimeFormat_Sec, str, local)
	return theTime.In(loc).Format(TimeFormat_TZ), nil
}

func TimeToZoneUTC4YMD(t int64) (string, error) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return "", err
	}
	return time.Unix(t, 0).In(loc).Format(TimeFormat_YMD), nil
}

func TimeToZoneUTCTZ(t int64) (string, error) {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return "", err
	}

	return time.Unix(t, 0).In(loc).Format(TimeFormat_TZ), nil
}
func TimeToZoneLocalTZ(t int64) (string, error) {
	local, _ := time.LoadLocation("Local")
	return time.Unix(t, 0).In(local).Format(TimeFormat_TZ), nil
}
func TimeToZoneUTCSEC(t int64) (string, error) {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return "", err
	}

	return time.Unix(t, 0).In(loc).Format(TimeFormat_Sec), nil
}

func TimeParseToGMT(t int64) (string, error) {
	loc, err := time.LoadLocation("GMT")
	if err != nil {
		return "", err
	}
	return time.Unix(t, 0).In(loc).Format(TimeFormat_Min), nil
}

func TimeUtcUnix(str string) int64 {
	strTime, err := time.Parse(TimeFormatSecZone1, str)
	if err != nil {
		return 0
	}
	return strTime.UTC().Unix()
}

func TimeUtcUnix_TSec(str string) int64 {
	strTime, err := time.Parse(TimeFormat_TSec, str)
	if err != nil {
		return 0
	}
	return strTime.UTC().Unix()
}

func TimeToString(str string) int64 {
	strTime, err := time.Parse(TimeFormat_Sec, str)
	if err != nil {
		return 0
	}
	return strTime.UTC().Unix()
}

func TimeToString_add3h(str string) int64 {
	strTime, err := time.Parse(TimeFormat_Sec, str)
	if err != nil {
		return 0
	}
	return strTime.Add(190 * time.Minute).UTC().Unix()
}

func TimeToString_before5days(str string) int64 {
	strTime, err := time.Parse(TimeFormat_Sec, str)
	if err != nil {
		return 0
	}
	return strTime.Add(-115 * time.Hour).UTC().Unix()
}

func TimeToStringFormat(str string) int64 {
	local, _ := time.LoadLocation("Local")
	strTime, err := time.ParseInLocation(TimeFormat_SecUTC, str, local)
	if err != nil {
		return 0
	}
	return strTime.Unix()
}

func TimeToStringToItime(str string) int64 {
	local, _ := time.LoadLocation("Local")
	strTime, err := time.ParseInLocation(TimeFormat_Date, str, local)
	if err != nil {
		return 0
	}
	return strTime.Unix()
}

// utc-4时间转时间戳
func TimeUTC_4StringToItime(str string) int64 {
	local, _ := time.LoadLocation("America/New_York")
	strTime, err := time.ParseInLocation(TimeFormat_Sec, str, local)
	if err != nil {
		return 0
	}
	return strTime.Unix()
}

func TimeUTC0StringToItime(str string) int64 {
	local, _ := time.LoadLocation("UTC")
	strTime, err := time.ParseInLocation(TimeFormat_Sec, str, local)
	if err != nil {
		return 0
	}
	return strTime.Unix()
}

func TimeUTCTZStringToItime(str string) int64 {
	local, _ := time.LoadLocation("UTC")
	strTime, err := time.ParseInLocation(TimeFormat_TZ, str, local)
	if err != nil {
		return 0
	}
	return strTime.Unix()
}

func TimeUTC0FullStringToItime(str string) int64 {
	local, _ := time.LoadLocation("UTC")
	strTime, err := time.ParseInLocation(TimeFormat_Full, str, local)
	if err != nil {
		return 0
	}
	return strTime.Unix()
}
func TimeUTC0TFullStringToItime(str string) int64 {
	local, _ := time.LoadLocation("America/New_York")
	strTime, err := time.ParseInLocation(TimeFormat_TFull, str, local)
	if err != nil {
		return 0
	}
	return strTime.Unix()
}
func TimeParseUtc_4(str string) int64 {
	_, offset := time.Now().Zone()
	t, _ := time.Parse(TimeFormat_TFull, str)
	fmt.Println(offset)
	return t.Unix() - int64(offset)
}
func TimeUTC0TZStringToItime(str string) int64 {
	local, _ := time.LoadLocation("UTC")
	strTime, err := time.ParseInLocation(TimeFormat_TZSEC, str, local)
	if err != nil {
		return 0
	}
	return strTime.Unix()
}
func GetMonthInt(t int64) int {
	return int(time.Unix(t, 0).Month())
}

/*
*
EST 美东时间
*/
func TimeParseToEST(str string) string {
	local, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(TimeFormat_Sec, str, local)
	location, _ := time.LoadLocation("America/New_York")
	return theTime.In(location).Format(TimeFormat_TSec) + "-04:00"

	//local, _ := time.LoadLocation("Local")
	//theTime, _ := time.ParseInLocation(TimeFormat_Sec, str, local)
	//location, _ := time.LoadLocation("EST")
	//return theTime.In(location).Format(TimeFormat_TSec) + "-04:00"
}

/*
*
Asia 北京时间
*/
func TimeParseToCN(str string) string {
	local, _ := time.LoadLocation("Asia/Shanghai")
	theTime, _ := time.ParseInLocation(TimeFormat_Sec, str, local)
	location, _ := time.LoadLocation("Local")
	str = theTime.In(location).Format(TimeFormat_TSec)
	return strconv.FormatInt(TimeUtcUnix_TSec(str), 10)
}

/*
*
Asia/Yekaterinburg
*/
func TimeParseToYK(str string) string {
	local, _ := time.LoadLocation("Asia/Yekaterinburg") //
	theTime, _ := time.ParseInLocation(TimeFormat_Sec, str, local)
	location, _ := time.LoadLocation("Local")
	str = theTime.In(location).Format(TimeFormat_TSec)
	return strconv.FormatInt(TimeUtcUnix_TSec(str), 10)
}

/*
*
CST时间转到美东时间 utc-4
*/
func TimeToZoneUTC4RFC3339(str string) (string, error) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return "", err
	}
	local, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(TimeFormat_Sec, str, local)
	return theTime.In(loc).Format(TimeRFC3339), nil
}

// 返回 "15:50"
func GetHourMin() string {
	return time.Now().Format(TimeFormat_HourMinute)
}
func GetTimeHour(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_TimeHour)
}
func TimeUTC0TZStringToIntTime(str string) int64 {
	local, _ := time.LoadLocation("UTC")
	strTime, err := time.ParseInLocation(TimeFormat_TimeHour, str, local)
	if err != nil {
		return 0
	}
	return strTime.Unix()
}
func GetTimeFull() string {
	return time.Now().Format(TimeFormat_TFull)
}

func GetYMD() string {
	return time.Now().Format(TimeFormat_IDate)
}

func GetTimeMs() string {
	return time.Now().Format(TimeFormat_LianTimeMs)
}

// 月日
func GetMd(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_DateTime)
}

// 年月日
func GetYmd(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_IDate)
}

func GetDay(t int64) string {
	str := time.Unix(t, 0).Format(TimeFormat_Date)
	s := strings.Split(str, "-")
	if len(s) < 3 {
		return ""
	}
	return s[2]
}

// 获取现在的时间
func GetNowTimeString() string {
	return time.Now().Format(TimeFormat_Sec)
}

func GetNowTimeStringTZ() string {
	return time.Now().Format(TimeFormat_TZ)
}

func GetUTCTimeString(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_SecUTC)
}

func GetUTCForMatTimeString(t int64) string {
	return time.Unix(t, 0).Format(TimeFormatSecZone1)
}

// 时间搓转时间字符串
func TimeIntToString(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_Sec)
}

func GetUtc8TimeString(now time.Time) string {
	var cstSh, _ = time.LoadLocation("Asia/Shanghai")
	return now.In(cstSh).Format(TimeFormat_Sec)
}

// 时间搓转时间字符串
func TimeIntToStringMonth(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_MonthSec)
}

// 时间搓转时间字符串
func TimeIntTo12PMString(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_Sec_12_PM)
}

// 时间搓转本地时间字符串
func TimeIntToLocalString(t int64) string {
	return time.Unix(t, 0).Local().Format(TimeFormat_Sec)
}

func TimeIntToStringZone(t int64) string {
	return time.Unix(t, 0).Format(TimeFormatSecZone)
}

func TimeIntToStringE8Zone(t int64) string {
	return time.Unix(t, 0).Format(TimeFormatSecE8Zone)
}

// 时间搓转时间字符串
func TimeIntToLianString(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_LianTime)
}

// 时间搓转时间字符串
func ToStringTime(t int64) string {
	if t == 0 {
		return ""
	}
	return time.Unix(t, 0).Format(TimeFormat_Sec)
}

// 时间搓转时间字符串
func TimeStrToTime(str string) int64 {
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(TimeFormat_Sec, str, loc)
	return theTime.Unix()
}
func TimeStrToTimeZone(str string) int64 {
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(TimeFormatSecZone, str, loc)
	return theTime.Unix()
}
func TimeStrToTimeE8Zone(str string) int64 {
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(TimeFormatSecE8Zone, str, loc)
	return theTime.Unix()
}
func IntToTimeZoneMinute(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_Minute)
}

func TimeStrYearToTime(t time.Time, str string) int64 {
	str = fmt.Sprintf("%d-%s", t.Year(), str)
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(TimeFormat_Sec_NoZero, str, loc)
	return theTime.Unix()
}

// 获取 月
func GetIMonth(t int64) int64 {
	return StrToInt64(time.Unix(t, 0).Format(TimeFormat_Month))
}

// 获取传入时间的周
func GetIWeek(t int64) int64 {
	year, week := time.Unix(t, 0).ISOWeek()
	return StrToInt64(fmt.Sprintf("%d%d", year, week))
}

func ParseMonthToTime(v int64) time.Time {
	t, _ := time.Parse(TimeFormat_Month, IntToStr(v))
	return t
}

// 获取传入的时间所在月份的第一天，即某月第一天的0点。如传入time.Now(), 返回当前月份的第一天0点时间。
func GetFirstDateOfMonth(d time.Time) time.Time {
	d = d.AddDate(0, 0, -d.Day()+1)
	return GetZeroTime(d)
}

// 获取某一天的0点时间
func GetZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

// 获取传入的时间所在月份的最后一天，即某月最后一天的0点。如传入time.Now(), 返回当前月份的最后一天0点时间。
func GetLastDateOfMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 1, -1)
}

// 获取上个月的零点
func GetPrevMonthStartTime() time.Time {
	d := time.Now()
	d = d.AddDate(0, -1, -d.Day()+1)
	return GetZeroTime(d)
}

// 获取上个月最后一天的零点，注意不是23:59:59
func GetPrevMonthEndTime() time.Time {
	return GetPrevMonthStartTime().AddDate(0, 1, -1)
}

// 获取idate
func GetIDate(t int64) int64 {
	return StrToInt64(time.Unix(t, 0).Format(TimeFormat_IDate))
}

// 获取idate2
func GetIDate2(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_IDate2)
}

/*
*
获取int64时间戳 天
*/
func GetTimeDay(t int64) int {
	return time.Unix(t, 0).Day()
}

// 获取idate
func GetStrIDate(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_IDate)
}

func GetStrDate(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_Date)
}
func GetStrHourMinuteTime(t int64) string {
	return time.Unix(t, 0).Format(TimeFormat_HourMinuteTime)
}

// str idata + n
func StrIDateInc(idate string, n int64) string {

	l := len(idate) //不检查 idate
	if l < 6 {
		return "20060102"
	}
	newDate := idate[0:l-4] + "-" + idate[l-4:l-2] + "-" + idate[l-2:l] + " 00:00:00"
	newIdate := TimeStrToTime(newDate) + n*(24*3600)
	return GetStrIDate(newIdate)
}

// int idata + n
func IDateInc(idate int64, n int64) int64 {
	return StrToInt64(StrIDateInc(IntToStr(idate), n))
}

// 获取idate
func GetNowStrIDate(nows ...time.Time) string {
	var now time.Time
	if len(nows) > 0 {
		now = nows[0]
	} else {
		now = time.Now()
	}
	return now.Format(TimeFormat_IDate)
}

// 获取现在到 今天晚点0点的时间差
func GetTomorryZeroTimeDiff(nows ...time.Time) int64 {
	var now time.Time
	if len(nows) > 0 {
		now = nows[0]
	} else {
		now = time.Now()
	}
	timeStr := now.Local().Format(TimeFormat_Date)
	t, _ := time.ParseInLocation(TimeFormat_Sec, timeStr+" 23:59:59", time.Local)
	return (t.Unix() + 1 - now.Unix())
}

// 明天0点
func GetTomorryZeroTime(nows ...time.Time) int64 {
	var now time.Time
	if len(nows) > 0 {
		now = nows[0]
	} else {
		now = time.Now()
	}
	timeStr := now.Local().Format(TimeFormat_Date)
	t, _ := time.ParseInLocation(TimeFormat_Sec, timeStr+" 23:59:59", time.Local)
	return t.Unix() + 1
}

// 今天0点
func GetDayZeroTime(nows ...time.Time) int64 {
	var now time.Time
	if len(nows) > 0 {
		now = nows[0]
	} else {
		now = time.Now()
	}
	timeStr := now.Local().Format(TimeFormat_Date)
	t, _ := time.ParseInLocation(TimeFormat_Sec, timeStr+" 00:00:00", time.Local)
	return t.Unix()
}

func GetDayZeroTimeWithTimezone(timezone *time.Location, nows ...time.Time) int64 {
	var now time.Time
	if len(nows) > 0 {
		now = nows[0]
	} else {
		now = time.Now()
	}
	timeStr := now.In(timezone).Format(TimeFormat_Date)
	t, _ := time.ParseInLocation(TimeFormat_Sec, timeStr+" 00:00:00", timezone)
	return t.Unix()
}

// 根据2006-01-01格式的日期，获取相应的0点
func GetDayZeroTimeByTimeStr(timeStr string) int64 {
	t, _ := time.ParseInLocation(TimeFormat_Sec, timeStr+" 00:00:00", time.Local)
	return t.Unix()
}

// 根据20060102格式的日期，获取相应的0点
func GetDayZeroTimeByIDate(idate int64) int64 {
	t, _ := time.ParseInLocation(TimeFormat_LianTime, fmt.Sprintf("%d000000", idate), time.Local)
	return t.Unix()
}

// 今天0点
func GetDayZero(t int64) int64 {
	timeStr := time.Unix(t, 0).Local().Format(TimeFormat_Date)
	tt, _ := time.ParseInLocation(TimeFormat_Sec, timeStr+" 00:00:00", time.Local)
	return tt.Unix()
}

// 获取45天前0点
func GetDataStartTime() int64 {
	day_start := GetDayZeroTime()
	return day_start - (45 * 24 * 3600)
}

// 获取45天前0点 str
func GetDataStartStrTime() string {
	return TimeIntToString(GetDataStartTime())
}

// 本周的零点
func GetWeekStartStr() string {
	return TimeIntToString(GetWeekStart())
}

// 下周的零点
func GetNextWeekStartStr() int64 {
	return GetWeekStart() + 86400*7
}

// 下一月 零点
func GetNextMonthZero() int64 {
	nowDate := GetIDate(time.Now().Unix())
	endDate := GetMonthEndDateByIdate(nowDate)
	nextMonthDate := IDateInc(endDate, 1)
	return GetDayZeroTimeByIDate(nextMonthDate)
}

// 本月的零点
func GetMonthStart() int64 {
	month := GetIDate(GetDayZeroTime()) / 100      //本月
	return GetTimeByIdate(IntToStr(month*100 + 1)) //本月第一秒
}

// 某时间段月零点
func GetMonthStartByTime(t int64) int64 {
	month := GetIDate(t) / 100                     //本月
	return GetTimeByIdate(IntToStr(month*100 + 1)) //本月第一秒
}

// 下一月 0点
func GetNextMonthZeroByItime(itime int64) int64 {
	nowDate := GetIDate(itime)
	endDate := GetMonthEndDateByIdate(nowDate)
	nextMonthDate := IDateInc(endDate, 1)
	return GetDayZeroTimeByIDate(nextMonthDate) //本月第一秒
}

// 下一月 0点
func GetNextMonthZeroByIdate(idate int64) int64 {
	endDate := GetMonthEndDateByIdate(idate)
	nextMonthDate := IDateInc(endDate, 1)
	return GetDayZeroTimeByIDate(nextMonthDate) //本月第一秒
}

// 某小时起始
func GetHourStartByTime(t int64) int64 {
	timeStr := time.Unix(t, 0).Format(TimeFormat_Hour)
	tt, _ := time.ParseInLocation(TimeFormat_Sec, timeStr+":00:00", time.Local)
	return tt.Unix()
}

// 本周的零点int
func GetWeekStart() int64 {
	todayStart := GetDayZeroTime()
	weekDay := time.Now().Weekday()
	i := 0
	switch weekDay {
	case time.Monday:
		{
			i = 0
		}
	case time.Tuesday:
		{
			i = 1
		}
	case time.Wednesday:
		{
			i = 2
		}
	case time.Thursday:
		{
			i = 3
		}
	case time.Friday:
		{
			i = 4
		}
	case time.Saturday:
		{
			i = 5
		}
	case time.Sunday:
		{
			i = 6
		}
	}
	return todayStart - int64(i*3600*24)
}

// 对应周开始日期零点, 周1 itime
func GetWeekStartItime(itime int64) int64 {
	zero := GetDayZero(itime)
	weekDay := time.Unix(zero, 0).Weekday()
	i := 0
	switch weekDay {
	case time.Monday:
		{
			i = 0
		}
	case time.Tuesday:
		{
			i = 1
		}
	case time.Wednesday:
		{
			i = 2
		}
	case time.Thursday:
		{
			i = 3
		}
	case time.Friday:
		{
			i = 4
		}
	case time.Saturday:
		{
			i = 5
		}
	case time.Sunday:
		{
			i = 6
		}
	}
	tmp := zero - int64(i*3600*24)
	return tmp
}

// 对应周开始日期, 周1 idate
func GetWeekStartByIdate(idate int64) int64 {
	r := idate % 100
	y := (idate % 10000) / 100
	n := idate / 10000
	zero := TimeStrToTime(fmt.Sprintf("%d-%02d-%02d 00:00:00", n, y, r))
	weekDay := time.Unix(zero, 0).Weekday()
	i := 0
	switch weekDay {
	case time.Monday:
		{
			i = 0
		}
	case time.Tuesday:
		{
			i = 1
		}
	case time.Wednesday:
		{
			i = 2
		}
	case time.Thursday:
		{
			i = 3
		}
	case time.Friday:
		{
			i = 4
		}
	case time.Saturday:
		{
			i = 5
		}
	case time.Sunday:
		{
			i = 6
		}
	}
	tmp := zero - int64(i*3600*24)
	return GetIDate(tmp)
}

// 对应月开始日期, idate
func GetMonthStartDateByIdate(idate int64) int64 {
	year := idate / 10000
	mouth := (idate % 10000) / 100
	return year*10000 + mouth*100 + 1
}

// 对应月结束日期, idate
func GetMonthEndDateByIdate(idate int64) int64 {
	year := idate / 10000
	mouth := (idate % 10000) / 100
	day := int64(31)
	if mouth == 4 || mouth == 6 || mouth == 9 || mouth == 11 {
		day = 30
	}
	if mouth == 2 {
		if year%3200 == 0 && year%172800 == 0 {
			day = 29
		} else if year%400 == 0 {
			day = 29
		} else if year%4 == 0 && year%100 != 0 {
			day = 29
		} else {
			day = 28
		}
	}
	return year*10000 + mouth*100 + day
}

// 第几周
func GetWeekByItime(itime int64) int {
	t := time.Unix(itime, 0)
	yearDay := t.YearDay()
	yearFirstDay := t.AddDate(0, 0, -yearDay+1)
	firstDayInWeek := int(yearFirstDay.Weekday())

	//今年第一周有几天
	firstWeekDays := 1
	if firstDayInWeek != 0 {
		firstWeekDays = 7 - firstDayInWeek + 1
	}
	var week int
	if yearDay <= firstWeekDays {
		week = 1
	} else {
		week = (yearDay-firstWeekDays)/7 + 2
	}
	return week
}

// idate 星期几
func GetWeekByIdate(idate int64) int32 {
	r := idate % 100
	y := (idate % 10000) / 100
	n := idate / 10000
	zore := TimeStrToTime(fmt.Sprintf("%d-%02d-%02d 00:00:00", n, y, r))
	weekDay := time.Unix(zore, 0).Weekday()
	return int32(weekDay)
}
func GetWeekDayUnix(sec int64) time.Weekday {
	t := time.Unix(sec, 0)
	return t.Weekday()
}

// 本周的零点int
func GetWeekByStrIdate(idate string) string {
	l := len(idate) //不检查 idate
	newDate := TimeStrToTime(idate[0:l-4] + "-" + idate[l-4:l-2] + "-" + idate[l-2:l] + " 00:00:00")
	weekDay := time.Unix(newDate, 0).Weekday()

	switch weekDay {
	case time.Sunday:
		return "日"
	case time.Monday:
		return "一"
	case time.Tuesday:
		return "二"
	case time.Wednesday:
		return "三"
	case time.Thursday:
		return "四"
	case time.Friday:
		return "五"
	case time.Saturday:
		return "六"
	}
	return ""
}

// idate 转 0点时间戳
func GetTimeByIdate(idate string) int64 {
	l := len(idate) //不检查 idate
	return TimeStrToTime(idate[0:l-4] + "-" + idate[l-4:l-2] + "-" + idate[l-2:l] + " 00:00:00")
}

// idate 转 0点时间戳
func GetStrTimeByIdate(idate string) string {
	l := len(idate) //不检查 idate
	return (idate[0:l-4] + "-" + idate[l-4:l-2] + "-" + idate[l-2:l] + " 00:00:00")
}

// 当前时间距离今日零点的差值 5分钟
func GetTimeMinutr5Diff() int64 {
	t := time.Now().Unix()
	diff := t - GetTimeBegin(t)
	return diff / 60 / 5
}

// 传入指定时间，返回当前时间的零点
func GetTimeBegin(t int64) int64 {
	str := time.Unix(t, 0).Format("2006-01-02")
	tt, _ := time.ParseInLocation(TimeFormat_Sec, str+" 00:00:00", time.Local)
	return tt.Unix()
}

// 转成整分0秒 2006-01-02 23:59:00
func ToTimeZeroSec(t int64) int64 {
	str := time.Unix(t, 0).Format(TimeFormat_Minute)
	tt, _ := time.ParseInLocation(TimeFormat_Sec, str+":00", time.Local)
	return tt.Unix()
}

func GetTimeToZoneUTC4HourMinuteTime(t int64) (string, error) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return "", err
	}
	return time.Unix(t, 0).In(loc).Format(TimeFormat_HourMinuteTime), nil
}

func GetTimeToZoneUTC4Date(t int64) (string, error) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return "", err
	}
	return time.Unix(t, 0).In(loc).Format(TimeFormat_Date), nil
}

// 下季度开始
func GetNextQuarterStart(now time.Time) time.Time {
	quarter := (int(now.Month())-1)/3 + 1
	nextQuarter := quarter + 1
	nextQuarterYear := now.Year()
	if nextQuarter > 4 {
		nextQuarter = 1
		nextQuarterYear++
	}
	nextQuarterMonth := (nextQuarter-1)*3 + 1
	nextQuarterStart := time.Date(nextQuarterYear, time.Month(nextQuarterMonth), 1, 0, 0, 0, 0, time.Local)
	return nextQuarterStart
}

// 下月开始
func GetNextMonthStart(now time.Time) time.Time {
	// 增加一个月并获取年月日
	nextMonth := now.AddDate(0, 1, 0)
	year, month, _ := nextMonth.Date()
	// 构造下个月第一天的时间
	firstOfNextMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	return firstOfNextMonth
}

// 下周开始
func GetNextWeekStart(now time.Time) time.Time {
	// 计算今天是本周的第几天（周日为0，周一为1，...，周六为6）
	weekday := now.Weekday()

	// 计算距离下周一的天数
	daysUntilNextMonday := (8 - weekday) % 7
	if daysUntilNextMonday == 0 {
		daysUntilNextMonday = 7
	}
	// 构造下周一的日期
	nextMonday := now.AddDate(0, 0, int(daysUntilNextMonday))

	// 将时间设置为午夜（0时0分0秒）
	nextMonday = time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 0, 0, 0, 0, time.Local)
	return nextMonday
}
