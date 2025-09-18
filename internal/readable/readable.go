package readable

import "fmt"

const (
	KiB uint64 = 1024
	MiB        = KiB * 1024
	GiB        = MiB * 1024
	TiB        = GiB * 1024
	PiB        = TiB * 1024
	EiB        = PiB * 1024
)

func amountString(size uint64) (uint64, string) {
	switch {
	case size < MiB:
		return KiB, "KiB"
	case size < GiB:
		return MiB, "MiB"
	case size < TiB:
		return GiB, "GiB"
	case size < PiB:
		return TiB, "TiB"
	case size < EiB:
		return PiB, "PiB"
	default:
		return EiB, "EiB"
	}
}

func ReadableSizeWithUnit(size uint64, unit uint64) float64 {
	return float64(size) / float64(unit)
}

func ReadableSizePair(size1, size2 uint64) string {
	maxSize := size1
	if size2 > size1 {
		maxSize = size2
	}
	unit, unitStr := amountString(maxSize)
	return fmt.Sprintf("%.2f / %.2f %s",
		ReadableSizeWithUnit(size1, unit),
		ReadableSizeWithUnit(size2, unit),
		unitStr,
	)
}
