package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"math"
	"regexp"
	"sort"

	jsoniter "github.com/json-iterator/go"
	uuid "github.com/satori/go.uuid"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"

	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net"

	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const UidModNum = 64 // 数据库UID取模数字

func GetTabUidModVal(uid uint64) int64 {
	return int64(uid) % UidModNum
}

// 进制 字节
const (
	BASE_TEN     = 10 //10进制
	BASE_TWO     = 2  //2进制
	BASE_SIXTEEN = 16

	BIT_THIRTY_TWO = 32 //32位
	BIT_SIXTY_FOUR = 64 //64位
)

var billSeq uint64

func GetRealMac(mac string) string {
	if mac == "" {
		return mac
	}
	mac = strings.Trim(mac, "[")
	mac = strings.Trim(mac, "]")
	return mac
}

// Substr returns the substr from start to length.
func Substr(s string, start, length int) string {
	bt := []rune(s)
	if start < 0 {
		start = 0
	}
	if start > len(bt) {
		start = start % len(bt)
	}
	var end int
	if (start + length) > (len(bt) - 1) {
		end = len(bt)
	} else {
		end = start + length
	}
	return string(bt[start:end])
}
func GetUUID() string {
	u := uuid.NewV4()
	// if err != nil {
	// 	//logs.Errorf("GetUUID err:%+v", err)
	// 	return ""
	// }
	return fmt.Sprintf("%s", u)
}

func ReplaceSpecial(s string) string {
	chars := []string{"]", "%", "'", "^", "\\\\", "[", ".", "(", ")", "-"}
	r := strings.Join(chars, "")
	re := regexp.MustCompile("[" + r + "]+")
	return re.ReplaceAllString(s, "")
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func SortSlice(data []int32) {
	sort.Slice(data, func(i, j int) bool {
		return data[i] < data[j]
	})
}

func CheckSliceEq(a, b []int32) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func IntsIdInIds(datas []int, ids []int) bool {
	for _, i := range datas {
		for _, v := range ids {
			if v == i {
				return true
			}
		}
	}
	return false
}

// int 转 string
func IntToStr(data int64) string {
	return strconv.FormatInt(data, BASE_TEN)
}

func IntArrToStr(data []int32) string {
	str, _ := jsoniter.MarshalToString(data)
	return str
}

// uint 转 string
func UintToStr(data uint64) string {
	return strconv.FormatInt(int64(data), BASE_TEN)
}

// string 转 int  //出错 就是 0 用的时候 注意
func StrToInt64(data string) int64 {
	ret := int64(0)
	ret, err := strconv.ParseInt(data, BASE_TEN, BIT_SIXTY_FOUR)
	if err != nil {
		return ret
	}
	return ret
}

// StrToInt32Array string 转数组 int32
func StrToInt32Array(data string) []int32 {
	if data == "" {
		return nil
	}
	arr := strings.Split(data, ",")
	var out []int32
	for _, i := range arr {
		a := int32(StrToInt64(i))
		if a > 0 {
			out = append(out, a)
		}
	}
	return out
}
func StrToInt(src string) int {
	_val, _ := strconv.ParseInt(src, BASE_TEN, BIT_SIXTY_FOUR)
	return int(_val)
}

func StrMapToInt(m map[string]string, key string) int {
	if m == nil {
		return 0
	}
	key = strings.ToLower(key)
	src, ok := m[key]
	if !ok {
		return 0
	}
	_val, _ := strconv.ParseInt(src, BASE_TEN, BIT_SIXTY_FOUR)
	return int(_val)
}

func StrMapToInt64(m map[string]string, key string) int64 {
	if m == nil {
		return 0
	}
	key = strings.ToLower(key)
	src, ok := m[key]
	if !ok {
		return 0
	}
	_val, _ := strconv.ParseInt(src, BASE_TEN, BIT_SIXTY_FOUR)
	return _val
}

func StrMapToFloat64(m map[string]string, key string) float64 {
	ret := float64(0)
	if m == nil {
		return ret
	}
	key = strings.ToLower(key)
	src, ok := m[key]
	if !ok {
		return ret
	}
	ret, err := strconv.ParseFloat(src, 64)
	if err != nil {
		return ret
	}
	return ret
}

// string 转 int  //出错 就是 0 用的时候 注意
func StrToUInt(data string) uint64 {
	ret := int64(0)
	ret, err := strconv.ParseInt(data, BASE_TEN, BIT_SIXTY_FOUR)
	if err != nil {
		return uint64(ret)
	}
	return uint64(ret)
}

// 16进制string 转 int  //出错 就是 0 用的时候 注意
func Str16ToInt(data string) int64 {
	ret := int64(0)
	ret, err := strconv.ParseInt(data, BASE_SIXTEEN, BIT_SIXTY_FOUR)
	if err != nil {
		return ret
	}
	return ret
}

func InsertArrayUint64(inArr []uint64, index int, value uint64) []uint64 {
	rear := append([]uint64{}, inArr[index:]...)
	inArr = append(inArr[0:index], value)
	inArr = append(inArr, rear...)
	return inArr
}

func StrToFloat(data string) float64 {
	ret := float64(0)
	ret, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return ret
	}
	return ret
}
func InArray[T comparable](arr []T, x T) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}
func InsertArrayInt32(inArr []int32, index int, value int32) []int32 {
	rear := append([]int32{}, inArr[index:]...)
	inArr = append(inArr[0:index], value)
	inArr = append(inArr, rear...)
	return inArr
}

