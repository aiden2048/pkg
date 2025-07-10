package console

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	isParsedOptions bool
	optionMap       = make(map[string]string)
)

func parseOptions() {
	if isParsedOptions {
		return
	}

	argNum := len(os.Args)
	for i := 1; i < argNum; i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			option := strings.Trim(os.Args[i], "-")
			nextArgIdx := i + 1
			if i+1 < argNum {
				if strings.HasPrefix(os.Args[nextArgIdx], "-") {
					optionMap[option] = ""
				} else {
					optionMap[option] = os.Args[nextArgIdx]
				}
			} else {
				optionMap[option] = ""
			}
		}
	}

	isParsedOptions = true
}

func OptInt(name string) int {
	val := OptStr(name)
	intVal, err := strconv.Atoi(val)
	if err == nil {
		return intVal
	}
	fmt.Printf("opt %s required type int, input %s\n\n", name, val)
	os.Exit(1)
	return 0
}

func OptFloat32(name string) float32 {
	val := OptStr(name)
	intVal, err := strconv.ParseFloat(val, 32)
	if err == nil {
		return float32(intVal)
	}
	fmt.Printf("opt %s required type uint32, input %s\n\n", name, val)
	os.Exit(1)
	return 0
}

func OptFloat64(name string) float64 {
	val := OptStr(name)
	intVal, err := strconv.ParseFloat(val, 64)
	if err == nil {
		return intVal
	}
	fmt.Printf("opt %s required type uint32, input %s\n\n", name, val)
	os.Exit(1)
	return 0
}

func OptUint32(name string) uint32 {
	val := OptStr(name)
	intVal, err := strconv.ParseUint(val, 10, 32)
	if err == nil {
		return uint32(intVal)
	}
	fmt.Printf("opt %s required type uint32, input %s\n\n", name, val)
	os.Exit(1)
	return 0
}

func OptUint32Default(name string, def uint32) uint32 {
	val := OptStrDefault(name, "")
	if val == "" {
		return def
	}
	intVal, err := strconv.ParseUint(val, 10, 32)
	if err == nil {
		return uint32(intVal)
	}
	fmt.Printf("opt %s required type uint32, input %s\n\n", name, val)
	os.Exit(1)
	return 0
}

func OptUint64(name string) uint64 {
	val := OptStr(name)
	intVal, err := strconv.ParseUint(val, 10, 64)
	if err == nil {
		return intVal
	}
	fmt.Printf("opt %s required type uint64, input %s\n\n", name, val)
	os.Exit(1)
	return 0
}

func OptUint64Default(name string, def uint64) uint64 {
	val := OptStrDefault(name, "")
	if val == "" {
		return def
	}
	intVal, err := strconv.ParseUint(val, 10, 64)
	if err == nil {
		return intVal
	}
	fmt.Printf("opt %s required type uint64, input %s\n\n", name, val)
	os.Exit(1)
	return 0
}

func OptInt64(name string) int64 {
	val := OptStr(name)
	intVal, err := strconv.ParseInt(val, 10, 64)
	if err == nil {
		return intVal
	}
	fmt.Printf("opt %s required type int, input %s, err %s\n\n", name, val, err)
	os.Exit(1)
	return 0
}

func OptInt64Default(name string, d uint64) int64 {
	val := OptStrDefault(name, fmt.Sprintf("%d", d))
	intVal, err := strconv.ParseInt(val, 10, 64)
	if err == nil {
		return intVal
	}
	fmt.Printf("opt %s required type int, input %s, err %s\n\n", name, val, err)
	os.Exit(1)
	return 0
}

func OptIntDefault(name string, d int) int {
	val := OptStrDefault(name, strconv.Itoa(d))
	intVal, err := strconv.Atoi(val)
	if err == nil {
		return intVal
	}
	fmt.Printf("opt %s required type int, input %s\n", name, val)
	os.Exit(1)
	return d
}

func OptStr(name string) string {
	parseOptions()
	val, ok := optionMap[name]
	if ok {
		return val
	}
	fmt.Printf("missed parameter -%s\n\n", name)
	os.Exit(1)
	return ""
}

func OptStrDefault(name string, d string) string {
	parseOptions()
	val, ok := optionMap[name]
	if ok {
		return val
	}
	return d
}

func OptStrSlice(name string) []string {
	str := OptStrDefault(name, "")
	slice := strings.Split(str, ",")
	return slice
}

func OptUint32Slice(name string) []uint32 {
	strSlice := OptStrSlice(name)
	intSlice := make([]uint32, 0)
	for _, str := range strSlice {
		intVal, err := strconv.ParseUint(str, 10, 32)
		if err == nil {
			intSlice = append(intSlice, uint32(intVal))
		} else {
			fmt.Printf("opt %s required type int, input %s\n", name, str)
			os.Exit(1)
		}
	}
	return intSlice
}

func OptUint64Slice(name string) []uint64 {
	strSlice := OptStrSlice(name)
	var intSlice []uint64
	for _, str := range strSlice {
		intVal, err := strconv.ParseUint(str, 10, 64)
		if err == nil {
			intSlice = append(intSlice, intVal)
		} else {
			fmt.Printf("opt %s required type int, input %s\n", name, str)
			os.Exit(1)
		}
	}
	return intSlice
}

func OptUint64SliceDefault(name string, def []uint64) []uint64 {
	val := OptStrDefault(name, "")
	if val == "" {
		return def
	}
	strSlice := strings.Split(val, ",")
	var intSlice []uint64
	for _, str := range strSlice {
		intVal, err := strconv.ParseUint(str, 10, 64)
		if err == nil {
			intSlice = append(intSlice, intVal)
		} else {
			fmt.Printf("opt %s required type int, input %s\n", name, str)
			os.Exit(1)
		}
	}
	return intSlice
}

func OptInt64Slice(name string) []int64 {
	strSlice := OptStrSlice(name)
	var intSlice []int64
	for _, str := range strSlice {
		intVal, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			intSlice = append(intSlice, intVal)
		} else {
			fmt.Printf("opt %s required type int, input %s\n", name, str)
			os.Exit(1)
		}
	}
	return intSlice
}

func OptHas(name string) bool {
	_, ok := optionMap[name]
	return ok
}

func OptBool(name string) bool {
	val := OptStr(name)
	val = strings.ToLower(val)
	if val == "0" || val == "false" {
		return false
	}
	return true
}

func OptBoolDefault(name string, b bool) bool {
	val := OptStrDefault(name, "")
	if len(val) == 0 {
		return b
	}
	val = strings.ToLower(val)
	if val == "0" || val == "false" {
		return false
	}
	return true
}

func OptStrPrompt(tips string) (string, error) {
	fmt.Println()
	fmt.Println(tips)
	var optValue string
	_, err := fmt.Scanf("%s", &optValue)
	if err != nil {
		return "", err
	}
	return optValue, nil
}
