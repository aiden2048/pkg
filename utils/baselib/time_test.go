package baselib

import (
	"fmt"
	"testing"
	"time"
)

func TestGetWeekDate(t *testing.T) {
	w := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}
	ds := []string{"20180514", "20180515", "20180516", "20180517", "20180518", "20180519", "20180520"}
	for _, d := range ds {
		now, _ := time.ParseInLocation(NumDateFormat, d, time.Local)
		date := Date{now}
		for i := range w {
			// fmt.Println(date, w[i], ds[i])
			if date.GetWeekDate(w[i]).NumString() != ds[i] {
				t.Fail()
			}
		}
	}
}

func TestNewDateByInt(t *testing.T) {
	d := 20180711
	date := NewDateByInt(d)
	fmt.Println(date)
	if date.YYYYMMDD() != "20180711" {
		t.Fail()
	}
}
