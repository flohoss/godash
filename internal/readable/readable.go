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

func ReadableSize(size uint64) string {
	unit, unitStr := amountString(size)
	return fmt.Sprintf("%.2f %s", float64(size)/float64(unit), unitStr)
}
