package baselib

import (
	"fmt"
	"strconv"
	"time"
)

const (
	NumMonthFormat  = "200601"
	NumDateFormat   = "20060102"
	NumHMDFormat    = "150405"
	NumHMFormat     = "1504"
	DateFormat      = "2006-01-02"
	TimeFormat      = "2006-01-02 15:04:05"
	TimeFormatMilli = "2006-01-02 15:04:05.000"
	TimeFormatMicro = "2006-01-02 15:04:05.000000"
	TimeFormatNano  = "2006-01-02 15:04:05.000000000"
	TimeForBillMill = "20060102150405.000"
)

// 日期如"2015-10-01 21:00:00" 转成时间戳
func ConvertDateToTimestamp(date string) int64 {
	timeTmp, _ := time.Parse(TimeFormat, date)
	return timeTmp.Unix() - 3600*8
}

type Time struct {
	time.Time
}

func Now() Time {
	return Time{time.Now()}
}

func NewTime(timestamp int64) Time {
	tm := time.Unix(timestamp, 0)
	return Time{tm}
}

func StrNewTime(timeStr string) Time {
	now, _ := StrNewTimeEx(timeStr)
	return now
}

func StrNewTimeEx(timeStr string) (Time, error) {
	if timeStr == zeroTimeStr {
		return zeroTime, nil
	}
	now, err := time.ParseInLocation(TimeFormat, timeStr, time.Local)
	if err != nil {
		return zeroTime, err
	}
	return Time{now}, nil
}

var zeroTime Time
var zeroTimeStr = "0000-00-00 00:00:00"
var zeroTimeJsonStr = `"` + zeroTimeStr + `"`

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	str := string(data)
	if str == zeroTimeJsonStr {
		*t = zeroTime
		return
	}
	now, err := time.ParseInLocation(`"`+TimeFormat+`"`, str, time.Local)
	*t = Time{now}
	return
}

func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(TimeFormat)+2)
	if t.IsZero() {
		b = append(b, []byte(zeroTimeJsonStr)...)
		return b, nil
	}
	b = append(b, '"')
	b = t.AppendFormat(b, TimeFormat)
	b = append(b, '"')
	return b, nil
}

func (t Time) String() string {
	if t.IsZero() {
		return zeroTimeStr
	}
	return t.Format(TimeFormat)
}

func (t Time) DateStr() string {
	return t.Format(DateFormat)
}

func (t Time) IsZero() bool {
	return t.Equal(zeroTime.Time)
}

func (t Time) BetweenTime(beginTime, endTime time.Time) bool {
	return (beginTime.Before(t.Time) || beginTime.Equal(t.Time)) && (endTime.After(t.Time) || endTime.Equal(t.Time))
}

func (t Time) Between(beginTime, endTime Time) bool {
	return t.BetweenTime(beginTime.Time, endTime.Time)
}

// YYYYMM 返回年月，如201805
func (t Time) YYYYMM() string {
	return t.Format(NumMonthFormat)
}

// YYYYMMDD 返回年月日，如20180504
func (t Time) YYYYMMDD() string {
	return t.Format(NumDateFormat)
}

// HHMMDD 返回时分秒 ，如120304
func (t Time) HHMMDD() string {
	return t.Format(NumHMDFormat)
}

// HHMM 返回时分秒 ，如120304
func (t Time) HHMM() string {
	return t.Format(NumHMFormat)
}

type Date struct {
	time.Time
}

func NowDate() Date {
	return Date{time.Now()}
}

func NewDate(timestamp int64) Date {
	tm := time.Unix(timestamp, 0)
	return Date{tm}
}

// NewDateByInt 参数格式: 20180710
func NewDateByInt(date int) Date {
	dateStr := fmt.Sprintf("%v", date)
	now, _ := time.ParseInLocation(NumDateFormat, dateStr, time.Local)
	return Date{now}
}

var zeroDate Date
var zeroDateStr = "0000-00-00"
var zeroDateJsonStr = `"` + zeroDateStr + `"`

func (t *Date) UnmarshalJSON(data []byte) (err error) {
	str := string(data)
	if str == zeroDateJsonStr {
		*t = zeroDate
		return
	}
	now, err := time.ParseInLocation(`"`+DateFormat+`"`, string(str), time.Local)
	*t = Date{now}
	return
}

func (t *Date) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(DateFormat)+2)
	if t.IsZero() {
		b = append(b, []byte(zeroDateJsonStr)...)
		return b, nil
	}
	b = append(b, '"')
	b = t.AppendFormat(b, DateFormat)
	b = append(b, '"')
	return b, nil
}

func (t Date) BeforeOrEqual(date Date) bool {
	return t.YYYYMMDD() <= date.YYYYMMDD()
}

func (t Date) Between(beginDate, endDate Date) bool {
	return beginDate.BeforeOrEqual(t) && t.BeforeOrEqual(endDate)
}

func (t Date) IsZero() bool {
	return t.Equal(zeroDate.Time)
}

func (t Date) String() string {
	if t.IsZero() {
		return zeroDateStr
	}
	return t.Format(DateFormat)
}

func (t Date) NumString() string {
	return t.Format(NumDateFormat)
}

