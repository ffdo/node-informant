package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io/ioutil"
	"os"
)

func DecompressGZip(in []byte) (data []byte, err error) {
	ir := bytes.NewReader(in)
	r, err := gzip.NewReader(ir)
	if err != nil {
		return
	}
	defer r.Close()
	data, err = ioutil.ReadAll(r)
	return
}

func Deflate(in []byte) (data []byte, err error) {
	ir := bytes.NewReader(in)
	r := flate.NewReader(ir)
	if err != nil {
		return
	}
	defer r.Close()
	data, err = ioutil.ReadAll(r)
	return
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
