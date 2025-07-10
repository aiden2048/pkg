package utils

import (
	"os"
	"testing"
)

func TestZip(t *testing.T) {
	f1, err := os.Open("./zip_test_txt.txt")
	if err != nil {
		panic(err)
	}
	err = Zip([]*os.File{f1}, f1.Name()+".zip")
	if err != nil {
		panic(err)
	}
}

func TestUnzip(t *testing.T) {
	err := Unzip("./zip_test_txt.txt.zip", "./")
	if err != nil {
		panic(err)
	}
}
