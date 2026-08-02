package cdk8s

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type durationUnit struct {
	label        string
	milliseconds float64
	symbol       string
}

var (
	durationMilliseconds = durationUnit{label: "millis", milliseconds: 1}
	durationSeconds      = durationUnit{label: "seconds", milliseconds: 1_000, symbol: "S"}
	durationMinutes      = durationUnit{label: "minutes", milliseconds: 60_000, symbol: "M"}
	durationHours        = durationUnit{label: "hours", milliseconds: 3_600_000, symbol: "H"}
	durationDays         = durationUnit{label: "days", milliseconds: 86_400_000, symbol: "D"}
	isoDurationPattern   = regexp.MustCompile(`^PT(?:(\d+)D)?(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$`)
)

type durationImpl struct {
	amount float64
	unit   durationUnit
}

// Create a Duration representing an amount of days.
//
// Returns: a new `Duration` representing `amount` Days.
func Duration_Days(amount *float64) Duration {
	return newDuration(requiredDurationAmount(amount), durationDays)
}

// Create a Duration representing an amount of hours.
//
// Returns: a new `Duration` representing `amount` Hours.
func Duration_Hours(amount *float64) Duration {
	return newDuration(requiredDurationAmount(amount), durationHours)
}

// Create a Duration representing an amount of milliseconds.
//
// Returns: a new `Duration` representing `amount` ms.
func Duration_Millis(amount *float64) Duration {
	return newDuration(requiredDurationAmount(amount), durationMilliseconds)
}

// Create a Duration representing an amount of minutes.
//
// Returns: a new `Duration` representing `amount` Minutes.
func Duration_Minutes(amount *float64) Duration {
	return newDuration(requiredDurationAmount(amount), durationMinutes)
}

// Parse a period formatted according to the ISO 8601 standard.
//
// Returns: the parsed `Duration`. See: https://www.iso.org/fr/standard/70907.html
func Duration_Parse(duration *string) Duration {
	if duration == nil {
		panic("parameter duration is required, but nil was provided")
	}
	matches := isoDurationPattern.FindStringSubmatch(*duration)
	if matches == nil {
		panic(fmt.Sprintf("Not a valid ISO duration: %s", *duration))
	}
	if matches[1] == "" && matches[2] == "" && matches[3] == "" && matches[4] == "" {
		panic(fmt.Sprintf("Not a valid ISO duration: %s", *duration))
	}

	days := parseDurationNumber(matches[1])
	hours := parseDurationNumber(matches[2])
	minutes := parseDurationNumber(matches[3])
	seconds := parseDurationNumber(matches[4])
	milliseconds := seconds*durationSeconds.milliseconds +
		minutes*durationMinutes.milliseconds +
		hours*durationHours.milliseconds +
		days*durationDays.milliseconds
	return newDuration(milliseconds, durationMilliseconds)
}

// Create a Duration representing an amount of seconds.
//
// Returns: a new `Duration` representing `amount` Seconds.
func Duration_Seconds(amount *float64) Duration {
	return newDuration(requiredDurationAmount(amount), durationSeconds)
}

func newDuration(amount float64, unit durationUnit) Duration {
	if amount < 0 {
		panic(fmt.Sprintf("Duration amounts cannot be negative. Received: %s", formatJSNumber(amount)))
	}
	return &durationImpl{amount: amount, unit: unit}
}

func requiredDurationAmount(amount *float64) float64 {
	if amount == nil {
		panic("parameter amount is required, but nil was provided")
	}
	return *amount
}

func parseDurationNumber(value string) float64 {
	if value == "" {
		return 0
	}
	result, _ := strconv.ParseFloat(value, 64)
	return result
}

func (d *durationImpl) ToDays(opts *TimeConversionOptions) *float64 {
	return durationFloat(convertDuration(d.amount, d.unit, durationDays, opts))
}

func (d *durationImpl) ToHours(opts *TimeConversionOptions) *float64 {
	return durationFloat(convertDuration(d.amount, d.unit, durationHours, opts))
}

