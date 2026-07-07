package units

import "testing"

func TestLookup(t *testing.T) {
	kg, ok := Lookup("kg")
	if !ok || kg.Dim != "mass" || kg.Factor != 1000 {
		t.Errorf("kg = %+v ok=%v", kg, ok)
	}
	if _, ok := Lookup("parsec"); ok {
		t.Error("parsec should be unknown")
	}
	// aliases share canonical names
	m1, _ := Lookup("minutes")
	m2, _ := Lookup("min")
	if m1.Name != m2.Name || m1.Factor != 60 {
		t.Errorf("minutes alias wrong: %+v vs %+v", m1, m2)
	}
}
