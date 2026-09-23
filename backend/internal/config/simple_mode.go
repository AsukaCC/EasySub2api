package config

type SimpleModeConfig struct {
	AutoCreateDefaultGroups bool `mapstructure:"auto_create_default_groups" yaml:"auto_create_default_groups"`
}
