// see https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// and https://github.com/golang/go/issues/25922

// +build tools

package tools

import (
	_ "github.com/cloudflare/cfssl/cmd/cfssl"
	_ "github.com/cloudflare/cfssl/cmd/cfssljson"
)
