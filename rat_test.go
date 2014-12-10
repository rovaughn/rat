package rat

import (
	"bytes"
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

func TestMarshal(t *testing.T) {
	marshalTests := []struct {
		numerator, denominator int64
		expect                 []byte
	}{
		{0, 1, []byte{0}},
		{100, 1, []byte{16, 100}},
		{-5, 1, []byte{16, 251, 255}},
		{10000, 1, []byte{32, 16, 39}},
		{100, 3, []byte{16, 100}},
		{3000, 32, []byte{32, 184, 11}},
	}

	for _, test := range marshalTests {
		actual := Int(test.numerator).Marshal()

		if !bytes.Equal(test.expect, actual) {
			t.Errorf("Expected %d/%d -> %v, got %v", test.numerator, test.denominator, test.expect, actual)
		}

		gobbed, err := big.NewRat(test.numerator, test.denominator).GobEncode()
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("%d/%d: %d vs %d", test.numerator, test.denominator, len(actual), len(gobbed))
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
