package crashreport

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"io/ioutil"
)

const (
	salt = "pepper"
	filePathFmt = "reports/%s.log"
)

// ReadAndDecode reads a stored JSON report, decodes it and returns the decoded report
func ReadAndDecode(id int64) (*CrashReport, error) {
	bytes, err := ReadReportJson(id)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	report, err := FromJson(bytes)

	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	return report, nil
}

// ReadReportJson reads a JSON report blob from storage and returns it
func ReadReportJson(id int64) ([]byte, error) {
	var err error

	filePath := fmt.Sprintf(filePathFmt, filenameHash(id))
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("%v\n", err)
		return nil, err
	}

	return ioutil.ReadFile(filePath)
}

// WriteReportJson writes a JSON report blob to storage
func WriteReportJson(id int64, jsonBytes []byte) error {
	filePath := fmt.Sprintf(filePathFmt, filenameHash(id))

	return ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
}

func filenameHash(id int64) string {
	hash := sha1.Sum([]byte(fmt.Sprintf("%d%s", id, salt)))
	return hex.EncodeToString(hash[:])
}
