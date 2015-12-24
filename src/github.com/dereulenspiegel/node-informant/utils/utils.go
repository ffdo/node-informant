package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
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
	data, err = ioutil.ReadAll(r)
	r.Close()
	return
}

func DeflateCompress(in []byte) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	flateWriter, err := flate.NewWriter(buf, flate.BestCompression)
	if err != nil {
		return nil, err
	}
	n, err := flateWriter.Write(in)
	if err != nil {
		return nil, err
	}
	if n != len(in) {
		return nil, fmt.Errorf("Wrote less bytes to flate compressor than data available (data %d bytes, written %d bytes)", n, len(in))
	}
	err = flateWriter.Flush()
	if err != nil {
		return nil, err
	}
	err = flateWriter.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
