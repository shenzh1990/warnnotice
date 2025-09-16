package settings

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

var config = flag.String("f", "./pkg/settings/config.yaml", "config file")

type Config struct {
	Db        DbConfig        `yaml:"Db"`
	Redis     RedisConfig     `yaml:"Redis"`
	Base      BaseConfig      `yaml:"Base"`
	LogCenter LogCenterConfig `yaml:"LogCenter"`
}
type BaseConfig struct {
	RunMode          string        `yaml:"RunMode"`
	HTTPPort         int           `yaml:"HTTPPort"`
	ReadTimeout      time.Duration `yaml:"ReadTimeout"`
	WriteTimeout     time.Duration `yaml:"WriteTimeout"`
	PageSize         int           `yaml:"PageSize"`
	JwtSecret        string        `yaml:"JwtSecret"`
	WsDiscardTimeout time.Duration `yaml:"WsDiscardTimeout"`
}
type DbConfig struct {
	DriverName string `yaml:"DriverName"`
	DBUrl      string `yaml:"DBUrl"`
}
type LogCenterConfig struct {
	Url    string `yaml:"Url"`
	System string `yaml:"System"`
}

type RedisConfig struct {
	RedisHost       string `yaml:"RedisHost"`
	RedisDB         string `yaml:"RedisDB"`
	RedisPwd        string `yaml:"RedisPwd"`
	Timeout         int64  `yaml:"Timeout"`
	PoolMaxIdle     int    `yaml:"PoolMaxIdle"`
	PoolMaxActive   int    `yaml:"PoolMaxActive"`
	PoolIdleTimeout int64  `yaml:"PoolIdleTimeout"`
	PoolWait        bool   `yaml:"PoolWait"`
}

func (c *Config) getConf(filepath string) *Config {
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		fmt.Println(err.Error())
	}
	return c
}
func init() {
	flag.Parse()
	InitConfig.getConf(*config)
	LoadBase()
}

var (
	InitConfig Config
	RunMode    string

	HTTPPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PageSize  int
	JwtSecret string
)

func LoadBase() {
	RunMode = InitConfig.Base.RunMode
	HTTPPort = InitConfig.Base.HTTPPort
	ReadTimeout = InitConfig.Base.ReadTimeout * time.Second
	WriteTimeout = InitConfig.Base.WriteTimeout * time.Second
	JwtSecret = InitConfig.Base.JwtSecret
	PageSize = InitConfig.Base.PageSize
}
