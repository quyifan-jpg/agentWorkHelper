package config

type Config struct {
	Name string
	Addr string
	MySQL struct {
		DSN string
	}
}
