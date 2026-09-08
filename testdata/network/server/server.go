package server

// Server sends events back to connected clients.
// architecture: Server
type Server struct{}

// architecture:link Server -> example/network/client.Client : pushes events
