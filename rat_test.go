package rat

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"
)

func TestString(t *testing.T) {
	stringTests := []struct {
		given  *Rat
		expect string
	}{
		{Int(1), ".01'00"},
		{Int(10), ".0a'00"},
		{Int(-100), ".9c'ff"},
		{Int(500), ".f401'00"},
		{Int(0), "!00"},
	}

	for _, test := range stringTests {
		actual := test.given.String()

		if actual != test.expect {
			t.Errorf("Expected %s, got %s", test.expect, actual)
		}
	}
}

func TestAdd(t *testing.T) {
	addTests := []struct {
		given1, given2, expect *Rat
	}{
		{Int(1), Int(2), Int(3)},
		{Int(1000), Int(500), Int(1500)},
		{Int(-30), Int(50), Int(20)},
		{Int(-30), Int(-50), Int(-80)},
		{Int(0), Int(-1), Int(-1)},
		{Ratio(1, 2), Ratio(1, 2), Int(1)},
		{Ratio(1, 2), Int(1), Ratio(3, 2)},
	}

	for _, test := range addTests {
		actual1 := test.given1.Add(test.given2)
		actual2 := test.given2.Add(test.given1)

		if !actual1.Eq(test.expect) {
			t.Errorf("Expected %s + %s = %s, got %s", test.given1, test.given2, test.expect, actual1)
		}

		if !actual2.Eq(test.expect) {
			t.Errorf("Expected %s + %s = %s, got %s", test.given2, test.given1, test.expect, actual2)
		}
	}
}

func TestMul(t *testing.T) {
	mulTests := []struct {
		given1, given2, expect *Rat
	}{
		{Int(3), Int(4), Int(12)},
		{Int(12), Int(32), Int(384)},
		{Int(-3), Int(12), Int(-36)},
		{Int(-3), Int(-2), Int(6)},
	}

	for _, test := range mulTests {
		actual1 := test.given1.Mul(test.given2)
		actual2 := test.given2.Mul(test.given1)
		if !actual1.Eq(test.expect) {
			t.Errorf("Expected %s * %s = %s, got %s", test.given1, test.given2, test.expect, actual1)
		}

		if !actual2.Eq(test.expect) {
			t.Errorf("Expected %s * %s = %s, got %s", test.given2, test.given1, test.expect, actual2)
		}
	}
}

func TestBadGobDecode(t *testing.T) {
	tests := []struct {
		buffer   []byte
		expected string
	}{
		{[]byte{}, "Gob is too short to decode"},
		{[]byte{0}, "Gob is too short to decode"},
	}

	for _, test := range tests {
		r := new(Rat)
		err := r.GobDecode(test.buffer)

		if err == nil || err.Error() != test.expected {
			t.Errorf("Expected %v, not %v", test.expected, err)
		}
	}
}

func TestGob(t *testing.T) {
	marshalTests := []struct {
		numerator, denominator int64
		expect                 []byte
	}{
		{3, 5, []byte{16, 103, 102}},
		{0, 1, []byte{0, 0}},
		{-1, 1, []byte{0, 255}},
		{100, 1, []byte{16, 100, 0}},
		{-5, 1, []byte{16, 251, 255}},
		{10000, 1, []byte{32, 16, 39, 0}},
		{100, 3, []byte{16, 204, 170}},
		{3000, 32, []byte{33, 192, 93, 0}},
	}

	for _, test := range marshalTests {
		ratio := Ratio(test.numerator, test.denominator)

		ratioGob, err := ratio.GobEncode()
		if err != nil {
			t.Error(err)
			continue
		}

		if !bytes.Equal(test.expect, ratioGob) {
			t.Errorf("Expected %d/%d (%s) -> %x, got %x",
				test.numerator, test.denominator, ratio, test.expect, ratioGob)
		}

		numerator := Int(test.numerator)
		numeratorGob, err := numerator.GobEncode()
		if err != nil {
			t.Error(err)
			continue
		}

		if test.denominator == 1 && !bytes.Equal(ratioGob, numeratorGob) {
			t.Errorf("Expected %d (%s) -> %x, got %x",
				test.numerator, numerator, ratioGob, numeratorGob)
		}

		product := ratio.Mul(Int(test.denominator))
		productGob, err := product.GobEncode()
		if err != nil {
			t.Error(err)
			continue
		}

		if !bytes.Equal(numeratorGob, productGob) {
			t.Errorf("Expected (%d/%d)*%d (%s) -> %x, got %x",
				test.numerator, test.denominator, test.denominator, product,
				numeratorGob, productGob)
		}

		decodedRatio := new(Rat)

		if err := decodedRatio.GobDecode(ratioGob); err != nil {
			t.Errorf("Error when decoding: %v", err)
			continue
		}

		if !decodedRatio.Eq(ratio) {
			t.Errorf("Unmarshaling %v -> %x -> %v", ratio, ratioGob, decodedRatio)
		}
	}
}