func ArrayInt32Intersection(arr1, arr2 []int32) []int32 {
	if len(arr1) > len(arr2) {
		return ArrayInt32Intersection(arr2, arr1)
	}
	arr2Map := make(map[int32]bool)
	for _, v := range arr2 {
		arr2Map[v] = true
	}
	arr := make([]int32, 0)
	for _, v := range arr1 {
		if _, ok := arr2Map[v]; ok {
			arr = append(arr, v)
		}
	}
	return arr
}
func ArrayInt32IntersectionExceptEmpty(arr1, arr2 []int32) []int32 {
	if len(arr1) <= 0 {
		return arr2
	}
	if len(arr2) <= 0 {
		return arr1
	}
	return ArrayInt32Intersection(arr1, arr2)
}

func InArrayStringArray(arr []string, x []string) bool {
	for _, v := range arr {
		for _, z := range x {
			if v == z {
				return true
			}
		}
	}
	return false
}
func StringHasPrefix(x string, arr []string) bool {
	for _, v := range arr {
		if strings.HasPrefix(x, v) {
			return true
		}
	}
	return false
}

func StringHasSub(x string, arr []string) bool {
	for _, v := range arr {
		if strings.Index(x, v) >= 0 {
			return true
		}
	}
	return false
}

func GetRandString(arr []string) string {
	if len(arr) <= 0 {
		return ""
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	i := rand.Intn(len(arr))
	return arr[i]
}

// min 到 max之间的随机数
func RandGen(min, max int) int {
	if max < min { //最大小于最小
		return 0
	}
	max = max - min + 1
	//rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.Intn(max) + min
}

// min 到 max之间的随机数
func Rand64Gen(min, max int64) int64 {
	if max < min { //最大小于最小
		return 0
	}
	max = max - min + 1
	return rand.Int63n(max) + min
}

// 随机整数
func RandInt64(min, max int64) int64 {
	if min >= max || max == 0 {
		return max
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.Int63n(max-min) + min
}

// zll 生成指定范围的随机小数,注:很不智能,所以限制min和max前面的数字必须相同例如min:0.00123456 max:0.00222336这样
func RandFloat64(min, max float64) float64 {
	//转为字符串
	my_min := strconv.FormatFloat(min, 'f', -1, 64)
	my_max := strconv.FormatFloat(max, 'f', -1, 64)

	if len(my_min) != len(my_max) {
		return 0
		//fmt.Println("错误")
	} else {
		qianzhui := ""
		houzhui1 := ""
		houzhui2 := ""
		sw := 0
		for i := 0; i < len(my_min); i++ {
			str1 := string(my_min[i])
			str2 := string(my_max[i])
			if sw == 0 {
				if my_min[i] == my_max[i] {
					qianzhui = qianzhui + str1
				} else {
					sw = 1
					houzhui1 = houzhui1 + str1
					houzhui2 = houzhui2 + str2
				}
			} else {
				houzhui1 = houzhui1 + str1
				houzhui2 = houzhui2 + str2
			}
		}
		//return 0,qianzhui+"**"+houzhui1+"**"+houzhui2
		//fmt.Println(qianzhui+"**"+houzhui1+"**"+houzhui2)
		//前缀字符串转整数
		qian1, _ := strconv.ParseInt(houzhui1, 10, 64)
		qian2, _ := strconv.ParseInt(houzhui2, 10, 64)
		//生成随机整数
		jj := RandInt64(qian1, qian2)
		//随机整数转字符串
		jj_str := strconv.FormatInt(jj, 10)
		//前缀拼接随机字符串
		all_str := qianzhui + jj_str
		//全字符串转浮点数
		jieguo, _ := strconv.ParseFloat(all_str, 64)
		return jieguo
	}
	return 0
}

// 获取随机float64 保留2位小数 Notice 不四舍五入
func GetRandomFloat64Ptive(min, max float64) float64 {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	min, max = math.Abs(min), math.Abs(max)
	min, max = GetMinFloat64WHW(min, max), GetMaxFloat64WHW(min, max)
	// 到这里确保 max>=min 并且二者一定是正数
	ret := GetMinFloat64WHW(min, max) + rand.Float64()*(max-min)
	// 不四舍五入
	//ret, _ = decimal.NewFromFloat(ret).RoundFloor(2).Float64()
	ret, _ = decimal.NewFromFloat(ret).Round(3).Float64()
	if ret > max {
		ret = max
	}
	if ret < min {
		ret = min
	}
	return ret
}

func GetRandomFloat64Ntive(min, max float64) float64 {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	min, max = math.Abs(min), math.Abs(max)
	min, max = GetMinFloat64WHW(min, max), GetMaxFloat64WHW(min, max)
	// 到这里确保 max>=min 并且二者一定是正数
	ret := GetMinFloat64WHW(min, max) + rand.Float64()*(max-min)
	// 不四舍五入
	//ret, _ = decimal.NewFromFloat(ret).RoundFloor(2).Float64()
	ret, _ = decimal.NewFromFloat(ret).Round(3).Float64()
	if ret > max {
		ret = max
	}
	if ret < min {
		ret = min
	}
	return -ret
}

func GetMaxFloat64WHW(min, max float64) float64 {
	if min >= max {
		return min
	}
	return max
}

func GetMinFloat64WHW(min, max float64) float64 {
	if min <= max {
		return min
	}
	return max
}

// 万分比概率检查
func CheckTenThousandPercent(peer int) bool {
	if peer <= 0 {
		return false
	}
	r := rand.Intn(10000)
	return r < peer
}

func SplitToInt(s string, sep string) []int {
	arrStr := strings.Split(s, sep)
	arr := make([]int, 0, len(arrStr))
	for _, v := range arrStr {
		if v == "" {
			continue
		}
		x, _ := strconv.Atoi(v)
		if x > 0 {
			arr = append(arr, x)
		}
	}
	return arr
}

func SplitToInt32(s string, sep string) []int32 {
	arrStr := strings.Split(s, sep)
	arr := make([]int32, 0, len(arrStr))
	for _, v := range arrStr {
		if v == "" {
			continue
		}
		x, _ := strconv.Atoi(v)
		if x > 0 {
			arr = append(arr, int32(x))
		}
	}
	return arr
}

// 这个函数慎用 obj 不能用指针
func Struct2Map(obj interface{}) map[string]interface{} {
	var data = make(map[string]interface{})
	if obj == nil {
		return data
	}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	for i := 0; i < t.NumField(); i++ {
		data[ /*strings.ToLower*/ (t.Field(i).Name)] = v.Field(i).Interface()
	}
	return data
}

// 用map填充结构
func FillStruct(data map[string]interface{}, obj interface{}) error {
	for k, v := range data {
		err := SetField(obj, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// 用map的值替换结构的值
func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()        //结构体属性值
	structFieldValue := structValue.FieldByName(name) //结构体单个属性值

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type() //结构体的类型
	val := reflect.ValueOf(value)              //map值的反射值

	var err error
	if structFieldType != val.Type() {
		val, err = TypeConversion(fmt.Sprintf("%v", value), structFieldValue.Type().Name()) //类型转换
		if err != nil {
			return err
		}
	}

	structFieldValue.Set(val)
	return nil
}

func AutoToJsonString(a interface{}) string {
	if a == nil {
		return ""
	}
	switch s := a.(type) {
	case error:
		e, ok := a.(error)
		if ok && e != nil {
			return a.(error).Error()
		} else {
			return ""
		}
	case string:
		return (s)
	case *string:
		return (*s)
	case bool:
		return strconv.FormatBool(s) //fmt.Sprintf("%t", s)
	case *bool:
		return strconv.FormatBool(*s) //fmt.Sprintf("%t", *s)
	case int:
		return strconv.FormatInt(int64(s), 10)
	case *int:
		return strconv.FormatInt(int64(*s), 10)
	case int32:
		return strconv.FormatInt(int64(s), 10)
	case *int32:
		return strconv.FormatInt(int64(*s), 10) //fmt.Sprintf("%d", *s)
	case int64:
		return strconv.FormatInt(s, 10)
	case *int64:
		return strconv.FormatInt((*s), 10) //fmt.Sprintf("%d", *s)
	case float64, float32:
		return fmt.Sprintf("%f", s)
	case *float64:
		return fmt.Sprintf("%f", *s)
	case uint:
		return strconv.FormatUint(uint64(s), 10)
	case *uint:
		return strconv.FormatUint(uint64(*s), 10)
	case uint32:
		return strconv.FormatUint(uint64(s), 10)
	case *uint32:
		return strconv.FormatUint(uint64(*s), 10)
	case uint64:
		return strconv.FormatUint(s, 10)
	case *uint64:
		return strconv.FormatUint((*s), 10)
	default:
		_s, _ := jsoniter.Marshal(s)
		var str bytes.Buffer
		_ = json.Indent(&str, _s, "", "    ")
		return str.String()

	}
}

func AutoToString(a interface{}) string {
	if a == nil {
		return ""
	}
	switch s := a.(type) {
	case error:
		e, ok := a.(error)
		if ok && e != nil {
			return a.(error).Error()
		} else {
			return ""
		}
	case string:
		return (s)
	case *string:
		return (*s)
	case bool:
		return strconv.FormatBool(s) //fmt.Sprintf("%t", s)
	case *bool:
		return strconv.FormatBool(*s) //fmt.Sprintf("%t", *s)
	case int:
		return strconv.FormatInt(int64(s), 10)
	case *int:
		return strconv.FormatInt(int64(*s), 10)
	case int32:
		return strconv.FormatInt(int64(s), 10)
	case *int32:
		return strconv.FormatInt(int64(*s), 10) //fmt.Sprintf("%d", *s)
	case int64:
		return strconv.FormatInt(s, 10)
	case *int64:
		return strconv.FormatInt((*s), 10) //fmt.Sprintf("%d", *s)
	case float64:
		return fmt.Sprintf("%g", s)
	case *float64:
		return fmt.Sprintf("%g", *s)
	case uint:
		return strconv.FormatUint(uint64(s), 10)
	case *uint:
		return strconv.FormatUint(uint64(*s), 10)
	case uint32:
		return strconv.FormatUint(uint64(s), 10)
	case *uint32:
		return strconv.FormatUint(uint64(*s), 10)
	case uint64:
		return strconv.FormatUint(s, 10)
	case *uint64:
		return strconv.FormatUint((*s), 10)
	default:
		_s, _ := jsoniter.MarshalToString(s)
		return _s
	}
}

func AutoToInt(a interface{}) (n int) {
	if a == nil {
		return 0
	}
	switch s := a.(type) {
	case string:
		n, _ = strconv.Atoi(s)
	case *string:
		n, _ = strconv.Atoi(*s)
	case bool:
		if s {
			n = 1
		} else {
			n = 0
		}
	case *bool:
		if *s {
			n = 1
		} else {
			n = 0
		}
	case int:
		n = s
	case *int:
		n = *s
	case int8:
		n = int(s)
	case *int8:
		n = int(*s)
	case int16:
		n = int(s)
	case *int16:
		n = int(*s)
	case int32:
		n = int(s)
	case *int32:
		n = int(*s)
	case int64:
		n = int(s)
	case *int64:
		n = int(*s)
	case float32:
		n = int(s)
	case *float32:
		n = int(*s)
	case float64:
		n = int(s)
	case *float64:
		n = int(*s)
	case uint:
		n = int(s)
	case *uint:
		n = int(*s)
	case uint8:
		n = int(s)
	case *uint8:
		n = int(*s)
	case uint16:
		n = int(s)
	case *uint16:
		n = int(*s)
	case uint32:
		n = int(s)
	case *uint32:
		n = int(*s)
	case uint64:
		n = int(s)
	case *uint64:
		n = int(*s)
	default:
		n = 0
	}
	return
}

func AutoToFloat(a interface{}) (n float64) {
	if a == nil {
		return 0
	}
	switch s := a.(type) {
	case string:
		n, _ = strconv.ParseFloat(s, 64)
	case *string:
		n, _ = strconv.ParseFloat(*s, 64)
	case bool:
		if s {
			n = 1
		} else {
			n = 0
		}
	case *bool:
		if *s {
			n = 1
		} else {
			n = 0
		}
	case int:
		n = float64(s)
	case *int:
		n = float64(*s)
	case int8:
		n = float64(s)
	case *int8:
		n = float64(*s)
	case int16:
		n = float64(s)
	case *int16:
		n = float64(*s)
	case int32:
		n = float64(s)
	case *int32:
		n = float64(*s)
	case int64:
		n = float64(s)
	case *int64:
		n = float64(*s)
	case float32:
		n = float64(s)
	case *float32:
		n = float64(*s)
	case float64:
		n = s
	case *float64:
		n = *s
	case uint:
		n = float64(s)
	case *uint:
		n = float64(*s)
	case uint8:
		n = float64(s)
	case *uint8:
		n = float64(*s)
	case uint16:
		n = float64(s)
	case *uint16:
		n = float64(*s)
	case uint32:
		n = float64(s)
	case *uint32:
		n = float64(*s)
	case uint64:
		n = float64(s)
	case *uint64:
		n = float64(*s)
	default:
		n = 0
	}
	return
}

// 类型转换
func TypeConversion(value string, ntype string) (reflect.Value, error) {
	if ntype == "string" {
		return reflect.ValueOf(value), nil
	} else if ntype == "time.Time" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
		return reflect.ValueOf(t), err
	} else if ntype == "Time" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
		return reflect.ValueOf(t), err
	} else if ntype == "int" {
		i, err := strconv.Atoi(value)
		return reflect.ValueOf(i), err
	} else if ntype == "int8" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(int8(i)), err
	} else if ntype == "int32" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(int64(i)), err
	} else if ntype == "int64" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(i), err
	} else if ntype == "float32" {
		i, err := strconv.ParseFloat(value, 64)
		return reflect.ValueOf(float32(i)), err
	} else if ntype == "float64" {
		i, err := strconv.ParseFloat(value, 64)
		return reflect.ValueOf(i), err
	}

	//else if .......增加其他一些类型的转换

	return reflect.ValueOf(value), errors.New("未知的类型：" + ntype)
}

