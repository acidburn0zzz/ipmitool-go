package ipmitool

import (
	"os"
)

type Config struct {
	Host string
	User string
	Pass string
}

func ConfigFromEnvironment() Config {
	return Config{
		Host: os.Getenv("IPMI_HOST"),
		User: os.Getenv("IPMI_USER"),
		Pass: os.Getenv("IPMI_PASS"),
	}
}

func (c *Config) args(cmd ...string) []string {
	args := append(make([]string, 0, 12+len(cmd)), "-I", "lanplus", "-R", "1", "-N", "1", "-H", c.Host, "-U", c.User, "-P", c.Pass)

	return append(args, cmd...)
}
