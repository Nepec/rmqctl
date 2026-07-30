package cli

import (
	"github.com/nepec/rmqctl/internal/api"
	"github.com/spf13/viper"
)

// ClientFromConfig builds a RabbitHoleClient from viper's current
// "hostname", "port", "username", and "password" values, as populated
// from flags, environment variables, and config file by the root command.
func ClientFromConfig() (*api.RabbitHoleClient, error) {
	host := viper.GetString("hostname")
	port := viper.GetInt("port")
	username := viper.GetString("username")
	password := viper.GetString("password")

	client, err := api.NewRabbitHoleClient(host, port, username, password)
	if err != nil {
		return nil, err
	}

	return client, nil
}
