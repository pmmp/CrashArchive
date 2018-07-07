package database

type Config struct {
	Username  string `json:"Username"`
	Password  string `json:"Password"`
	Hostname  string `json:"Hostname"`
	Port      int    `json:"Port"`
	Name      string `json:"Name"`
	Parameter string `json:"Parameter"`
}
