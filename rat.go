package rat

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

type Rat struct {
	mantissa []uint8
	radix    int
	quote    int
}

func (a *Rat) Eq(b *Rat) bool {
	return a.radix == b.radix && a.quote == b.quote && bytes.Equal(a.mantissa, b.mantissa)
}

func (a *Rat) String() string {
	var buf bytes.Buffer

	dst := make([]byte, 2)

	for i := range a.mantissa {
		if i == a.radix && i == a.quote {
			buf.WriteByte('!')
		} else if i == a.radix {
			buf.WriteByte('.')
		} else if i == a.quote {
			buf.WriteByte('\'')
		}

		hex.Encode(dst, a.mantissa[i:i+1])
		buf.Write(dst)
	}

	return buf.String()
}

type sumState struct {
	carry   uint8
	cursor1 int
	cursor2 int
}

func (a *Rat) Add(b *Rat) *Rat {
	if b.radix < a.radix {
		a, b = b, a
	}

	states := make([]sumState, 0)
	radixDiff := a.radix - b.radix

	c := &Rat{
		radix:    a.radix,
		mantissa: make([]uint8, 0),
	}

	var cursor1 int
	var cursor2 int
	var sum uint16

	for {
		states = append(states, sumState{
			carry:   uint8(sum),
			cursor1: cursor1,
			cursor2: cursor2,
		})

		if cursor2 < radixDiff {
			sum += uint16(a.mantissa[cursor1])
		} else {
			sum += uint16(a.mantissa[cursor1]) + uint16(b.mantissa[cursor2])
		}

		c.mantissa = append(c.mantissa, uint8(sum&0xff))
		sum >>= 8

		cursor1++
		if cursor1 == len(a.mantissa) {
			cursor1 = a.quote
		}

		cursor2++
		if cursor2 == len(b.mantissa) {
			cursor2 = b.quote
		}

		for i, state := range states {
			if state.carry == uint8(sum) &&
				state.cursor1 == cursor1 &&
				state.cursor2 == cursor2 {
				c.quote = i
				return c
			}
		}
	}
}

func Uint(n uint64) *Rat {
	c := &Rat{
		mantissa: make([]uint8, 9),
		radix:    0,
	}

	binary.LittleEndian.PutUint64(c.mantissa, n)

	for i := 8; i >= 0; i-- {
		if c.mantissa[i] != 0 {
			c.quote = i + 1
			c.mantissa = c.mantissa[0 : i+2]
			break
		}
	}

	return c
}

func Int(n int64) *Rat {
	if n < 0 {
		return Uint(uint64(-n)).Negate()
	} else {
		return Uint(uint64(n))
	}
}

func (a *Rat) Complement() *Rat {
	c := &Rat{
		mantissa: make([]uint8, len(a.mantissa)),
		radix:    a.radix,
		quote:    a.quote,
	}

	for i := range c.mantissa {
		c.mantissa[i] = 0xff - a.mantissa[i]
	}

	return c
}

func (a *Rat) Negate() *Rat {
	return a.Complement().Add(Uint(1))
}

func (a *Rat) Sub(b *Rat) *Rat {
	return a.Add(b.Negate())
}
