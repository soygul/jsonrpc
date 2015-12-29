package jsonrpc

import (
	"sync"

	"github.com/neptulon/cmap"
	"github.com/neptulon/neptulon/client"
)

// Client is a Neptulon JSON-RPC client.
type Client struct {
	Middleware
	Sender
	client *client.Client // Inner Neptulon client.
}

// NewClient creates a new Client object.
// msgWG = (optional) sets the given *sync.WaitGroup reference to be used for counting active gorotuines that are used for handling incoming/outgoing messages.
// disconnHandler = (optional) registers a function to handle client disconnection events.
func NewClient(msgWG *sync.WaitGroup, disconnHandler func(client *client.Client)) *Client {
	return UseClient(client.NewClient(msgWG, disconnHandler))
}

// UseClient wraps an established Neptulon Client into a JSON-RPC Client.
func UseClient(client *client.Client) *Client {
	c := Client{client: client}
	c.client.MiddlewareIn(c.Middleware.neptulonMiddleware)
	c.Sender = NewSender(func(connID string, msg []byte) error { return c.client.Send(msg) })
	return &c
}

// ConnID is a randomly generated unique client connection ID.
func (c *Client) ConnID() string {
	return c.client.ConnID()
}

// Session is a thread-safe data store for storing arbitrary data for this connection session.
func (c *Client) Session() *cmap.CMap {
	return c.client.Session()
}

// SetDeadline set the read/write deadlines for the connection, in seconds.
func (c *Client) SetDeadline(seconds int) {
	c.client.SetDeadline(seconds)
}

// UseTLS enables Transport Layer Security for the connection.
// ca = Optional CA certificate to be used for verifying the server certificate. Useful for using self-signed server certificates.
// clientCert, clientCertKey = Optional certificate/privat key pair for TLS client certificate authentication.
// All certificates/private keys are in PEM encoded X.509 format.
func (c *Client) UseTLS(ca, clientCert, clientCertKey []byte) {
	c.client.UseTLS(ca, clientCert, clientCertKey)
}

// Connect connectes to the server at given network address and starts receiving messages.
func (c *Client) Connect(addr string, debug bool) error {
	return c.client.Connect(addr, debug)
}

// Close closes a client connection.
func (c *Client) Close() error {
	return c.client.Close()
}
