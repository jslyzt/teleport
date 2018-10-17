package main

import (
	tp "github.com/jslyzt/teleport"
)

func main() {
	srv := tp.NewPeer(
		tp.PeerConfig{
			PrintDetail: false,
			ListenPort:  9090,
		},
		new(earlyResult),
	)
	srv.ListenAndServe()
}

type earlyResult struct{}

func (e *earlyResult) Name() string {
	return "early_result"
}

func (e *earlyResult) PostAccept(sess tp.PreSession) *tp.Rerror {
	var rigthURI bool
	input, rerr := sess.Receive(func(header Header) interface{} {
		if header.URI() == "/early/ping" {
			rigthURI = true
			return new(map[string]string)
		}
		return nil
	})
	if rerr != nil {
		return rerr
	}

	var result string
	if !rigthURI {
		rerr = tp.NewRerror(10005, "unexpected request", "")
	} else {
		body := *input.Body().(*map[string]string)
		if body["author"] != "henrylee2cn" {
			rerr = tp.NewRerror(10005, "incorrect author", body["author"])
		} else {
			rerr = nil
			result = "OK"
		}
	}
	return sess.Send(
		"/early/pong",
		result,
		rerr,
	)
}
