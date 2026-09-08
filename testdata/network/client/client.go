package client

// Client communicates over the network; it has no field pointing to the server.
// architecture: Client
type Client struct{}

// architecture:link Client -> example/network/server.Server : sends requests
// architecture:link Client -> external:payments-api : HTTPS
