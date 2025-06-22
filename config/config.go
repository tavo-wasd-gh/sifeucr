package config

import (
	"os"
	"reflect"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type Env struct {
	Production  bool   `env:"PRODUCTION"`
	Debug       bool   `env:"DEBUG"`
	Port        string `env:"PORT" req:"1"`
	Secret      string `env:"SECRET" req:"1"`
	AllowOrigin string `env:"API_ALLOW_ORIGIN"`
	DBConnDvr   string `env:"DB_CONNDVR"`
	DBConnStr   string `env:"DB_CONNSTR"`
}

func Init() (*Env, error) {
	godotenv.Load()

	var missing []string
	cfg := &Env{}

	t := reflect.TypeOf(*cfg)
	v := reflect.ValueOf(cfg).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envVar := field.Tag.Get("env")
		req := field.Tag.Get("req")

		if envVar == "" {
			continue
		}

		value := os.Getenv(envVar)

		if req == "1" && value == "" {
			missing = append(missing, envVar)
			continue
		}

		switch field.Type.Kind() {
		case reflect.Bool:
			boolVal, _ := strconv.ParseBool(value)
			v.Field(i).SetBool(boolVal)
		case reflect.Int:
			intVal, _ := strconv.Atoi(value)
			v.Field(i).SetInt(int64(intVal))
		case reflect.String:
			v.Field(i).SetString(value)
		}
	}

	if len(missing) > 0 {
		return nil, logger.Errorf("missing environment variables: %s", missing)
	}

	return cfg, nil
}