// 按百分比检查概率
func CheckRateByPer(per int) bool {
	if per <= 0 {
		return false
	}
	if per >= 100 {
		return true
	}
	n := rand.Intn(100)
	return n < per
}

// 按分母检查概率
func CheckRateByBase(b int) bool {
	if b < 1 {
		return false
	}
	return rand.Intn(b) == 0
}

// 万分位检查概率
func CheckRateByPer10000(per int) bool {
	if per <= 0 {
		return false
	}
	if per >= 10000 {
		return true
	}
	n := rand.Intn(10000)
	return n < per
}
func IdsIntToStr(ids []int32) string {
	str := ""
	for _, v := range ids {
		if str == "" {
			str += fmt.Sprintf("%d", v)
		} else {
			str += fmt.Sprintf(",%d", v)
		}
	}
	return str
}

func IdsStrToStr(ids []string) string {
	str := ""
	for _, v := range ids {
		if str == "" {
			str += fmt.Sprintf(" '%v' ", v)
		} else {
			str += fmt.Sprintf(" ,'%v' ", v)
		}
	}
	return str
}

func StrInStrs(str string, strs []string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

/*
*
数据1有任何值在数据2当中
*/
func StrArrInStrArr(arr1, arr2 []string) bool {
	for _, s := range arr1 {
		for _, v := range arr2 {
			if s == v {
				return true
			}
		}
	}
	return false
}

func GenBillNo(prefixFmt string, args ...interface{}) string {
	i := atomic.AddUint64(&billSeq, 1)
	prefix := fmt.Sprintf(prefixFmt, args...)

	return fmt.Sprintf("%s|%03d", prefix, i)
}

func IsSlic(data interface{}) bool {
	if data == nil {
		return false
	}
	v := reflect.ValueOf(data)
	k := v.Kind()
	if k == reflect.Slice {
		return true
	}
	return false
}

// 汉字字符串截取
func TextIntercept(text string, l int) string {
	if text == "" {
		return text
	}
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.ReplaceAll(text, "\n", "")
	nameRune := []rune(text)
	if len(nameRune) <= l {
		return text
	}
	return string(nameRune[0 : len(nameRune)-l])
}

func GetSlice(data interface{}) string {
	if data == "" {
		return ""
	}
	v := reflect.ValueOf(data)
	k := v.Kind()
	if k == reflect.Slice {
		str, err := jsoniter.MarshalToString(data)
		if err != nil {
			return ""
		}
		return str
	}
	return ""
}

// 获取字符串
func GetString(data interface{}) string {
	return fmt.Sprintf("%v", data)
}

func Decimal(a float64) float64 {
	rum, _ := strconv.ParseFloat(fmt.Sprintf("%.02f", a), 64)
	return rum
}
func Decimal3(a float64) float64 {
	rum, _ := strconv.ParseFloat(fmt.Sprintf("%.03f", a), 64)
	return rum
}

// 获取int64 如果不是numbel 类型 返回0
func GetInt64(data interface{}) int64 {
	if data == nil {
		return 0
	}
	v := reflect.ValueOf(data)
	k := v.Kind()
	switch k {
	case reflect.Int:
		if t, ok := data.(int); ok {
			return int64(t)
		}
	case reflect.Int8:
		if t, ok := data.(int8); ok {
			return int64(t)
		}
	case reflect.Int16:
		if t, ok := data.(int16); ok {
			return int64(t)
		}
	case reflect.Int32:
		if t, ok := data.(int32); ok {
			return int64(t)
		}
	case reflect.Int64:
		if t, ok := data.(int64); ok {
			return int64(t)
		}
	case reflect.Uint:
		if t, ok := data.(uint); ok {
			return int64(t)
		}
	case reflect.Uint8:
		if t, ok := data.(uint8); ok {
			return int64(t)
		}
	case reflect.Uint16:
		if t, ok := data.(uint16); ok {
			return int64(t)
		}
	case reflect.Uint32:
		if t, ok := data.(uint32); ok {
			return int64(t)
		}
	case reflect.Uint64:
		if t, ok := data.(uint64); ok {
			return int64(t)
		}
	case reflect.Float64:
		if t, ok := data.(float64); ok {
			return int64(t)
		}
	case reflect.Float32:
		if t, ok := data.(float32); ok {
			return int64(t)
		}
	case reflect.Ptr:
		if t, ok := data.(*int); ok {
			return int64(*t)
		}
		if t, ok := data.(*int8); ok {
			return int64(*t)
		}
		if t, ok := data.(*int16); ok {
			return int64(*t)
		}
		if t, ok := data.(*int32); ok {
			return int64(*t)
		}
		if t, ok := data.(*int64); ok {
			return int64(*t)
		}
		if t, ok := data.(*uint); ok {
			return int64(*t)
		}
		if t, ok := data.(*uint8); ok {
			return int64(*t)
		}
		if t, ok := data.(*uint16); ok {
			return int64(*t)
		}
		if t, ok := data.(*uint32); ok {
			return int64(*t)
		}
		if t, ok := data.(*uint64); ok {
			return int64(*t)
		}
		if t, ok := data.(*float64); ok {
			return int64(*t)
		}
		if t, ok := data.(*float32); ok {
			return int64(*t)
		}
	}
	return 0
}

func GetFloat64(data interface{}) float64 {
	if data == nil {
		return 0
	}
	v := reflect.ValueOf(data)
	k := v.Kind()
	switch k {
	case reflect.Float32:
		if t, ok := data.(float32); ok {
			return float64(t)
		}
	case reflect.Float64:
		if t, ok := data.(float64); ok {
			return float64(t)
		}
	case reflect.Int:
		if t, ok := data.(int); ok {
			return float64(t)
		}
	case reflect.Int8:
		if t, ok := data.(int8); ok {
			return float64(t)
		}
	case reflect.Int16:
		if t, ok := data.(int16); ok {
			return float64(t)
		}
	case reflect.Int32:
		if t, ok := data.(int32); ok {
			return float64(t)
		}
	case reflect.Int64:
		if t, ok := data.(int64); ok {
			return float64(t)
		}
	case reflect.Uint:
		if t, ok := data.(uint); ok {
			return float64(t)
		}
	case reflect.Uint8:
		if t, ok := data.(uint8); ok {
			return float64(t)
		}
	case reflect.Uint16:
		if t, ok := data.(uint16); ok {
			return float64(t)
		}
	case reflect.Uint32:
		if t, ok := data.(uint32); ok {
			return float64(t)
		}
	case reflect.Uint64:
		if t, ok := data.(uint64); ok {
			return float64(t)
		}
	case reflect.Ptr:
		if t, ok := data.(*int); ok {
			return float64(*t)
		}
		if t, ok := data.(*int8); ok {
			return float64(*t)
		}
		if t, ok := data.(*int16); ok {
			return float64(*t)
		}
		if t, ok := data.(*int32); ok {
			return float64(*t)
		}
		if t, ok := data.(*int64); ok {
			return float64(*t)
		}
		if t, ok := data.(*uint); ok {
			return float64(*t)
		}
		if t, ok := data.(*uint8); ok {
			return float64(*t)
		}
		if t, ok := data.(*uint16); ok {
			return float64(*t)
		}
		if t, ok := data.(*uint32); ok {
			return float64(*t)
		}
		if t, ok := data.(*uint64); ok {
			return float64(*t)
		}
	}
	return 0
}

// 读取csv 文件
func ReadCsvData(data string, title ...string) map[string][]string {
	rfile := strings.NewReader(data)
	selfcsv := csv.NewReader(rfile)
	ret := map[string][]string{}
	if selfcsv == nil || len(title) == 0 {
		return ret
	}
	i := int32(0)
	tm := map[string]int{}
	for _, t := range title {
		tm[t] = 0
		ret[t] = []string{}
	}
	idm := map[int]string{}
	for {
		tmp, err := selfcsv.Read()
		if err != nil {
			if err != io.EOF {

			}
			break
		}
		if i == 0 { //是title
			for id, ti := range tmp {
				if _, ok := tm[ti]; ok {
					tm[ti] = id //这个标题在第几行
					idm[id] = ti
				}
			}
		} else {
			for id, dd := range tmp {
				if ti, ok := idm[id]; ok { //有需要
					if old, ok := ret[ti]; ok {
						old = append(old, dd)
						ret[ti] = old
					} else {
						ret[ti] = []string{dd}
					}
				}
			}
		}
		i++
	}
	return ret
}

// 是否是本机Ip
func IsLocalIP(ip string) (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}
	for i := range addrs {
		intf, _, err := net.ParseCIDR(addrs[i].String())
		if err != nil {
			return false, err
		}
		if net.ParseIP(ip).Equal(intf) {
			return true, nil
		}
	}
	return false, nil
}

