package database

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"time"

	"facette.io/natsort"

	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func New(config *Config) (*DB, error) {
	if config.Username == "" || config.Password == "" {
		return nil, errors.New("Username and password for mysql database not set in config.json")
	}
	db, err := sqlx.Connect("mysql", DSN(config))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping db")
	}

	return &DB{db}, nil
}

func (db *DB) UpdateTables() {
	log.Printf("updating tables")
	db.Exec("ALTER TABLE crash_report_blobs ADD COLUMN access_token CHAR(32) DEFAULT ''")
	db.Exec("ALTER TABLE crash_reports ADD COLUMN fork BOOL DEFAULT FALSE")
	db.Exec("ALTER TABLE crash_reports ADD COLUMN modified BOOL DEFAULT FALSE")
	db.Exec("DROP INDEX duplicate ON crash_reports")
	db.Exec("CREATE INDEX bool_filters ON crash_reports (duplicate, fork, modified)")
	log.Printf("finished updating tables")
}

var queryInsertReport = `INSERT INTO crash_reports
		(plugin, pluginInvolvement, version, build, file, message, line, type, os, submitDate, reportDate, duplicate, reporterName, reporterEmail, fork, modified)
	VALUES
	(:plugin, :pluginInvolvement, :version, :build, :file, :message, :line, :type, :os, :submitDate, :reportDate, :duplicate, :reporterName, :reporterEmail, :fork, :modified)`
const queryInsertBlob = `INSERT INTO crash_report_blobs (id, crash_report_json, access_token) VALUES (?, ?, ?)`

func (db *DB) InsertReport(report *crashreport.CrashReport, reporterName string, reporterEmail string, originalData []byte, accessToken string) (int64, error) {
	res, err := db.NamedExec(queryInsertReport, &crashreport.Report{
		Plugin:            report.Data.Plugin,
		PluginInvolvement: report.Data.PluginInvolvement,
		Version:           report.Version.Get(true),
		Build:             report.Version.Build,
		File:              report.Error.File,
		Message:           report.Error.Message,
		Line:              report.Error.Line,
		Type:              report.Error.Type,
		OS:                report.Data.General.OS,
		SubmitDate:        time.Now().Unix(),
		ReportDate:        report.ReportDate.Unix(),
		Duplicate:         report.Duplicate,
		ReporterName:      reporterName,
		ReporterEmail:     reporterEmail,
		Fork:              report.Fork,
		Modified:          report.Modified,
	})

	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, fmt.Errorf("failed to get last insert ID: %d", id)
	}

	stmt, err := db.Preparex(queryInsertBlob)
	if err != nil {
		return -1, err
	}

	var zlibBuf bytes.Buffer
	zw, _:= zlib.NewWriterLevel(&zlibBuf, zlib.BestCompression)
	_, err = zw.Write(originalData)
	if err != nil {
		return -1, err
	}
	zw.Close()

	_, err = stmt.Exec(id, zlibBuf.Bytes(), accessToken)
	defer stmt.Close()
	if err != nil {
		return -1, fmt.Errorf("failed to execute prepared statement: %v", err)
	}

	return id, nil
}

func (db *DB) FetchRawReport(id int64) ([]byte, string, error) {
	query := "SELECT crash_report_json, access_token FROM crash_report_blobs WHERE id = ?;"
	var result struct {
		ZlibBytes []byte `db:"crash_report_json"`
		AccessToken string `db:"access_token"`
	}
	err := db.Get(&result, query, id)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	zr, err := zlib.NewReader(bytes.NewReader(result.ZlibBytes))
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	defer zr.Close()

	decompressBytes, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decompress data from db: %v", err)
	}
	return decompressBytes, result.AccessToken, nil
}

func (db *DB) FetchReport(id int64) (*crashreport.CrashReport, string, error) {
	bytes, accessToken, err := db.FetchRawReport(id)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	report, err := crashreport.FromJson(bytes)
	if err != nil {
		return nil, "", err
	}
	return report, accessToken, nil
}

func (db *DB) CheckDuplicate(report *crashreport.CrashReport) (bool, error) {
	var dupes int
	err := db.Get(&dupes, "SELECT COUNT(id) FROM crash_reports WHERE message = ? AND file = ? AND line = ? AND duplicate = false;", report.Error.Message, report.Error.File, report.Error.Line)
	if err != nil {
		return false, err
	}

	return dupes != 0, nil
}

func (db *DB) AuthenticateUser(username string, password []byte) (user.UserInfo, error) {
	var result struct {
		PasswordHash []byte `db:"passwordHash"`
		Permission int `db:"permission"`
	}
	err := db.Get(&result, "SELECT passwordHash, permission FROM users WHERE username = ? LIMIT 1", username);
	if err != nil {
		return user.DefaultUserInfo(), fmt.Errorf("database error: %v", err)
	}
	err = user.VerifyPassword(result.PasswordHash, password)
	if err != nil {
		return user.DefaultUserInfo(), fmt.Errorf("failed to verify password: %v", err)
	}
	return user.UserInfo{
		Name: username,
		Permission: user.UserPermission(result.Permission),
	}, nil
}

func (db *DB) AddUser(username string, password []byte, permission user.UserPermission) error {
	passwordHash, err := user.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	_, err2 := db.Query(
		"INSERT INTO users (username, passwordHash, permission) VALUES (?, ?, ?)",
		username,
		passwordHash,
		int(permission),
	)
	return err2
}

func (db *DB) GetKnownVersions() ([]string, error) {
	knownVersions := []string{}
	err := db.Select(&knownVersions, `SELECT DISTINCT version FROM crash_reports`)
	if err != nil {
		return nil, fmt.Errorf("error fetching known versions: %v\n", err)
	}
	log.Printf("Found %d known versions\n", len(knownVersions))
	reverseNatsort := func(a, b int) bool {
		return natsort.Compare(knownVersions[b], knownVersions[a])
	}
	sort.Slice(knownVersions, reverseNatsort)
	return knownVersions, nil
}
