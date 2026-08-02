package cdk8s_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

func TestDuration(t *testing.T) {
	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/duration.test.ts#L3
	t.Run("negative amount", func(t *testing.T) {
		coreRequirePanicContains(t, "negative", func() {
			cdk8s.Duration_Seconds(coreFloat(-1))
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/duration.test.ts#L7
	t.Run("Duration in seconds", func(t *testing.T) {
		duration := cdk8s.Duration_Seconds(coreFloat(300))
		if got := *duration.ToSeconds(nil); got != 300 {
			t.Fatalf("seconds = %v", got)
		}
		if got := *duration.ToMinutes(nil); got != 5 {
			t.Fatalf("minutes = %v", got)
		}
		coreRequirePanicContains(t, "'300 seconds' cannot be converted into a whole number of days", func() {
			duration.ToDays(nil)
		})
		coreRequireClose(t, *duration.ToDays(&cdk8s.TimeConversionOptions{Integral: coreBool(false)}), 300.0/86_400)
		if got := *cdk8s.Duration_Seconds(coreFloat(60 * 60 * 24)).ToDays(nil); got != 1 {
			t.Fatalf("days = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/duration.test.ts#L16
	t.Run("Duration in minutes", func(t *testing.T) {
		duration := cdk8s.Duration_Minutes(coreFloat(5))
		if got := *duration.ToSeconds(nil); got != 300 {
			t.Fatalf("seconds = %v", got)
		}
		if got := *duration.ToMinutes(nil); got != 5 {
			t.Fatalf("minutes = %v", got)
		}
		coreRequirePanicContains(t, "'5 minutes' cannot be converted into a whole number of days", func() {
			duration.ToDays(nil)
		})
		coreRequireClose(t, *duration.ToDays(&cdk8s.TimeConversionOptions{Integral: coreBool(false)}), 300.0/86_400)
		if got := *cdk8s.Duration_Minutes(coreFloat(60 * 24)).ToDays(nil); got != 1 {
			t.Fatalf("days = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/duration.test.ts#L26
	t.Run("Duration in hours", func(t *testing.T) {
		duration := cdk8s.Duration_Hours(coreFloat(5))
		if got := *duration.ToSeconds(nil); got != 18_000 {
			t.Fatalf("seconds = %v", got)
		}
		if got := *duration.ToMinutes(nil); got != 300 {
			t.Fatalf("minutes = %v", got)
		}
		coreRequirePanicContains(t, "'5 hours' cannot be converted into a whole number of days", func() {
			duration.ToDays(nil)
		})
		coreRequireClose(t, *duration.ToDays(&cdk8s.TimeConversionOptions{Integral: coreBool(false)}), 5.0/24)
		if got := *cdk8s.Duration_Hours(coreFloat(24)).ToDays(nil); got != 1 {
			t.Fatalf("days = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/duration.test.ts#L36
	t.Run("seconds to milliseconds", func(t *testing.T) {
		if got := *cdk8s.Duration_Seconds(coreFloat(5)).ToMilliseconds(nil); got != 5_000 {
			t.Fatalf("milliseconds = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/duration.test.ts#L41
	t.Run("Duration in days", func(t *testing.T) {
		duration := cdk8s.Duration_Days(coreFloat(1))
		if got := *duration.ToSeconds(nil); got != 86_400 {
			t.Fatalf("seconds = %v", got)
		}
		if got := *duration.ToMinutes(nil); got != 1_440 {
			t.Fatalf("minutes = %v", got)
		}
		if got := *duration.ToDays(nil); got != 1 {
			t.Fatalf("days = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/duration.test.ts#L48
	t.Run("toIsoString", func(t *testing.T) {
		cases := []struct {
			duration cdk8s.Duration
			want     string
		}{
			{cdk8s.Duration_Seconds(coreFloat(0)), "PT0S"},
			{cdk8s.Duration_Minutes(coreFloat(0)), "PT0S"},
			{cdk8s.Duration_Hours(coreFloat(0)), "PT0S"},
			{cdk8s.Duration_Days(coreFloat(0)), "PT0S"},
			{cdk8s.Duration_Seconds(coreFloat(5)), "PT5S"},
			{cdk8s.Duration_Minutes(coreFloat(5)), "PT5M"},
			{cdk8s.Duration_Hours(coreFloat(5)), "PT5H"},
			{cdk8s.Duration_Days(coreFloat(5)), "PT5D"},
			{cdk8s.Duration_Seconds(coreFloat(1 + 60*(1+60*(1+24)))), "PT1D1H1M1S"},
		}
		for _, testCase := range cases {
			if got := coreStringValue(testCase.duration.ToIsoString()); got != testCase.want {
				t.Errorf("ISO string = %q, want %q", got, testCase.want)
			}
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/duration.test.ts#L62
	t.Run("parse", func(t *testing.T) {
		cases := map[string]float64{
			"PT0S":       0,
			"PT0M":       0,
			"PT0H":       0,
			"PT0D":       0,
			"PT5S":       5,
			"PT5M":       300,
			"PT5H":       18_000,
			"PT5D":       432_000,
			"PT1D1H1M1S": 1 + 60*(1+60*(1+24)),
		}
		for expression, want := range cases {
			if got := *cdk8s.Duration_Parse(coreString(expression)).ToSeconds(nil); got != want {
				t.Errorf("%s seconds = %v, want %v", expression, got, want)
			}
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/duration.test.ts#L76
	t.Run("to human string", func(t *testing.T) {
		cases := []struct {
			duration cdk8s.Duration
			want     string
		}{
			{cdk8s.Duration_Minutes(coreFloat(0)), "0 minutes"},
			{cdk8s.Duration_Minutes(coreFloat(10)), "10 minutes"},
			{cdk8s.Duration_Minutes(coreFloat(1)), "1 minute"},
			{cdk8s.Duration_Minutes(coreFloat(62)), "1 hour 2 minutes"},
			{cdk8s.Duration_Seconds(coreFloat(3666)), "1 hour 1 minute"},
			{cdk8s.Duration_Millis(coreFloat(3000)), "3 seconds"},
			{cdk8s.Duration_Millis(coreFloat(3666)), "3 seconds 666 millis"},
			{cdk8s.Duration_Millis(coreFloat(3.6)), "3.6 millis"},
		}
		for _, testCase := range cases {
			if got := coreStringValue(testCase.duration.ToHumanString()); got != testCase.want {
				t.Errorf("human string = %q, want %q", got, testCase.want)
			}
		}
	})
}