func TrimStr(str string) string {
	str = strings.Trim(str, " ")
	str = strings.Trim(str, "\n")
	str = strings.Trim(str, "\r")
	return str
}

func GenUrl(host string, port int32, usPort int32) string {
	// if strings.HasPrefix(host, "http") == false || strings.HasPrefix(host, "HTTP") == false {
	// 	host = "https://" + host
	// 	if !(port == 443 || port == 80) {
	// 		host += ":" + strconv.Itoa(int(port))
	// 	}
	// }
	host = strings.TrimSpace(host)
	if port > 0 {
		return fmt.Sprintf("https://%s:%d", host, port)
	}
	if usPort > 0 {
		return fmt.Sprintf("http://%s:%d", host, usPort)
	}

	return ""
}

// 10005 / 10000  = 1 返回 true
func CheckIntMold(a int64, str string) bool {
	b, _ := strconv.ParseInt(str, 10, 64)
	if a <= 0 || b <= 0 {
		return false
	}
	return b/a == 1
}

func GetOsBuildId(str string) string {
	if str == "" {
		return ""
	}
	s := strings.Split(str, "|")
	if len(s) > 2 {
		return s[2]
	}
	return ""
}

// 数组去重
func GetArrayUnique(data []int32) []int32 {
	out := []int32{}
	for _, v := range data {
		if !InArray(out, v) {
			out = append(out, v)
		}
	}
	return out
}

