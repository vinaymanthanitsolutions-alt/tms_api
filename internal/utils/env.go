package utils

import (
	"log"
	"os"
)

//  var JwtSecret = []byte(GetEnv("SECRET_KEY"))

func GetEnv(key string) string {

	value := os.Getenv(key)
	var allowed bool = false

	if key == "DB_PASSWORD" {
		allowed = true
	}

	if value == "" && !allowed {
		log.Fatal("Error: " + key + "")
	}

	return value

}
