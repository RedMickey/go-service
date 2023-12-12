package dbAdapter

import (
	"os"
	"strconv"
)

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

func GetPgDbConfig() DbConfig {
	port, err := strconv.Atoi(os.Getenv("Port"))

	if err != nil {
		panic(err)
	}

	return DbConfig{
		os.Getenv("Host"),
		port,
		os.Getenv("User"),
		os.Getenv("Password"),
		os.Getenv("Dbname"),
	}
}
