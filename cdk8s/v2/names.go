package cdk8s

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/purecdk8s/purecdk8s/constructs/v10"
)

const (
	defaultNameMaxLength = 63
	nameHashLength       = 8
)

var (
	validDNSLabelComponent = regexp.MustCompile(`^[0-9a-z-]+$`)
	validLabelValue        = regexp.MustCompile(`^(([0-9a-zA-Z][0-9a-zA-Z-_.]*)?[0-9a-zA-Z])?$`)
	invalidDNSCharacters   = regexp.MustCompile(`[^0-9a-zA-Z-_.]`)
	invalidLabelDelimiter  = regexp.MustCompile(`[^0-9a-zA-Z-_.]`)
	invalidLabelStart      = regexp.MustCompile(`^[^0-9a-zA-Z]+`)
)

// Names_ToDnsLabel generates a stable RFC-1123 DNS label for scope.
func Names_ToDnsLabel(scope constructs.Construct, options *NameOptions) *string {
	if scope == nil {
		panic("parameter scope is required, but nil was provided")
	}

	maxLen, delimiter, includeHash, extra := nameOptions(options)
	if maxLen < nameHashLength && includeHash {
		panic(fmt.Sprintf("minimum max length for object names is %d (required for hash)", nameHashLength))
	}

	node := scope.Node()
	components := strings.Split(namesString(node.Path()), "/")
	components = append(components, extra...)
	if len(components) == 1 && validDNSLabelComponent.MatchString(components[0]) && len(components[0]) <= maxLen {
		result := components[0]
		return &result
	}

	for index, component := range components {
		components[index] = normalizeDNSName(component, maxLen)
	}
	if includeHash {
		components = append(components, calculateNameHash(node))
	}
	result := humanName(components, delimiter, maxLen)
	return &result
}

// Names_ToLabelValue generates a stable Kubernetes label value for scope.
func Names_ToLabelValue(scope constructs.Construct, options *NameOptions) *string {
	if scope == nil {
		panic("parameter scope is required, but nil was provided")
	}

	maxLen, delimiter, includeHash, extra := nameOptions(options)
	if maxLen < nameHashLength && includeHash {
		panic(fmt.Sprintf("minimum max length for label is %d (required for hash)", nameHashLength))
	}
	if invalidLabelDelimiter.MatchString(delimiter) {
		panic(`delim should not contain "[^0-9a-zA-Z-_.]"`)
	}

	node := scope.Node()
	components := strings.Split(namesString(node.Path()), "/")
	components = append(components, extra...)
	if len(components) == 1 && validLabelValue.MatchString(components[0]) && len(components[0]) <= maxLen {
		result := components[0]
		return &result
	}

	for index, component := range components {
		components[index] = normalizeLabelValue(component, maxLen)
	}
	if includeHash {
		components = append(components, calculateNameHash(node))
	}
	result := humanName(components, delimiter, maxLen)
	result = invalidLabelStart.ReplaceAllString(result, "")
	return &result
}

func nameOptions(options *NameOptions) (maxLen int, delimiter string, includeHash bool, extra []string) {
	maxLen = defaultNameMaxLength
	delimiter = "-"
	includeHash = true
	if options == nil {
		return
	}
	if options.MaxLen != nil {
		switch {
		case math.IsNaN(*options.MaxLen):
			maxLen = 0
		case *options.MaxLen > float64(math.MaxInt):
			maxLen = math.MaxInt
		case *options.MaxLen < float64(math.MinInt):
			maxLen = math.MinInt
		default:
			maxLen = int(*options.MaxLen)
		}
	}
	if options.Delimiter != nil {
		delimiter = *options.Delimiter
	}
	if options.IncludeHash != nil {
		includeHash = *options.IncludeHash
	}
	if options.Extra != nil {
		extra = make([]string, 0, len(*options.Extra))
		for _, component := range *options.Extra {
			extra = append(extra, namesString(component))
		}
	}
	return
}

func normalizeDNSName(component string, maxLen int) string {
	component = strings.ToLower(component)
	component = invalidDNSCharacters.ReplaceAllString(component, "")
	return javascriptSubstr(component, maxLen)
}

func normalizeLabelValue(component string, maxLen int) string {
	component = invalidDNSCharacters.ReplaceAllString(component, "")
	return javascriptSubstr(component, maxLen)
}

func javascriptSubstr(value string, length int) string {
	if length <= 0 {
		return ""
	}
	if len(value) <= length {
		return value
	}
	return value[:length]
}

func calculateNameHash(node constructs.Node) string {
	if os.Getenv("CDK8S_LEGACY_HASH") != "" {
		digest := sha256.Sum256([]byte(namesString(node.Path())))
		return hex.EncodeToString(digest[:])[:nameHashLength]
	}
	address := namesString(node.Addr())
	if len(address) <= nameHashLength {
		return address
	}
	return address[:nameHashLength]
}

func humanName(components []string, delimiter string, maxLen int) string {
	reversed := make([]string, len(components))
	for index := range components {
		reversed[index] = components[len(components)-1-index]
	}

	deduplicated := make([]string, 0, len(reversed))
	for _, component := range reversed {
		if len(deduplicated) == 0 || component != deduplicated[len(deduplicated)-1] {
			deduplicated = append(deduplicated, component)
		}
	}

	joined := strings.Join(deduplicated, "/")
	joined = javascriptSliceFromZero(joined, maxLen)
	parts := strings.Split(joined, "/")
	for left, right := 0, len(parts)-1; left < right; left, right = left+1, right-1 {
		parts[left], parts[right] = parts[right], parts[left]
	}

	nonempty := parts[:0]
	for _, part := range parts {
		if part != "" {
			nonempty = append(nonempty, part)
		}
	}
	joined = strings.Join(nonempty, delimiter)

	// JavaScript's split("") splits into UTF-16 code units. Normalized names are
	// ASCII, so splitting into bytes has the same result.
	var delimited []string
	if delimiter == "" {
		delimited = make([]string, 0, len(joined))
		for index := 0; index < len(joined); index++ {
			delimited = append(delimited, joined[index:index+1])
		}
	} else {
		delimited = strings.Split(joined, delimiter)
	}
	result := delimited[:0]
	for _, part := range delimited {
		if part == "" || strings.EqualFold(part, "resource") || strings.EqualFold(part, "default") {
			continue
		}
		result = append(result, part)
	}
	return strings.Join(result, delimiter)
}

func javascriptSliceFromZero(value string, end int) string {
	if end >= 0 {
		if end >= len(value) {
			return value
		}
		return value[:end]
	}
	end = len(value) + end
	if end <= 0 {
		return ""
	}
	return value[:end]
}

func namesString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
