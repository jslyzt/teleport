package websocket

import (
	"net"
	"path"
	"strings"

	tp "github.com/jslyzt/teleport"
	ws "github.com/jslyzt/teleport/mixer/websocket/websocket"
)

// NewDialPlugin creates a websocket plugin for client.
func NewDialPlugin(pattern string) tp.Plugin {
	pattern = path.Join("/", strings.TrimRight(pattern, "/"))
	if pattern == "/" {
		pattern = ""
	}
	return &clientPlugin{pattern}
}

type clientPlugin struct {
	pattern string
}

var (
	_ tp.PostDialPlugin = new(clientPlugin)
)

func (*clientPlugin) Name() string {
	return "websocket"
}

func (c *clientPlugin) PostDial(sess tp.PreSession) *tp.Rerror {
	var location, origin string
	if sess.Peer().TLSConfig() == nil {
		location = "ws://" + sess.RemoteAddr().String() + c.pattern
		origin = "ws://" + sess.LocalAddr().String() + c.pattern
	} else {
		location = "wss://" + sess.RemoteAddr().String() + c.pattern
		origin = "wss://" + sess.LocalAddr().String() + c.pattern
	}
	cfg, err := ws.NewConfig(location, origin)
	if err != nil {
		return tp.NewRerror(tp.CodeDialFailed, "upgrade to websocket failed", err.Error())
	}
	var rerr *tp.Rerror
	sess.ModifySocket(func(conn net.Conn) (net.Conn, tp.ProtoFunc) {
		conn, err := ws.NewClient(cfg, conn)
		if err != nil {
			rerr = tp.NewRerror(tp.CodeDialFailed, "upgrade to websocket failed", err.Error())
			return nil, nil
		}
		return conn, NewWsProtoFunc(sess.GetProtoFunc())
	})
	return rerr
}
