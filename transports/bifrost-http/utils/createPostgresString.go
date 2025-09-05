package utils

import "fmt"

func CreatePostgresLink(
	Host string,
	Port int,
	User string,
	Password string,
	DBName string,
	SSLMode string,
) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		User, Password, Host, Port, DBName, SSLMode,
	)
}
