package main

import (
	tp "github.com/jslyzt/teleport"
)

func main() {
	cli := tp.NewPeer(
		tp.PeerConfig{
			PrintDetail: false,
		},
		new(earlyCall),
	)
	defer cli.Close()
	_, err := cli.Dial(":9090")
	if err != nil {
		tp.Fatalf("%v", err)
	}
}

type earlyCall struct{}

func (e *earlyCall) Name() string {
	return "early_call"
}

func (e *earlyCall) PostDial(sess tp.PreSession) *tp.Rerror {
	rerr := sess.Send(
		"/early/ping",
		map[string]string{
			"author": "henrylee2cn",
		},
		nil,
	)
	if rerr != nil {
		return rerr
	}

	input, rerr := sess.Receive(func(header Header) interface{} {
		if header.URI() == "/early/pong" {
			return new(string)
		}
		tp.Panicf("Received an unexpected response: %s", header.URI())
		return nil
	})
	if rerr != nil {
		return rerr
	}
	tp.Infof("result: %v", input.String())
	return nil
}
