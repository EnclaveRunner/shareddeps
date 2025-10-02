package shareddeps

import "reflect"

type BaseConfig struct {
	HumanReadableOutput bool   `mapstructure:"human_readable_output" validate:"required"`
	LogLevel            string `mapstructure:"log_level"             validate:"required,oneof=debug info warn error"`
	Port                string `mapstructure:"port"                  validate:"required,port"`
}

func LoadAppConfig(configType reflect.Type) {
	
}
