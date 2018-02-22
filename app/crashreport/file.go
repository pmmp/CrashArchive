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
	salt string = "pepper"
)

func ReadFile(id int64) (*CrashReport, error) {
	var err error

	filePath := fmt.Sprintf("reports/%s.bin", filenameHash(id))
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("%v\n", err)
		return nil, err
	}

	bytes, err := ioutil.ReadFile(filePath)
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
func (r *CrashReport) WriteFile(id int64, name, email string) error {
	filePath := fmt.Sprintf("./reports/%s.bin", filenameHash(id))

	return ioutil.WriteFile(filePath, r.WriteZlib(), os.ModePerm)
}

func filenameHash(id int64) string {
	hash := sha1.Sum([]byte(fmt.Sprintf("%d%s", id, salt)))
	return hex.EncodeToString(hash[:])
}
