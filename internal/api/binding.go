package api

// Binding links a source exchange to a destination (queue or exchange)
// on a vhost, routed by RoutingKey.
type Binding struct {
	Source          string
	Vhost           string
	Destination     string
	DestinationType string
	RoutingKey      string
}
