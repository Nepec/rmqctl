package api

type Binding struct {
	Source          string
	Vhost           string
	Destination     string
	DestinationType string
	RoutingKey      string
}
