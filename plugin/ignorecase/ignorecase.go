// Package ignorecase dynamically ignoring the case of path
package ignorecase

import (
	"strings"

	tp "github.com/jslyzt/teleport"
)

// NewIgnoreCase Returns a ignoreCase plugin.
func NewIgnoreCase() *ignoreCase {
	return &ignoreCase{}
}

type ignoreCase struct{}

var (
	_ tp.PostReadCallHeaderPlugin = new(ignoreCase)
	_ tp.PostReadPushHeaderPlugin = new(ignoreCase)
)

func (i *ignoreCase) Name() string {
	return "ignoreCase"
}

func (i *ignoreCase) PostReadCallHeader(ctx tp.ReadCtx) *tp.Rerror {
	// Dynamic transformation path is lowercase
	ctx.URIObject().Path = strings.ToLower(ctx.URIObject().Path)
	return nil
}

func (i *ignoreCase) PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror {
	// Dynamic transformation path is lowercase
	ctx.URIObject().Path = strings.ToLower(ctx.URIObject().Path)
	return nil
}