// YYYYMM 返回年月，如201805
func (t Date) YYYYMM() string {
	return t.Format(NumMonthFormat)
}

// YYYYMMDD 返回年月日，如20180504
func (t Date) YYYYMMDD() string {
	return t.Format(NumDateFormat)
}

// GetWeekDate 获取当周某天
func (t Date) GetWeekDate(w time.Weekday) Date {
	d := w - t.Weekday()
	// 在中国，周日是最后一天，因此针对周日做个修正
	if w == time.Sunday && d != 0 {
		d += 7
	} else if t.Weekday() == time.Sunday && d != 0 {
		d -= 7
	}
	return Date{t.AddDate(0, 0, int(d))}
}

// NQGtime2Date00 将当前时刻转换为今天的00:00
func NQGtime2Date00(specTime *Time) *Date {
	dateStrSpec := `"` + specTime.DateStr() + `"`
	dateSpecTime00 := Date{}
	dateSpecTime00.UnmarshalJSON([]byte(dateStrSpec))
	return &dateSpecTime00
}

//本周一的年月日， 如20180402
func GetMondayDateStr(tm time.Time) string {
	d := Date{tm}
	return d.GetWeekDate(time.Monday).NumString()
}

func GetThisMondayStr() string {
	return GetMondayDateStr(time.Now())
}

//本周天的年月日， 如20180408
func GetSunDayDateStr(tm time.Time) string {
	d := Date{tm}
	return d.GetWeekDate(time.Sunday).NumString()
}
func GetThisSundayStr() string {
	return GetSunDayDateStr(time.Now())
}

//本月第一天/最后一天的年月日， 如20180401
func GetFirstDayOfMonth(tm time.Time) string {
	currentYear, currentMonth, _ := tm.Date()
	currentLocation := tm.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	return fmt.Sprintf("%04d%02d%02d", firstOfMonth.Year(), firstOfMonth.Month(), firstOfMonth.Day())
}
func GetFirstOfThisMonth() string {
	return GetFirstDayOfMonth(time.Now())
}

func GetLastDayOfMonth(tm time.Time) string {
	currentYear, currentMonth, _ := tm.Date()
	currentLocation := tm.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	return fmt.Sprintf("%04d%02d%02d", lastOfMonth.Year(), lastOfMonth.Month(), lastOfMonth.Day())
}
func GetLastOfThisMonth() string {
	return GetLastDayOfMonth(time.Now())
}

//获取本月最后一天的时间戳
func GetLastDayTimestampOfMonth(tm time.Time) int64 {
	currentYear, currentMonth, _ := tm.Date()
	currentLocation := tm.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	return time.Date(lastOfMonth.Year(), lastOfMonth.Month(), lastOfMonth.Day(), 0, 0, 0, 0, time.Local).Unix()
}

//获取距今n天的年月日
func GetDayString(n int) string {
	nTime := time.Now()
	wantTime := nTime.AddDate(0, 0, n)
	return wantTime.Format("20060102")
}

//获取当天/昨天/明天的年月日， 如20180403
func GetTodayString() string {
	return GetDayString(0)
}
func GetYesterdayString() string {
	return GetDayString(-1)
}
func GetTomorrowString() string {
	return GetDayString(1)
}

//判断是否是闰年
func IsLeapYear(y string) bool {
	year, _ := strconv.Atoi(y)
	if year%4 == 0 && year%100 != 0 || year%400 == 0 {
		return true
	}

	return false
}

//判断是否同一天
func IsOneDay(time1, time2 int64) bool {
	tm1 := time.Unix(time1, 0)
	tm2 := time.Unix(time2, 0)
	if tm1.Format("20060102") != tm2.Format("20060102") {
		return false
	}
	return true
}

//判断是否同一小时
func IsOneHour(time1, time2 int64) bool {
	tm1 := time.Unix(time1, 0)
	tm2 := time.Unix(time2, 0)
	if tm1.Format("20060102 03 PM") != tm2.Format("20060102 03 PM") {
		return false
	}
	return true
}

//判断是否同一分钟
func IsOneMinute(time1, time2 int64) bool {
	tm1 := time.Unix(time1, 0)
	tm2 := time.Unix(time2, 0)
	if tm1.Format("20060102 03:04 PM") != tm2.Format("20060102 03:04 PM") {
		return false
	}
	return true
}

//判断是否同一周
func IsOneWeek(time1, time2 int64) bool {
	tm1 := time.Unix(time1, 0)
	tm2 := time.Unix(time2, 0)
	if GetMondayDateStr(tm1) != GetMondayDateStr(tm2) {
		return false
	}
	return true
}

//判断是否同一个月
func IsOneMonth(time1, time2 int64) bool {
	tm1 := time.Unix(time1, 0)
	tm2 := time.Unix(time2, 0)
	if tm1.Format("200601") != tm2.Format("200601") {
		return false
	}
	return true
}

//判断是否同一年
func IsOneYear(time1, time2 int64) bool {
	tm1 := time.Unix(time1, 0)
	tm2 := time.Unix(time2, 0)
	if tm1.Format("2006") != tm2.Format("2006") {
		return false
	}
	return true
}
