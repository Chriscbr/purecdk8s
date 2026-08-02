//go:build tools

// Package apicheck records the constructs API comparison inputs.
package apicheck

import (
	_ "github.com/Chriscbr/purecdk8s/constructs/v10"
	_ "github.com/Chriscbr/purecdk8s/tools/api-compat-check"
	_ "github.com/aws/constructs-go/constructs/v10"
)
