package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/henrylee2cn/goutil"
	tp "github.com/jslyzt/teleport"
)

// A proxy plugin for handling unknown calling or pushing.

// Proxy creates a proxy plugin for handling unknown calling and pushing.
func Proxy(fn func(*Label) Forwarder) tp.Plugin {
	return &proxy{
		callForwarder: func(label *Label) CallForwarder {
			return fn(label)
		},
		pushForwarder: func(label *Label) PushForwarder {
			return fn(label)
		},
	}
}

// Call creates a proxy plugin for handling unknown calling.
func Call(fn func(*Label) CallForwarder) tp.Plugin {
	return &proxy{callForwarder: fn}
}

// Push creates a proxy plugin for handling unknown pushing.
func Push(fn func(*Label) PushForwarder) tp.Plugin {
	return &proxy{pushForwarder: fn}
}

type (
	// Forwarder the object used to call and push
	Forwarder interface {
		CallForwarder
		PushForwarder
	}
	// CallForwarder the object used to call
	CallForwarder interface {
		Call(uri string, arg interface{}, result interface{}, setting ...tp.MessageSetting) tp.CallCmd
	}
	// PushForwarder the object used to push
	PushForwarder interface {
		Push(uri string, arg interface{}, setting ...tp.MessageSetting) *tp.Rerror
	}
	// Label proxy label information
	Label struct {
		SessionID, RealIP, URI string
	}
	proxy struct {
		callForwarder func(*Label) CallForwarder
		pushForwarder func(*Label) PushForwarder
	}
)

var (
	_ tp.PostNewPeerPlugin = new(proxy)
)

func (p *proxy) Name() string {
	return "proxy"
}

func (p *proxy) PostNewPeer(peer tp.EarlyPeer) error {
	if p.callForwarder != nil {
		peer.SetUnknownCall(p.call)
	}
	if p.pushForwarder != nil {
		peer.SetUnknownPush(p.push)
	}
	return nil
}

func (p *proxy) call(ctx tp.UnknownCallCtx) (interface{}, *tp.Rerror) {
	var (
		label    Label
		settings = make([]tp.MessageSetting, 1, 8)
	)
	label.SessionID = ctx.Session().ID()
	settings[0] = tp.WithSeq(getSeq(label.SessionID + "@" + ctx.Seq()))
	ctx.VisitMeta(func(key, value []byte) {
		settings = append(settings, tp.WithAddMeta(string(key), string(value)))
	})
	var (
		result      []byte
		realIPBytes = ctx.PeekMeta(tp.MetaRealIP)
	)
	if len(realIPBytes) == 0 {
		label.RealIP = ctx.IP()
		settings = append(settings, tp.WithAddMeta(tp.MetaRealIP, label.RealIP))
	} else {
		label.RealIP = goutil.BytesToString(realIPBytes)
	}
	label.URI = ctx.URI()
	callcmd := p.callForwarder(&label).Call(label.URI, ctx.InputBodyBytes(), &result, settings...)
	callcmd.InputMeta().VisitAll(func(key, value []byte) {
		ctx.SetMeta(goutil.BytesToString(key), goutil.BytesToString(value))
	})
	rerr := callcmd.Rerror()
	if rerr != nil && rerr.Code < 200 && rerr.Code > 99 {
		rerr.Code = tp.CodeBadGateway
		rerr.Message = tp.CodeText(tp.CodeBadGateway)
	}
	return result, rerr
}

func (p *proxy) push(ctx tp.UnknownPushCtx) *tp.Rerror {
	var (
		label    Label
		settings = make([]tp.MessageSetting, 1, 8)
	)
	label.SessionID = ctx.Session().ID()
	settings[0] = tp.WithSeq(getSeq(label.SessionID + "@" + ctx.Seq()))
	ctx.VisitMeta(func(key, value []byte) {
		settings = append(settings, tp.WithAddMeta(string(key), string(value)))
	})
	if realIPBytes := ctx.PeekMeta(tp.MetaRealIP); len(realIPBytes) == 0 {
		label.RealIP = ctx.IP()
		settings = append(settings, tp.WithAddMeta(tp.MetaRealIP, label.RealIP))
	} else {
		label.RealIP = goutil.BytesToString(realIPBytes)
	}
	label.URI = ctx.URI()
	rerr := p.pushForwarder(&label).Push(label.URI, ctx.InputBodyBytes(), settings...)
	if rerr != nil && rerr.Code < 200 && rerr.Code > 99 {
		rerr.Code = tp.CodeBadGateway
		rerr.Message = tp.CodeText(tp.CodeBadGateway)
	}
	return rerr
}

var peerName = filepath.Base(os.Args[0])
var incr int64
var mutex sync.Mutex

// getSeq creates a new sequence with some prefix string.
func getSeq(prefix ...string) string {
	mutex.Lock()
	seq := fmt.Sprintf("%s[%d]", peerName, incr)
	incr++
	mutex.Unlock()
	for _, p := range prefix {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		seq = p + ">" + seq
	}
	return seq
}
