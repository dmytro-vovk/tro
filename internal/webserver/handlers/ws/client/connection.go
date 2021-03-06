package client

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"

	"github.com/dmytro-vovk/tro/internal/jsonrpc"
	"github.com/gorilla/websocket"
)

type connection struct {
	conn          *websocket.Conn
	methods       map[string]rpcHandler
	subscriptions map[string]struct{}
	sendC         chan interface{}
	doneC         chan struct{}
	mutex         sync.RWMutex
}

func NewConnection(conn *websocket.Conn, methods map[string]rpcHandler) *connection {
	return &connection{
		conn:          conn,
		methods:       methods,
		subscriptions: map[string]struct{}{},
		sendC:         make(chan interface{}, 1),
		doneC:         make(chan struct{}),
	}
}

func (c *connection) Run() {
	go c.receiver()
	go c.sender()

	<-c.doneC
}

func (c *connection) Notify(method string, params interface{}) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if _, ok := c.subscriptions[method]; !ok {
		return
	}

	payload, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}

	c.notify(jsonrpc.Request{
		Version: "2.0",
		Method:  method,
		Params:  payload,
	})
}

// todo: we can use return statement for try again send the request
func (c *connection) notify(notice jsonrpc.Request) bool {
	select {
	case c.sendC <- notice:
		//logrus.Printf("[%s] Sending message:\n%s", c.conn.RemoteAddr(), notice) // too noisy
		return true
	default:
		// try to change channel size
		logrus.Printf("[%s] Couldn't send notification:\n%s", c.conn.RemoteAddr(), notice)
		return false
	}
}

func (c *connection) receiver() {
	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Printf("[%s] Unexpected close error: %v", c.conn.RemoteAddr(), err)
			}

			close(c.doneC)
			return
		}

		switch msgType {
		case websocket.TextMessage:
			go c.handleTextMessage(msg)
		default:
			logrus.Printf("[%s] Unknown message type: %d", c.conn.RemoteAddr(), msgType)
		}
	}
}

func (c *connection) sender() {
	for {
		select {
		case resp := <-c.sendC:
			switch t := resp.(type) {
			case jsonrpc.Response:
			case jsonrpc.Request:
			default:
				logrus.Panicf("unknown response type: %T", t)
			}

			if err := c.conn.WriteJSON(resp); err != nil {
				logrus.Printf("[%s] Error sending message: %s", c.conn.RemoteAddr(), err)
			}
		case <-c.doneC:
			return
		}
	}
}

func (c *connection) handleTextMessage(msg []byte) {
	var req jsonrpc.Request
	if err := json.Unmarshal(msg, &req); err != nil {
		logrus.Printf("[%s] Error decoding request: %s", c.conn.RemoteAddr(), err)
		logrus.Printf("[%s] Request: %s", c.conn.RemoteAddr(), msg)
		c.sendC <- req.ErrorResponse(err)
		return
	}

	if err := req.Valid(); err != nil {
		logrus.Printf("[%s] Invalid request object: %s", c.conn.RemoteAddr(), err)
		c.sendC <- req.ErrorResponse(err)
		return
	}

	c.handleRequest(req)
}

func (c *connection) handleRequest(req jsonrpc.Request) {
	if req.IsNotification() {
		c.handleNotification(req)
		return
	}

	if fn, ok := c.methods[req.Method]; ok {
		data, err := fn.call(req.Params)
		if err != nil {
			logrus.Printf("[%s] RPC call %s(%s) error: %s", c.conn.RemoteAddr(), req.Method, req.Params, err.Error())
			c.sendC <- req.ErrorResponse(err)
		} else {
			c.sendC <- req.Response(data)
		}
	} else {
		logrus.Printf("[%s] Requested method %q doesn't exist", c.conn.RemoteAddr(), req.Method)
		c.sendC <- req.ErrorResponse(fmt.Errorf("method %q doesn't exist", req.Method))
	}
}

func (c *connection) handleNotification(notice jsonrpc.Request) {
	var method string
	if err := json.Unmarshal(notice.Params, &method); err != nil {
		logrus.Printf("[%s] Error decoding method name: %s", c.conn.RemoteAddr(), err)
		logrus.Printf("[%s] Params: %s", c.conn.RemoteAddr(), notice.Params)
		return
	}

	switch notice.Method {
	case "subscribe":
		c.subscribe(method)
	case "unsubscribe":
		c.unsubscribe(method)
	}
}

func (c *connection) subscribe(method string) {
	logrus.Printf("[%s] Subscribing to %q", c.conn.RemoteAddr(), method)
	c.mutex.Lock()
	c.subscriptions[method] = struct{}{}
	c.mutex.Unlock()
}

func (c *connection) unsubscribe(method string) {
	logrus.Printf("[%s] Unsubscribing from %q", c.conn.RemoteAddr(), method)
	c.mutex.Lock()
	delete(c.subscriptions, method)
	c.mutex.Unlock()
}