func TestRShift(t *testing.T) {
	rshiftTests := []struct {
		given, expected *Rat
	}{
		{Int(1000), Int(3)},
		{Int(3), Int(0)},
		{Int(-1000), Int(-4)},
	}

	for _, test := range rshiftTests {
		actual := test.given.RShift()

		if !actual.Eq(test.expected) {
			t.Errorf("Expected %s >> 1 = %s, got %s", test.given, test.expected, actual)
		}
	}
}

func TestDiv(t *testing.T) {
	divTests := []struct {
		dividend, divisor *Rat
	}{
		{Int(384), Int(256)},
		{Int(-3), Int(-2)},
		{Int(5), Int(5)},
		{Int(5), Int(3)},
		{Int(-5), Int(3)},
		{Int(-3), Int(-2)},
		{Int(10), Int(-32)},
	}

	for _, test := range divTests {
		quotient := test.dividend.Div(test.divisor)

		quotient2 := test.dividend.Negate().Div(test.divisor.Negate())
		if !quotient.Eq(quotient2) {
			t.Errorf("Expected %s / %s = %s = %s / %s = %s", test.dividend, test.divisor, quotient, test.dividend.Negate(), test.divisor.Negate(), quotient2)
		}

		product1 := quotient.Mul(test.divisor)
		if !product1.Eq(test.dividend) {
			t.Errorf("Expected (%s / %s) * %s = %s * %s = %s, got %s", test.dividend, test.divisor, test.divisor, quotient, test.divisor, test.dividend, product1)
		}

		quotient3 := test.dividend.Div(quotient)
		if !quotient3.Eq(test.divisor) {
			t.Errorf("Expected %s / (%s / %s) = %s / %s = %s, got %s", test.dividend, test.dividend, test.divisor, test.dividend, quotient, test.divisor, quotient3)
		}
	}
}

func TestSetRat(t *testing.T) {
	tests := []struct {
		numerator   int64
		denominator int64
	}{
		{3, 100},
		{3, 10},
		{1, 2},
		{100, 1000},
		{-1, -3},
	}

	actual := big.NewRat(1, 1)

	for _, test := range tests {
		rat := Ratio(test.numerator, test.denominator)
		rat.SetRat(actual)

		expect := big.NewRat(test.numerator, test.denominator)

		if actual.Cmp(expect) != 0 {
			t.Errorf("Expected %d/%d to end up with %s, got %s\n",
				test.numerator, test.denominator, expect, actual)
		}
	}
}

func BenchmarkDiv(b *testing.B) {
	numerator := Int(99999)
	denominator := Int(99998)

	for i := 0; i < b.N; i++ {
		numerator.Div(denominator)
	}
}

func ExampleReadme() {
	compare := func(numerator int64, denominator int64) {
		gobRat, err := Ratio(numerator, denominator).GobEncode()
		if err != nil {
			fmt.Println(err)
			return
		}

		gobBigRat, err := big.NewRat(numerator, denominator).GobEncode()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Fraction %d/%d\n", numerator, denominator)
		fmt.Printf("rat.Rat: (%d bytes) %x\n", len(gobRat), gobRat)
		fmt.Printf("big.Rat: (%d bytes) %x\n", len(gobBigRat), gobBigRat)
	}

	compare(1, 3)
	// Output:
	// Fraction 1/3
	// rat.Rat: (3 bytes) 10abaa
	// big.Rat: (7 bytes) 02000000010103
}