func UniqueIds(ids []int32) []int32 {
	if len(ids) <= 0 {
		return ids
	}
	out := []int32{}
	for _, v := range ids {
		if !InArray(out, v) {
			out = append(out, v)
		}
	}
	return out
}

func UniqueIdsInt64(ids []int64) []int64 {
	if len(ids) <= 0 {
		return ids
	}
	out := []int64{}
	for _, v := range ids {
		if !IdInIdsINT64(v, out) {
			out = append(out, v)
		}
	}
	return out
}

func UniqueIdsUint64(ids []uint64) []uint64 {
	if len(ids) <= 0 {
		return ids
	}
	out := []uint64{}
	for _, v := range ids {
		if !IdInIdsUINT64(v, out) {
			out = append(out, v)
		}
	}
	return out
}

func IdInIdsINT64(id int64, ids []int64) bool {
	for _, v := range ids {
		if id == v {
			return true
		}
	}
	return false
}
func IdInIdsUINT64(id uint64, ids []uint64) bool {
	for _, v := range ids {
		if id == v {
			return true
		}
	}
	return false
}

func UniqueStrIds(ids []string) []string {
	if len(ids) <= 0 {
		return ids
	}
	out := []string{}
	for _, v := range ids {
		if !InArray(out, v) {
			out = append(out, v)
		}
	}
	return out
}

