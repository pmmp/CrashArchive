package database

import "fmt"

func DSN(c *Config) string {
	return c.Username +
		":" +
		c.Password +
		"@tcp(" +
		c.Hostname +
		":" +
		fmt.Sprintf("%d", c.Port) +
		")/" +
		c.Name + c.Parameter
}