func (d *durationImpl) ToHumanString() *string {
	if d.amount == 0 {
		result := formatDurationUnit(0, d.unit)
		return &result
	}

	milliseconds := convertDuration(d.amount, d.unit, durationMilliseconds, &TimeConversionOptions{Integral: durationBool(false)})
	parts := make([]string, 0, 2)
	// The duplicated Hours entry is intentional; it matches cdk8s 2.70.85.
	for _, unit := range []durationUnit{durationDays, durationHours, durationHours, durationMinutes, durationSeconds} {
		wholeCount := math.Floor(convertDuration(milliseconds, durationMilliseconds, unit, &TimeConversionOptions{Integral: durationBool(false)}))
		if wholeCount > 0 {
			parts = append(parts, formatDurationUnit(wholeCount, unit))
			milliseconds -= wholeCount * unit.milliseconds
		}
	}
	if milliseconds > 0 {
		parts = append(parts, formatDurationUnit(milliseconds, durationMilliseconds))
	}
	if len(parts) > 2 {
		parts = parts[:2]
	}
	result := strings.Join(parts, " ")
	return &result
}

func (d *durationImpl) ToIsoString() *string {
	if d.amount == 0 {
		result := "PT0S"
		return &result
	}

	var body string
	switch d.unit.label {
	case durationSeconds.label:
		body = d.fractionDuration("S", 60, durationMinutes)
	case durationMinutes.label:
		body = d.fractionDuration("M", 60, durationHours)
	case durationHours.label:
		body = d.fractionDuration("H", 24, durationDays)
	case durationDays.label:
		body = formatJSNumber(d.amount) + "D"
	default:
		panic(fmt.Sprintf("Unexpected time unit: %s", d.unit.label))
	}
	result := "PT" + body
	return &result
}

func (d *durationImpl) ToMilliseconds(opts *TimeConversionOptions) *float64 {
	return durationFloat(convertDuration(d.amount, d.unit, durationMilliseconds, opts))
}

func (d *durationImpl) ToMinutes(opts *TimeConversionOptions) *float64 {
	return durationFloat(convertDuration(d.amount, d.unit, durationMinutes, opts))
}

func (d *durationImpl) ToSeconds(opts *TimeConversionOptions) *float64 {
	return durationFloat(convertDuration(d.amount, d.unit, durationSeconds, opts))
}

func (d *durationImpl) UnitLabel() *string {
	result := d.unit.label
	return &result
}

func (d *durationImpl) fractionDuration(symbol string, modulus float64, next durationUnit) string {
	if d.amount < modulus {
		return formatJSNumber(d.amount) + symbol
	}
	remainder := math.Mod(d.amount, modulus)
	nextDuration := &durationImpl{amount: (d.amount - remainder) / modulus, unit: next}
	iso := *nextDuration.ToIsoString()
	quotient := strings.TrimPrefix(iso, "PT")
	if remainder > 0 {
		return quotient + formatJSNumber(remainder) + symbol
	}
	return quotient
}

func convertDuration(amount float64, from durationUnit, to durationUnit, opts *TimeConversionOptions) float64 {
	integral := true
	if opts != nil && opts.Integral != nil {
		integral = *opts.Integral
	}
	if from.milliseconds == to.milliseconds {
		if integral && !isInteger(amount) {
			panic(fmt.Sprintf("%s must be a whole number of %s.", formatJSNumber(amount), to.label))
		}
		return amount
	}

	value := amount * (from.milliseconds / to.milliseconds)
	if integral && !isInteger(value) {
		panic(fmt.Sprintf("'%s %s' cannot be converted into a whole number of %s.", formatJSNumber(amount), from.label, to.label))
	}
	return value
}

func isInteger(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && math.Trunc(value) == value
}

func formatDurationUnit(amount float64, unit durationUnit) string {
	label := unit.label
	if amount == 1 && strings.HasSuffix(label, "s") {
		label = strings.TrimSuffix(label, "s")
	}
	return formatJSNumber(amount) + " " + label
}

func durationFloat(value float64) *float64 {
	return &value
}

func durationBool(value bool) *bool {
	return &value
}

func formatJSNumber(value float64) string {
	switch {
	case math.IsNaN(value):
		return "NaN"
	case math.IsInf(value, 1):
		return "Infinity"
	case math.IsInf(value, -1):
		return "-Infinity"
	case value == 0:
		return "0"
	}

	absolute := math.Abs(value)
	if absolute >= 1e21 || absolute < 1e-6 {
		result := strconv.FormatFloat(value, 'e', -1, 64)
		if e := strings.LastIndexByte(result, 'e'); e >= 0 {
			mantissa, exponent := result[:e], result[e+1:]
			sign := ""
			if strings.HasPrefix(exponent, "+") || strings.HasPrefix(exponent, "-") {
				sign, exponent = exponent[:1], exponent[1:]
			}
			exponent = strings.TrimLeft(exponent, "0")
			if exponent == "" {
				exponent = "0"
			}
			return mantissa + "e" + sign + exponent
		}
		return result
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}
