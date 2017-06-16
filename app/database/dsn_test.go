package database

import "testing"

func TestDSN(t *testing.T) {
	c := &Config{"root", "toor", "127.0.0.1", 3306, "crasharchive", "?parseTime=true"}
	expected := "root:toor@tcp(127.0.0.1:3306)/crasharchive?parseTime=true"
	dsn := DSN(c)
	if dsn != expected {
		t.Fatalf("expected %v got %v\n", expected, dsn)
	}
}
