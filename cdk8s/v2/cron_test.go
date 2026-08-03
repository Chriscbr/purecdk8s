package cdk8s_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

func TestCron(t *testing.T) {
	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/cron.test.ts#L4
	t.Run("cron expression for running every minute", func(t *testing.T) {
		if got := coreStringValue(cdk8s.Cron_EveryMinute().ExpressionString()); got != "* * * * *" {
			t.Fatalf("expression = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/cron.test.ts#L8
	t.Run("cron expression for running every hour", func(t *testing.T) {
		if got := coreStringValue(cdk8s.Cron_Hourly().ExpressionString()); got != "0 * * * *" {
			t.Fatalf("expression = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/cron.test.ts#L12
	t.Run("cron expression for running every day", func(t *testing.T) {
		if got := coreStringValue(cdk8s.Cron_Daily().ExpressionString()); got != "0 0 * * *" {
			t.Fatalf("expression = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/cron.test.ts#L16
	t.Run("cron expression for running every week", func(t *testing.T) {
		if got := coreStringValue(cdk8s.Cron_Weekly().ExpressionString()); got != "0 0 * * 0" {
			t.Fatalf("expression = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/cron.test.ts#L20
	t.Run("cron expression for running every month", func(t *testing.T) {
		if got := coreStringValue(cdk8s.Cron_Monthly().ExpressionString()); got != "0 0 1 * *" {
			t.Fatalf("expression = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/cron.test.ts#L24
	t.Run("cron expression for running every year", func(t *testing.T) {
		if got := coreStringValue(cdk8s.Cron_Annually().ExpressionString()); got != "0 0 1 1 *" {
			t.Fatalf("expression = %q", got)
		}
	})

	expression := &cdk8s.CronOptions{
		Minute:  coreString("5"),
		Hour:    coreString("*"),
		Day:     coreString("2"),
		Month:   coreString("*"),
		WeekDay: coreString("*"),
	}

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/cron.test.ts#L28
	t.Run("custom schedule cron expression", func(t *testing.T) {
		if got := coreStringValue(cdk8s.Cron_Schedule(expression).ExpressionString()); got != "5 * 2 * *" {
			t.Fatalf("expression = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/cron.test.ts#L39
	t.Run("custom schedule cron expression using default initialization", func(t *testing.T) {
		if got := coreStringValue(cdk8s.NewCron(nil).ExpressionString()); got != "* * * * *" {
			t.Fatalf("expression = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/cron.test.ts#L44
	t.Run("custom schedule cron expression using initialization", func(t *testing.T) {
		if got := coreStringValue(cdk8s.NewCron(expression).ExpressionString()); got != "5 * 2 * *" {
			t.Fatalf("expression = %q", got)
		}
	})
}
