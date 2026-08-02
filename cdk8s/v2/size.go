package cdk8s

import (
	"fmt"
	"math"
)

type storageUnit struct {
	label        string
	kibibytes    float64
	abbreviation string
}

var (
	storageKibibytes = storageUnit{label: "kibibytes", kibibytes: 1, abbreviation: "Ki"}
	storageMebibytes = storageUnit{label: "mebibytes", kibibytes: 1024, abbreviation: "Mi"}
	storageGibibytes = storageUnit{label: "gibibytes", kibibytes: 1024 * 1024, abbreviation: "Gi"}
	storageTebibytes = storageUnit{label: "tebibytes", kibibytes: 1024 * 1024 * 1024, abbreviation: "Ti"}
	storagePebibytes = storageUnit{label: "pebibytes", kibibytes: 1024 * 1024 * 1024 * 1024, abbreviation: "Pi"}
)

type sizeImpl struct {
	amount float64
	unit   storageUnit
}

// Create a Storage representing an amount gibibytes.
//
// 1 GiB = 1024 MiB.
func Size_Gibibytes(amount *float64) Size {
	return newSize(requiredSizeAmount(amount), storageGibibytes)
}

// Create a Storage representing an amount kibibytes.
//
// 1 KiB = 1024 bytes.
func Size_Kibibytes(amount *float64) Size {
	return newSize(requiredSizeAmount(amount), storageKibibytes)
}

// Create a Storage representing an amount mebibytes.
//
// 1 MiB = 1024 KiB.
func Size_Mebibytes(amount *float64) Size {
	return newSize(requiredSizeAmount(amount), storageMebibytes)
}

// Create a Storage representing an amount pebibytes.
//
// 1 PiB = 1024 TiB.
func Size_Pebibyte(amount *float64) Size {
	return newSize(requiredSizeAmount(amount), storagePebibytes)
}

// Create a Storage representing an amount tebibytes.
//
// 1 TiB = 1024 GiB.
func Size_Tebibytes(amount *float64) Size {
	return newSize(requiredSizeAmount(amount), storageTebibytes)
}

func requiredSizeAmount(amount *float64) float64 {
	if amount == nil {
		panic("parameter amount is required, but nil was provided")
	}
	return *amount
}

func newSize(amount float64, unit storageUnit) Size {
	if amount < 0 {
		panic(fmt.Sprintf("Storage amounts cannot be negative. Received: %s", formatJSNumber(amount)))
	}
	return &sizeImpl{amount: amount, unit: unit}
}

func (s *sizeImpl) AsString() *string {
	result := formatJSNumber(s.amount) + s.unit.abbreviation
	return &result
}

func (s *sizeImpl) ToGibibytes(opts *SizeConversionOptions) *float64 {
	return sizeFloat(convertSize(s.amount, s.unit, storageGibibytes, opts))
}

func (s *sizeImpl) ToKibibytes(opts *SizeConversionOptions) *float64 {
	return sizeFloat(convertSize(s.amount, s.unit, storageKibibytes, opts))
}

func (s *sizeImpl) ToMebibytes(opts *SizeConversionOptions) *float64 {
	return sizeFloat(convertSize(s.amount, s.unit, storageMebibytes, opts))
}

func (s *sizeImpl) ToPebibytes(opts *SizeConversionOptions) *float64 {
	return sizeFloat(convertSize(s.amount, s.unit, storagePebibytes, opts))
}

func (s *sizeImpl) ToTebibytes(opts *SizeConversionOptions) *float64 {
	return sizeFloat(convertSize(s.amount, s.unit, storageTebibytes, opts))
}

func convertSize(amount float64, from storageUnit, to storageUnit, opts *SizeConversionOptions) float64 {
	rounding := SizeRoundingBehavior_FAIL
	if opts != nil && opts.Rounding != "" {
		rounding = opts.Rounding
	}
	if rounding != SizeRoundingBehavior_FAIL &&
		rounding != SizeRoundingBehavior_FLOOR &&
		rounding != SizeRoundingBehavior_NONE {
		panic(fmt.Sprintf("invalid SizeRoundingBehavior: %s", rounding))
	}
	if from.kibibytes == to.kibibytes {
		return amount
	}
	value := amount * (from.kibibytes / to.kibibytes)
	switch rounding {
	case SizeRoundingBehavior_NONE:
		return value
	case SizeRoundingBehavior_FLOOR:
		return math.Floor(value)
	default:
		if !isInteger(value) {
			panic(fmt.Sprintf("'%s %s' cannot be converted into a whole number of %s.", formatJSNumber(amount), from.label, to.label))
		}
		return value
	}
}

func sizeFloat(value float64) *float64 {
	return &value
}
