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
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
	KafkaCFG `yaml:"kafka_cfg" env-required:"true"`
	RedisCFG  `yaml:"redis_cfg" env-required:"true"`
	EnrichmentURLS []string `yaml:"enrichment_URLs" env-required:"true"`
	EnrichmentTimeoutMS string `yaml:"enrichment_timeout_ms" env-default:"1"`
}

type HTTPServer struct {
	Address     string    `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration    `yaml:"idle_timeout" env-default:"60s"`
}

type KafkaCFG struct{
	KafkaURL string `yaml:"kafka_URL" env-default:"localhost:9093"`
	KafkaProducerTopic string `yaml:"kafka_producer_topic" env-default:"FIOfailed"`
	KafkaConsumerTopic string `yaml:"kafka_consumer_topic" env-default:"FIO"`
	KafkaDelayMS string `yaml:"kafka_delay_ms" env-default:"5"`
	KafkaConsumerGroup  string  `yaml:"kafka_consumer_groupID" env-default:"0"`
}

type RedisCFG struct{
	RedisURL string  `yaml:"redis_URL" env-default:"localhost:6379"`
	TTL  time.Duration `yaml:"ttl" env-default:"1s"`
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