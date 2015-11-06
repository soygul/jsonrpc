package jsonrpc

import (
	"encoding/json"
	"errors"

	"github.com/neptulon/neptulon"
)

// Client is a client implementation for JSON-RPC 2.0 protocol for Neptulon framework.
// Client implementations in other programming languages might be provided in separate repositories so check the documentation.
type Client struct {
	conn *neptulon.Conn
}

// Dial creates a new client connection to a given network address with optional CA and/or a client certificate (PEM encoded X.509 cert/key).
// Debug mode logs all raw TCP communication.
func Dial(addr string, ca []byte, clientCert []byte, clientCertKey []byte, debug bool) (*Client, error) {
	c, err := neptulon.Dial(addr, ca, clientCert, clientCertKey, debug)
	if err != nil {
		return nil, err
	}

	return &Client{conn: c}, nil
}

// SetReadDeadline set the read deadline for the connection in seconds.
func (c *Client) SetReadDeadline(seconds int) {
	c.conn.SetReadDeadline(seconds)
}

// ReadMsg reads a message off of a client connection and returns a request, response, or notification message depending on what server sent.
// Optionally, you can pass in a data structure that the returned JSON-RPC response result data will be serialized into (same for request params).
// Otherwise json.Unmarshal defaults apply.
// This function blocks until a message is read from the connection or connection timeout occurs.
func (c *Client) ReadMsg(resultData interface{}, paramsData interface{}) (req *Request, res *Response, not *Notification, err error) {
	var data []byte
	if data, err = c.conn.Read(); err != nil {
		return
	}

	msg := message{}
	if err = json.Unmarshal(data, &msg); err != nil {
		return
	}

	// if incoming message is a request or response
	if msg.ID != "" {
		// if incoming message is a request
		if msg.Method != "" {
			var p interface{}
			if paramsData != nil {
				p = paramsData
			}

			if err = json.Unmarshal(msg.Params, &p); err != nil {
				return
			}

			req = &Request{ID: msg.ID, Method: msg.Method, Params: p}
			return
		}

		// if incoming message is a response
		var r interface{}
		if resultData != nil {
			r = resultData
		}

		if msg.Result != nil {
			if err = json.Unmarshal(msg.Result, &r); err != nil {
				return
			}
		}

		res = &Response{ID: msg.ID, Result: r}
		if msg.Error != nil {
			res.Error = &ResError{Code: msg.Error.Code, Message: msg.Error.Message}
			if err = json.Unmarshal(msg.Error.Data, &res.Error.Data); err != nil {
				return
			}
		}

		return
	}

	// if incoming message is a notification
	if msg.Method != "" {
		not = &Notification{Method: msg.Method, Params: msg.Params}
	}

	err = errors.New("Received a malformed message.")
	return
}

// WriteRequest writes a JSON-RPC request message to a client connection with structured params object and auto generated request ID.
func (c *Client) WriteRequest(method string, params interface{}) (reqID string, err error) {
	id, err := neptulon.GenID()
	if err != nil {
		return "", err
	}

	return id, c.WriteMsg(Request{ID: id, Method: method, Params: params})
}

// WriteRequestArr writes a JSON-RPC request message to a client connection with array params and auto generated request ID.
func (c *Client) WriteRequestArr(method string, params ...interface{}) (reqID string, err error) {
	return c.WriteRequest(method, params)
}

// WriteNotification writes a JSON-RPC notification message to a client connection with structured params object.
func (c *Client) WriteNotification(method string, params interface{}) error {
	return c.WriteMsg(Notification{Method: method, Params: params})
}

// WriteNotificationArr writes a JSON-RPC notification message to a client connection with array params.
func (c *Client) WriteNotificationArr(method string, params ...interface{}) error {
	return c.WriteNotification(method, params)
}

// WriteResponse writes a JSON-RPC response message to a client connection.
func (c *Client) WriteResponse(id string, result interface{}, err *ResError) error {
	return c.WriteMsg(Response{ID: id, Result: result, Error: err})
}

// WriteMsg writes any JSON-RPC message to a client connection.
func (c *Client) WriteMsg(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := c.conn.Write(data); err != nil {
		return err
	}

	return nil
}

// Close closes a client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}