package cdk8s_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

func TestSize(t *testing.T) {
	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/size.test.ts#L3
	t.Run("negative amount", func(t *testing.T) {
		coreRequirePanicContains(t, "negative", func() {
			cdk8s.Size_Kibibytes(coreFloat(-1))
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/size.test.ts#L7
	t.Run("Size in kibibytes", func(t *testing.T) {
		size := cdk8s.Size_Kibibytes(coreFloat(4_294_967_296))
		if got := *size.ToKibibytes(nil); got != 4_294_967_296 {
			t.Fatalf("kibibytes = %v", got)
		}
		if got := *size.ToMebibytes(nil); got != 4_194_304 {
			t.Fatalf("mebibytes = %v", got)
		}
		if got := *size.ToGibibytes(nil); got != 4_096 {
			t.Fatalf("gibibytes = %v", got)
		}
		if got := *size.ToTebibytes(nil); got != 4 {
			t.Fatalf("tebibytes = %v", got)
		}
		coreRequirePanicContains(t, "'4294967296 kibibytes' cannot be converted into a whole number", func() {
			size.ToPebibytes(nil)
		})
		coreRequireClose(t, *size.ToPebibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_NONE}), 4_294_967_296.0/(1024*1024*1024*1024))
		if got := *cdk8s.Size_Kibibytes(coreFloat(4 * 1024 * 1024)).ToGibibytes(nil); got != 4 {
			t.Fatalf("gibibytes = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/size.test.ts#L20
	t.Run("Size in mebibytes", func(t *testing.T) {
		size := cdk8s.Size_Mebibytes(coreFloat(4_194_304))
		if got := *size.ToKibibytes(nil); got != 4_294_967_296 {
			t.Fatalf("kibibytes = %v", got)
		}
		if got := *size.ToMebibytes(nil); got != 4_194_304 {
			t.Fatalf("mebibytes = %v", got)
		}
		if got := *size.ToGibibytes(nil); got != 4_096 {
			t.Fatalf("gibibytes = %v", got)
		}
		if got := *size.ToTebibytes(nil); got != 4 {
			t.Fatalf("tebibytes = %v", got)
		}
		coreRequirePanicContains(t, "'4194304 mebibytes' cannot be converted into a whole number", func() {
			size.ToPebibytes(nil)
		})
		coreRequireClose(t, *size.ToPebibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_NONE}), 4_194_304.0/(1024*1024*1024))
		if got := *cdk8s.Size_Mebibytes(coreFloat(4 * 1024)).ToGibibytes(nil); got != 4 {
			t.Fatalf("gibibytes = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/size.test.ts#L33
	t.Run("Size in gibibyte", func(t *testing.T) {
		size := cdk8s.Size_Gibibytes(coreFloat(5))
		if got := *size.ToKibibytes(nil); got != 5_242_880 {
			t.Fatalf("kibibytes = %v", got)
		}
		if got := *size.ToMebibytes(nil); got != 5_120 {
			t.Fatalf("mebibytes = %v", got)
		}
		if got := *size.ToGibibytes(nil); got != 5 {
			t.Fatalf("gibibytes = %v", got)
		}
		coreRequirePanicContains(t, "'5 gibibytes' cannot be converted into a whole number", func() {
			size.ToTebibytes(nil)
		})
		coreRequireClose(t, *size.ToTebibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_NONE}), 5.0/1024)
		coreRequirePanicContains(t, "'5 gibibytes' cannot be converted into a whole number", func() {
			size.ToPebibytes(nil)
		})
		coreRequireClose(t, *size.ToPebibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_NONE}), 5.0/(1024*1024))
		if got := *cdk8s.Size_Gibibytes(coreFloat(4096)).ToTebibytes(nil); got != 4 {
			t.Fatalf("tebibytes = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/size.test.ts#L47
	t.Run("Size in tebibyte", func(t *testing.T) {
		size := cdk8s.Size_Tebibytes(coreFloat(5))
		if got := *size.ToKibibytes(nil); got != 5_368_709_120 {
			t.Fatalf("kibibytes = %v", got)
		}
		if got := *size.ToMebibytes(nil); got != 5_242_880 {
			t.Fatalf("mebibytes = %v", got)
		}
		if got := *size.ToGibibytes(nil); got != 5_120 {
			t.Fatalf("gibibytes = %v", got)
		}
		if got := *size.ToTebibytes(nil); got != 5 {
			t.Fatalf("tebibytes = %v", got)
		}
		coreRequirePanicContains(t, "'5 tebibytes' cannot be converted into a whole number", func() {
			size.ToPebibytes(nil)
		})
		coreRequireClose(t, *size.ToPebibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_NONE}), 5.0/1024)
		if got := *cdk8s.Size_Tebibytes(coreFloat(4096)).ToPebibytes(nil); got != 4 {
			t.Fatalf("pebibytes = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/size.test.ts#L60
	t.Run("Size in pebibyte", func(t *testing.T) {
		size := cdk8s.Size_Pebibyte(coreFloat(5))
		if got := *size.ToKibibytes(nil); got != 5_497_558_138_880 {
			t.Fatalf("kibibytes = %v", got)
		}
		if got := *size.ToMebibytes(nil); got != 5_368_709_120 {
			t.Fatalf("mebibytes = %v", got)
		}
		if got := *size.ToGibibytes(nil); got != 5_242_880 {
			t.Fatalf("gibibytes = %v", got)
		}
		if got := *size.ToTebibytes(nil); got != 5_120 {
			t.Fatalf("tebibytes = %v", got)
		}
		if got := *size.ToPebibytes(nil); got != 5 {
			t.Fatalf("pebibytes = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/size.test.ts#L70
	t.Run("rounding behavior", func(t *testing.T) {
		size := cdk8s.Size_Mebibytes(coreFloat(5_200))
		coreRequirePanicContains(t, "cannot be converted into a whole number", func() { size.ToGibibytes(nil) })
		coreRequirePanicContains(t, "cannot be converted into a whole number", func() {
			size.ToGibibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_FAIL})
		})
		if got := *size.ToGibibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_FLOOR}); got != 5 {
			t.Fatalf("floor gibibytes = %v", got)
		}
		if got := *size.ToTebibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_FLOOR}); got != 0 {
			t.Fatalf("floor tebibytes = %v", got)
		}
		coreRequireClose(t, *size.ToKibibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_FLOOR}), 5_324_800)
		if got := *size.ToGibibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_NONE}); got != 5.078125 {
			t.Fatalf("fractional gibibytes = %v", got)
		}
		coreRequireClose(t, *size.ToTebibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_NONE}), 5200.0/(1024*1024))
		if got := *size.ToKibibytes(&cdk8s.SizeConversionOptions{Rounding: cdk8s.SizeRoundingBehavior_NONE}); got != 5_324_800 {
			t.Fatalf("kibibytes = %v", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/size.test.ts#L85
	t.Run("asString function gives abbreviated units", func(t *testing.T) {
		cases := map[string]cdk8s.Size{
			"10Ki": cdk8s.Size_Kibibytes(coreFloat(10)),
			"10Mi": cdk8s.Size_Mebibytes(coreFloat(10)),
			"10Gi": cdk8s.Size_Gibibytes(coreFloat(10)),
			"10Ti": cdk8s.Size_Tebibytes(coreFloat(10)),
			"10Pi": cdk8s.Size_Pebibyte(coreFloat(10)),
		}
		for want, size := range cases {
			if got := coreStringValue(size.AsString()); got != want {
				t.Errorf("size = %q, want %q", got, want)
			}
		}
	})
}
