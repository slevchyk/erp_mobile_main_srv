package models

type Config struct {
	Auth AuthConfig `json:"auth"`
	DB   DBConfig `json:"db"`
	WinService winService `json:"win_service"`
}

type AuthConfig struct {
	User string
	Password string
}

type DBConfig struct {
	Name string
	User string
	Password string
}

type winService struct {
	Name        string `json:"name"`
	LongName    string `json:"long_name"`
	Description string `json:"description"`
}

type CloudDBUsers struct {
	ID int `json:"id"`
	IDSettings int `json:"id_settings"`
	Phone string `json:"phone"`
	Pin int `json:"pin"`
}

type CloudDBSettings struct {
	ID int `json:"id,omitempty"`
	Alias string `json:"alias,omitempty"`
	SrvIP string `json:"srv_ip"`
	SrvUser string `json:"srv_user"`
	SrvPassword string  `json:"srv_password"`
}