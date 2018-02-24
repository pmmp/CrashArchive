package crashreport

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"io/ioutil"
)

var (
	salt             = "pepper"
	reportPathFormat = "reports/%s.bin"
)

// ReadFile reads a zlib blob from disk and decodes it into a CrashReport struct
func ReadFile(id int64) (*CrashReport, error) {
	bytes, err := ReadRawFile(id)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	var report CrashReport
	err = report.ReadZlib(bytes)

	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	return &report, nil
}

// ReadRawFile returns the raw zlib-compressed crashdump bytes stored on disk
func ReadRawFile(id int64) ([]byte, error) {
	var err error

	filePath := fmt.Sprintf(reportPathFormat, filenameHash(id))
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("%v\n", err)
		return nil, err
	}

	return ioutil.ReadFile(filePath)
}

// WriteRawFile writes zlib-compressed crashdump bytes to a file on disk
func WriteRawFile(id int64, zlibBytes []byte) error {
	filePath := fmt.Sprintf(reportPathFormat, filenameHash(id))

	return ioutil.WriteFile(filePath, zlibBytes, os.ModePerm)
}

func filenameHash(id int64) string {
	hash := sha1.Sum([]byte(fmt.Sprintf("%d%s", id, salt)))
	return hex.EncodeToString(hash[:])
}
