//go:build tools

// Package apicheck records the API comparison tool and its inputs.
package apicheck

import (
	_ "github.com/Chriscbr/purecdk8s/constructs/v10"
	_ "github.com/aws/constructs-go/constructs/v10"
	_ "golang.org/x/exp/cmd/apidiff"
)
