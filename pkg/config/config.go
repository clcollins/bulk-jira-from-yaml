package config

var AppConfig Config

type Config struct {
	Username string
	Token    string
	Host     string
	Port     int
}
