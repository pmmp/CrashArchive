package crashreport

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var (
	salt string = "pepper"
)

func ReadFile(id int64) (*CrashReport, map[string]interface{}, error) {
	var err error

	var jsonData map[string]interface{}
	filePath := fmt.Sprintf("reports/%s.log", filenameHash(id))
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("%v\n", err)
		//app.tmpl.ExecuteTemplate(w, "error", map[string]interface{}{
		//	"Message": "Report not found",
		//	"URL":     "/home",
		//})
		return nil, jsonData, err
	}

	fin, err := os.Open(filePath)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, jsonData, err
	}
	err = json.NewDecoder(fin).Decode(&jsonData)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, jsonData, err
	}
	// report, err := app.DB.GetReport(reportID)
	//log.Printf("%#v\n", jsonData)

	report, err := DecodeCrashReport(jsonData["report"].(string))
	if err != nil {
		log.Printf("%v\n", err)
		return nil, jsonData, err
	}

	return report, jsonData, nil
}
func (r *CrashReport) WriteFile(id int64, name, email string) error {
	data := map[string]interface{}{
		"report":        r.EncodeCrashReport(),
		"reportId":      id,
		"name":          name,
		"email":         email,
		"attachedIssue": false,
	}

	fout, err := os.Create(fmt.Sprintf("./reports/%s.log", filenameHash(id)))
	if err != nil {
		return err
	}
	defer fout.Close()

	return json.NewEncoder(fout).Encode(data)
}

func filenameHash(id int64) string {
	hash := sha1.Sum([]byte(fmt.Sprintf("%d%s", id, salt)))
	return hex.EncodeToString(hash[:])
}
