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

	bytes, err := ReadRawFile(id)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	report, err := DecodeCrashReport(string(bytes))

	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	return report, nil
}

func ReadRawFile(id int64) ([]byte, error) {
	var err error

	filePath := fmt.Sprintf("reports/%s.log", filenameHash(id))
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("%v\n", err)
		return nil, err
	}

	return ioutil.ReadFile(filePath)
}

func (r *CrashReport) WriteFile(id int64) error {
	filePath := fmt.Sprintf("./reports/%s.log", filenameHash(id))

	return ioutil.WriteFile(filePath, []byte(r.EncodeCrashReport()), os.ModePerm)
}

func filenameHash(id int64) string {
	hash := sha1.Sum([]byte(fmt.Sprintf("%d%s", id, salt)))
	return hex.EncodeToString(hash[:])
}