func RemoveArrayItem(arr []int32, remove []int32) []int32 {
	if len(remove) <= 0 {
		return arr
	}
	data := []int32{}
	for _, v := range arr {
		for _, k := range remove {
			if v == k {
				continue
			}
		}
		data = append(data, v)
	}
	return data
}

/*
*
判断是否是IPv6地址
*/
func CheckIPv6(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	fmt.Println(ip)
	if ip == nil {

		return false
	}
	for i := 0; i < len(ipStr); i++ {
		switch ipStr[i] {
		case '.':
			return false
		case ':':
			return true
		}
	}
	return false
}

// 数组 slice 转为逗号分隔的string字符串
func ArraytoString(a interface{}) string {
	b, err := json.Marshal(a)
	if err != nil {
		return string(b)
	}
	return string(b)
}

// TobsonNoNil  去除结构体的空值，转为bson
func TobsonNoNil(sdata interface{}) (ro bson.M, err error) {
	data, err := bson.Marshal(sdata)
	if err != nil {
		return
	}
	ro = bson.M{}
	err = bson.Unmarshal(data, ro)
	if err != nil {
		return
	}
	for k, v := range ro {
		switch v.(type) {
		case string:
			if v.(string) == "" {
				delete(ro, k)
			}
		case int:
			if v.(int) == 0 {
				delete(ro, k)
			}
		case int8:
			if v.(int8) == 0 {
				delete(ro, k)
			}
		case int16:
			if v.(int16) == 0 {
				delete(ro, k)
			}
		case int32:
			if v.(int32) == 0 {
				delete(ro, k)
			}
		case int64:
			if v.(int64) == 0 {
				delete(ro, k)
			}
		case uint64:
			if v.(uint64) == 0 {
				delete(ro, k)
			}
		case float32:
			if v.(float32) == 0 {
				delete(ro, k)
			}
		case float64:
			if v.(float64) == 0 {
				delete(ro, k)
			}
		case []interface{}:
			if len(v.([]interface{})) == 0 {
				delete(ro, k)
			}
		default:
			//
		}
	}
	return
}

