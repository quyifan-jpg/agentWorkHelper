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
	WS struct {
		Host string `mapstructure:"Host"` // WebSocket 服务地址
		Port int    `mapstructure:"Port"` // WebSocket 服务端口
	} `mapstructure:"WS"`
	Upload struct {
		SavePath string `mapstructure:"SavePath"` // 文件保存路径
		Host     string `mapstructure:"Host"`     // 文件访问主机地址
	} `mapstructure:"Upload"`
}
