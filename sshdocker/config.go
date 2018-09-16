package sshdocker

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Config
type Config struct {
	// ConfigFile
	ConfigFile string `yaml:"-"`
	// Addr is listen address by the web server.
	Addr string `yaml:"addr"`
	// Debug mode
	Debug bool `yaml:"debug"`
	// PublicKeyAuthentication
	PublicKeyAuthentication bool `yaml:"public_key_authentication"`
	// AuthorizedKeysFile
	AuthorizedKeysFile string `yaml:"authorized_keys_file"`
	// AuthorizedKeys
	AuthorizedKeys []string `yaml:"authorized_keys"`
	// HostKeyFile
	HostKeyFile string `yaml:"host_key_file"`
	// ContainerLabel
	ContainerLabel string `yaml:"container_label"`
	// Runtimes
	Runtimes map[string]*RuntimeConfig `yaml:"runtimes"`
}

// RuntimeConfig
type RuntimeConfig struct {
	// name
	Name string `yaml:"-"`
	// PublicKeyAuthentication
	PublicKeyAuthentication *bool `yaml:"public_key_authentication"`
	// AuthorizedKeysFile
	AuthorizedKeysFile *string `yaml:"authorized_keys_file"`
	// AuthorizedKeys
	AuthorizedKeys *[]string `yaml:"authorized_keys"`
	// Image
	Image string `yaml:"image"`
	// Options
	Options []string `yaml:"options"`
	// Command
	Command []string `yaml:"command"`

	Container *struct {
		// Image
		Image string `yaml:"image"`
		// Options
		Options []string `yaml:"options"`
		// Command
		Command []string `yaml:"command"`
	} `yaml:"container"`
}

func NewConfig() *Config {
	return &Config{
		ConfigFile:              "",
		Addr:                    ":2222",
		Debug:                   false,
		ContainerLabel:          "sshdocker",
		PublicKeyAuthentication: false,
		AuthorizedKeysFile:      "",
		AuthorizedKeys:          []string{},
		Runtimes:                map[string]*RuntimeConfig{},
	}
}

func (c *Config) LoadConfigFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, c)
	if err != nil {
		return err
	}

	for k, v := range c.Runtimes {
		v.Name = k
	}

	c.ConfigFile = path

	return nil
}

func (c *Config) Reload() error {
	return c.LoadConfigFile(c.ConfigFile)
}
