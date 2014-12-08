package rat

import (
	"testing"
)

type stringTest struct {
	given  *Rat
	expect string
}

var stringTests = []stringTest{
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

type addTest struct {
	given1, given2, expect *Rat
}

var addTests = []addTest{
	{Int(1), Int(2), Int(3)},
	{Int(1000), Int(500), Int(1500)},
	{Int(-30), Int(50), Int(20)},
	{Int(-30), Int(-50), Int(-80)},
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
