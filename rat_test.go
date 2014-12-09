package rat

import (
	"testing"
)

var stringTests = []struct {
	given  *Rat
	expect string
}{
	{Int(1), ".01'00"},
	{Int(10), ".0a'00"},
	{Int(-100), ".9c'ff"},
	{Int(500), ".f401'00"},
}

func TestString(t *testing.T) {
	for _, test := range stringTests {
		actual := test.given.String()

		if actual != test.expect {
			t.Errorf("Expected %s, got %s", test.expect, actual)
		}
	}
}

var addTests = []struct {
	given1, given2, expect *Rat
}{
	{Int(1), Int(2), Int(3)},
	{Int(1000), Int(500), Int(1500)},
	{Int(-30), Int(50), Int(20)},
	{Int(-30), Int(-50), Int(-80)},
	{Int(0), Int(-1), Int(-1)},
}

func TestAdd(t *testing.T) {
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

var mulTests = []struct {
	given1, given2, expect *Rat
}{
	{Int(3), Int(4), Int(12)},
	{Int(12), Int(32), Int(384)},
	{Int(-3), Int(12), Int(-36)},
	{Int(-3), Int(-2), Int(6)},
}

func TestMul(t *testing.T) {
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
