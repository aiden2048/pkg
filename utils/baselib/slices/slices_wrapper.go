package slices

// InsertInt is  add value just ahead specified index
func InsertInt(slice *[]int, index int, value int) {
	rear := append([]int{}, (*slice)[index:]...)
	*slice = append(append((*slice)[:index], value), rear...)
}

// InsertInt32 is  add value just ahead specified index
func InsertInt32(slice *[]int32, index int, value int32) {
	rear := append([]int32{}, (*slice)[index:]...)
	*slice = append(append((*slice)[:index], value), rear...)
}

// InsertInt64 is  add value just ahead specified index
func InsertInt64(slice *[]int64, index int, value int64) {
	rear := append([]int64{}, (*slice)[index:]...)
	*slice = append(append((*slice)[:index], value), rear...)
}

// InsertUint64 is  add value just ahead specified index
func InsertUint64(slice *[]uint64, index int, value uint64) {
	rear := append([]uint64{}, (*slice)[index:]...)
	*slice = append(append((*slice)[:index], value), rear...)
}

// InsertString is add value just ahead specified index
func InsertString(slice *[]string, index int, value string) {
	rear := append([]string{}, (*slice)[index:]...)
	*slice = append(append((*slice)[:index], value), rear...)
}
