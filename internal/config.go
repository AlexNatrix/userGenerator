package internal

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true"`
	StoragePath string `yaml:"storage_path" env-required:"true" `
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string    `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration    `yaml:"idle_timeout" env-default:"60s"`
}

func LoadConfig() (Config,error){
	const op = "config.LoadConfig"
	configPath:=os.Getenv("CONFIG_PATH")
	var cfg Config
	if configPath == ""{
		return cfg,errors.New("var CONFIG_PATH is not set")
	}

	if _,err := os.Stat(configPath);err!=nil{
		return cfg,fmt.Errorf("config file %s is not exist %s: %w",configPath,op,err)
	}

	if err:=cleanenv.ReadConfig(configPath,&cfg);err!=nil{
		return cfg,fmt.Errorf("cant read config file: %s %w",op,err)
	}
	return cfg, nil
}