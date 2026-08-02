//go:build tools

// Package apicheck records the cdk8s API comparison tool and its inputs.
package apicheck

import (
	_ "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	_ "github.com/Chriscbr/purecdk8s/tools/api-compat-check"
	_ "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)
