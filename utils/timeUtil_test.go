package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestGetDayZero(t *testing.T) {
	s := time.Now().Format("2006-01-02")
	fmt.Println(s)
	st, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		panic(err)
	}
	fmt.Println(GetDayZero(st.Unix()))
	fmt.Println(GetDayZeroTime(time.Now()))
}
