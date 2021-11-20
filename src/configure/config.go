package configure

import (
	"bytes"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func checkErr(err error) {
	if err != nil {
		logrus.WithError(err).Fatal("config")
	}
}

func New() *Config {
	config := viper.New()
	config.SetConfigType("yaml")

	b, err := json.Marshal(Config{
		LogLevel: "info",
		Config:   "config.yaml",
	})

	checkErr(err)
	tmp := viper.New()
	tmp.SetConfigType("json")
	checkErr(tmp.ReadConfig(bytes.NewBuffer(b)))
	checkErr(config.MergeConfigMap(tmp.AllSettings()))

	pflag.String("config", "config.yaml", "Config file location")
	pflag.Bool("noheader", false, "Disable the startup header")
	pflag.Parse()
	checkErr(config.BindPFlags(pflag.CommandLine))

	config.SetConfigFile(config.GetString("config"))
	if err := config.ReadInConfig(); err == nil {
		checkErr(config.MergeInConfig())
	}

	cfg := Config{}

	config.SetEnvPrefix("7TV")
	config.AllowEmptyEnv(true)
	config.AutomaticEnv()

	checkErr(config.Unmarshal(&cfg))

	initLogging(cfg.LogLevel)

	return &cfg
}

type Config struct {
	LogLevel string `json:"log_level,omitempty" mapstructure:"log_level,omitempty"`
	Config   string `json:"config,omitempty" mapstructure:"config,omitempty"`
	NoHeader bool   `json:"noheader,omitempty" mapstructure:"noheader,omitempty"`
	NoLogs   bool   `json:"nologs,omitempty" mapstructure:"nologs,omitempty"`

	// Aws
	Aws struct {
		AccessToken string `json:"access_token,omitempty" mapstructure:"access_token,omitempty"`
		SecretKey   string `json:"secret_key,omitempty" mapstructure:"secret_key,omitempty"`
		Region      string `json:"region,omitempty" mapstructure:"region,omitempty"`
	} `json:"aws,omitempty" mapstructure:"aws,omitempty"`

	Rmq struct {
		ServerURL       string `json:"server_url,omitempty" mapstructure:"server_url,omitempty"`
		JobQueueName    string `json:"job_queue_name,omitempty" mapstructure:"job_queue_name,omitempty"`
		ResultQueueName string `json:"result_queue_name,omitempty" mapstructure:"result_queue_name,omitempty"`
		UpdateQueueName string `json:"update_queue_name,omitempty" mapstructure:"update_queue_name,omitempty"`
	} `json:"rmq,omitempty" mapstructure:"rmq,omitempty"`

	WorkingDir      string `json:"working_dir,omitempty" mapstructure:"working_dir,omitempty"`
	MaxTaskDuration int    `json:"max_task_duration,omitempty" mapstructure:"max_task_duration,omitempty"`
	Av1Decoder      string `json:"av1_decoder,omitempty" mapstructure:"av1_decoder,omitempty"`
	Av1Encoder      string `json:"av1_encoder,omitempty" mapstructure:"av1_encoder,omitempty"`
}
