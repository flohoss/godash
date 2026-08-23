package readable

import "testing"

func TestReadableSize(t *testing.T) {
	tests := []struct {
		name string
		size uint64
		want string
	}{
		{"zero bytes", 0, "0.00 B"},
		{"single byte", 1, "1.00 B"},
		{"just below kibibyte", 1023, "1023.00 B"},
		{"kibibyte", 1024, "1.00 KiB"},
		{"mebibyte", 1024 * 1024, "1.00 MiB"},
		{"gibibyte", 1024 * 1024 * 1024, "1.00 GiB"},
		{"tebibyte", 1 << 40, "1.00 TiB"},
		{"pebibyte", 1 << 50, "1.00 PiB"},
		{"exbibyte", 1 << 60, "1.00 EiB"},
		{"fractional mebibyte", 3 * 1024 * 1024 / 2, "1.50 MiB"},
		{"above exbibyte", 1<<60 + 1, "1.00 EiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadableSize(tt.size); got != tt.want {
				t.Errorf("ReadableSize(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestReadableSizePair(t *testing.T) {
	tests := []struct {
		name  string
		size1 uint64
		size2 uint64
		want  string
	}{
		{"equal sizes", 1024, 1024, "1.00 / 1.00 KiB"},
		{"second larger sets unit", 512, 2048, "0.50 / 2.00 KiB"},
		{"first larger sets unit", 2048, 512, "2.00 / 0.50 KiB"},
		{"different units", 1024, 1024 * 1024, "0.00 / 1.00 MiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadableSizePair(tt.size1, tt.size2); got != tt.want {
				t.Errorf("ReadableSizePair(%d, %d) = %q, want %q", tt.size1, tt.size2, got, tt.want)
			}
		})
	}
}

func TestReadableSizeWithUnit(t *testing.T) {
	if got := ReadableSizeWithUnit(1024, 1024); got != 1.0 {
		t.Errorf("ReadableSizeWithUnit(1024, 1024) = %v, want 1", got)
	}
	if got := ReadableSizeWithUnit(512, 1024); got != 0.5 {
		t.Errorf("ReadableSizeWithUnit(512, 1024) = %v, want 0.5", got)
	}
}

func TestAmountStringBoundaries(t *testing.T) {
	tests := []struct {
		name string
		size uint64
		unit uint64
		str  string
	}{
		{"below kibibyte", KiB - 1, 1, "B"},
		{"at kibibyte", KiB, KiB, "KiB"},
		{"at mebibyte", MiB, MiB, "MiB"},
		{"at gibibyte", GiB, GiB, "GiB"},
		{"at tebibyte", TiB, TiB, "TiB"},
		{"at pebibyte", PiB, PiB, "PiB"},
		{"at exbibyte", EiB, EiB, "EiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit, str := amountString(tt.size)
			if unit != tt.unit || str != tt.str {
				t.Errorf("amountString(%d) = (%d, %q), want (%d, %q)", tt.size, unit, str, tt.unit, tt.str)
			}
		})
	}
}
