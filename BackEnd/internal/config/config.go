package config

type Config struct {
	Name string `mapstructure:"Name"`
	Host string `mapstructure:"Host"`
	Port int    `mapstructure:"Port"`
	MySQL struct {
		DSN string `mapstructure:"DSN"`
	} `mapstructure:"MySQL"`
	Auth struct {
		Secret string `mapstructure:"Secret"`
		Expire int64  `mapstructure:"Expire"`
	} `mapstructure:"Auth"`
}