// Tobson  保留结构体的空值，转为bson
func Tobson(sdata interface{}) (ro bson.M, err error) {
	data, err := bson.Marshal(sdata)
	if err != nil {
		return
	}
	ro = bson.M{}
	err = bson.Unmarshal(data, ro)
	if err != nil {
		return
	}
	for k, v := range ro {
		switch v.(type) {
		case string:
			ro[k] = strings.Trim(v.(string), " ")
		}
	}
	return
}

func StructConv(in, out any) (err error) {
	rb, err := jsoniter.Marshal(in)
	if err != nil {
		return
	}
	err = jsoniter.Unmarshal(rb, out)
	if err != nil {
		return
	}
	return
}

/*
*
两个数组，B比A长，满足A跟B的前几位一样
*/
func CompareIntSlice(a, b []int) bool {
	if len(a) > len(b) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

// 高精度乘积
func BigMui(x, y float64) int64 {
	amountInt, _ := new(big.Float).Mul(new(big.Float).SetFloat64(x), new(big.Float).SetFloat64(y)).Float64()
	return int64(math.Ceil(amountInt - 0.5))
}

// 高精度除
func BigQuo(x, y float64) int64 {
	amountInt, _ := new(big.Float).Quo(new(big.Float).SetFloat64(x), new(big.Float).SetFloat64(y)).Float64()
	return int64(math.Ceil(amountInt - 0.5))
}

// 单词数量
func CountWords(str string) int {
	arr := strings.Fields(str)
	return len(arr)
}

// 获取代码行号
func GetLine(p ...int) int {
	skip := 1
	if len(p) > 0 {
		skip = p[0]
	}
	_, _, line, _ := runtime.Caller(skip)
	return line
}

func GetCallFile(n int) string {
	pc, _, line, ok := runtime.Caller(1 + n)
	if !ok {
		//_ = "???"
		line = 0
	}
	//if ss := strings.Split(file, "/"); len(ss) > 2 {
	//	file = strings.Join(ss[len(ss)-2:], "/")
	//}
	funcName := runtime.FuncForPC(pc).Name()
	if ss := strings.Split(funcName, "/"); len(ss) > 1 {
		funcName = strings.Join(ss[len(ss)-1:], "/")
	}
	return fmt.Sprintf("%s:%d", funcName, line)
}

// SafeDivDecimal 高精度浮点数运算
func SafeDivDecimal(first interface{}, second interface{}) decimal.Decimal {
	firstDecimal := GetDecimal(first)
	secondDecimal := GetDecimal(second)

	if abs, _ := secondDecimal.Abs().Float64(); abs < 1e-9 {
		//判断被除数是否等于0,防止出现除0操作
		return decimal.NewFromFloat(0)
	}
	divResult := firstDecimal.Div(secondDecimal)
	return divResult
}

// SafeDivFloat 高精度浮点数运算,返回浮点数
func SafeDivFloat(first interface{}, second interface{}) float64 {
	fVal, _ := SafeDivDecimal(first, second).Float64()
	return fVal
}

// GetDecimal 数值转换成Decimal
func GetDecimal(data interface{}) decimal.Decimal {
	if data == nil {
		return decimal.NewFromFloat(0)
	}
	v := reflect.ValueOf(data)
	k := v.Kind()
	switch k {
	case reflect.Int:
		if t, ok := data.(int); ok {
			return decimal.NewFromInt(int64(t))
		}
	case reflect.Int8:
		if t, ok := data.(int8); ok {
			return decimal.NewFromInt(int64(t))
		}
	case reflect.Int16:
		if t, ok := data.(int16); ok {
			return decimal.NewFromInt(int64(t))
		}
	case reflect.Int32:
		if t, ok := data.(int32); ok {
			return decimal.NewFromInt32(t)
		}
	case reflect.Int64:
		if t, ok := data.(int64); ok {
			return decimal.NewFromInt(t)
		}
	case reflect.Uint:
		if t, ok := data.(uint); ok {
			return decimal.NewFromInt(int64(t))
		}
	case reflect.Uint8:
		if t, ok := data.(uint8); ok {
			return decimal.NewFromInt(int64(t))
		}
	case reflect.Uint16:
		if t, ok := data.(uint16); ok {
			return decimal.NewFromInt(int64(t))
		}
	case reflect.Uint32:
		if t, ok := data.(uint32); ok {
			return decimal.NewFromInt(int64(t))
		}
	case reflect.Uint64:
		if t, ok := data.(uint64); ok {
			return decimal.NewFromInt(int64(t))
		}
	case reflect.Float64:
		if t, ok := data.(float64); ok {
			return decimal.NewFromFloat(t)
		}
	case reflect.Float32:
		if t, ok := data.(float32); ok {
			return decimal.NewFromFloat32(t)
		}
	case reflect.Ptr:
		if t, ok := data.(*int); ok {
			return decimal.NewFromInt(int64(*t))
		}
		if t, ok := data.(*int8); ok {
			return decimal.NewFromInt(int64(*t))
		}
		if t, ok := data.(*int16); ok {
			return decimal.NewFromInt(int64(*t))
		}
		if t, ok := data.(*int32); ok {
			return decimal.NewFromInt32(int32(*t))
		}
		if t, ok := data.(*int64); ok {
			return decimal.NewFromInt(int64(*t))
		}
		if t, ok := data.(*uint); ok {
			return decimal.NewFromInt(int64(*t))
		}
		if t, ok := data.(*uint8); ok {
			return decimal.NewFromInt(int64(*t))
		}
		if t, ok := data.(*uint16); ok {
			return decimal.NewFromInt(int64(*t))
		}
		if t, ok := data.(*uint32); ok {
			return decimal.NewFromInt(int64(*t))
		}
		if t, ok := data.(*uint64); ok {
			return decimal.NewFromInt(int64(*t))
		}
		if t, ok := data.(*float64); ok {
			return decimal.NewFromFloat(*t)
		}
		if t, ok := data.(*float32); ok {
			return decimal.NewFromFloat32(*t)
		}
	}
	return decimal.NewFromFloat(0)
}

func Md5(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}
