package cli

import (
	"fmt"
	"log/slog"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/spf13/viper"
)

func ClientFromConfig() (*api.RabbitHoleAdapter, error) {
	host := viper.GetString("hostname")
	port := viper.GetInt("port")
	username := viper.GetString("username")
	password := viper.GetString("password")

	adapter, err := api.NewRabbitHoleAdapter(host, port, username, password)
	if err != nil {
		return nil, err
	}

	return adapter, nil
}

func ResolveVhosts(c api.RabbitClient, requested []string) ([]string, error) {
	if len(requested) == 1 && requested[0] == "*" {
		slog.Debug("wildcard '*' detected, fetching all vhosts from broker...")
		vs, err := c.ListVhosts()
		if err != nil {
			return nil, err
		}

		vhosts := make([]string, 0, len(vs))
		for _, v := range vs {
			vhosts = append(vhosts, v.Name)
		}

		return vhosts, nil
	}

	return requested, nil
}

func ValidateQueueType(t string) error {
	switch t {
	case "", "classic", "quorum":
	// ok
	default:
		return fmt.Errorf("queue type may wither by 'classic' or 'quorum', got %q", t)
	}

	return nil
}
