// Package units defines the measurement units hotgrin understands: each unit
// belongs to a dimension (mass, length, time, volume) and has a factor to that
// dimension's base unit. Values of the same dimension convert and combine;
// mixing dimensions is a provable mistake.
package units

// Unit describes one unit of measure.
type Unit struct {
	Name   string  // canonical spelling, e.g. "kg"
	Dim    string  // "mass", "length", "time", "volume"
	Factor float64 // multiply by this to reach the dimension's base unit
}

// table: base units are g (mass), m (length), s (time), l (volume).
var table = map[string]Unit{
	// mass
	"mg": {"mg", "mass", 0.001},
	"g":  {"g", "mass", 1},
	"kg": {"kg", "mass", 1000},
	"t":  {"t", "mass", 1_000_000},
	// length
	"mm": {"mm", "length", 0.001},
	"cm": {"cm", "length", 0.01},
	"m":  {"m", "length", 1},
	"km": {"km", "length", 1000},
	// time
	"ms":      {"ms", "time", 0.001},
	"s":       {"s", "time", 1},
	"seconds": {"s", "time", 1},
	"min":     {"min", "time", 60},
	"minutes": {"min", "time", 60},
	"h":       {"h", "time", 3600},
	"hours":   {"h", "time", 3600},
	// volume
	"ml": {"ml", "volume", 0.001},
	"l":  {"l", "volume", 1},
}

// Lookup returns the unit for a spelling, if it is one.
func Lookup(name string) (Unit, bool) {
	u, ok := table[name]
	return u, ok
}

// Is reports whether name is a known unit.
func Is(name string) bool {
	_, ok := table[name]
	return ok
}
