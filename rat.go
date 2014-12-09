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

func (a *Rat) Marshal() []byte {
	var b []byte

	if len(a.mantissa) <= 0x10 {
		b = make([]byte, 1+len(a.mantissa))
		b[0] = uint8(a.radix) | uint8(a.quote)<<4
		copy(b[1:], a.mantissa)
	} else if len(a.mantissa) <= 0x100 {
		b = make([]byte, 2+len(a.mantissa))
		b[0] = uint8(a.radix)
		b[1] = uint8(a.quote)
		copy(b[2:], a.mantissa)
	} else if len(a.mantissa) <= 0x10000 {
		b = make([]byte, 4+len(a.mantissa))
		binary.LittleEndian.PutUint16(b[0:2], uint16(a.radix))
		binary.LittleEndian.PutUint16(b[2:4], uint16(a.quote))
		copy(b[4:], a.mantissa)
	} else if len(a.mantissa) <= 0x100000000 {
		b = make([]byte, 8+len(a.mantissa))
		binary.LittleEndian.PutUint32(b[0:4], uint32(a.radix))
		binary.LittleEndian.PutUint32(b[4:8], uint32(a.quote))
		copy(b[8:], a.mantissa)
	} else {
		b = make([]byte, 16+len(a.mantissa))
		binary.LittleEndian.PutUint64(b[0:8], uint64(a.radix))
		binary.LittleEndian.PutUint64(b[8:16], uint64(a.quote))
		copy(b[16:], a.mantissa)
	}

	last := len(a.mantissa) - 1
	if a.quote == last && a.mantissa[last] == 0 {
		b = b[:len(b)-1]
	}

	return b
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

func (a *Rat) RShift() *Rat {
	if a.quote == 0 {
		m := make([]uint8, 2*len(a.mantissa)-1)
		copy(m, a.mantissa[1:])
		copy(m[len(a.mantissa)-1:], a.mantissa)
		return &Rat{
			mantissa: m,
			radix:    a.radix,
			quote:    0,
		}
	} else {
		m := make([]uint8, len(a.mantissa)-1)
		copy(m, a.mantissa[1:])
		return &Rat{
			mantissa: m,
			radix:    a.radix,
			quote:    a.quote - 1,
		}
	}
}

func (a *Rat) Mul(b *Rat) *Rat {
	type mulstate struct {
		cursor1, cursor2 int
		carry            *Rat
	}

	outerTable := make([]mulstate, 0)
	outerMantissa := make([]uint8, 0)
	outerState := mulstate{0, 0, Uint8(0)}

	for {
	nextOuterLoop:
		for i := range outerTable {
			if outerTable[i].cursor1 == outerState.cursor1 && outerTable[i].cursor2 == outerState.cursor2 && outerTable[i].carry.Eq(outerState.carry) {
				return &Rat{
					mantissa: outerMantissa,
					quote:    i,
					radix:    a.radix + b.radix,
				}
			}
		}

		outerTable = append(outerTable, outerState)

		firstSum := Uint(uint64(a.mantissa[outerState.cursor1]) * uint64(b.mantissa[outerState.cursor2])).Add(outerState.carry)

		outerMantissa = append(outerMantissa, firstSum.mantissa[0])

		innerState := mulstate{outerState.cursor1, 1, firstSum.RShift()}
		innerTable := make([]mulstate, 0)
		outerCarry := make([]uint8, 0)

		for {
			for i := range innerTable {
				if innerTable[i].cursor1 == innerState.cursor1 && innerTable[i].cursor2 == innerState.cursor2 && innerTable[i].carry.Eq(innerState.carry) {
					outerState.cursor1++
					if outerState.cursor1 == len(a.mantissa) {
						outerState.cursor1 = a.quote
					}

					outerState.cursor2 = 0
					outerState.carry = &Rat{
						mantissa: outerCarry,
						quote:    i,
					}

					goto nextOuterLoop
				}
			}

			innerTable = append(innerTable, innerState)

			product := Uint(uint64(a.mantissa[innerState.cursor1]) * uint64(b.mantissa[innerState.cursor2])).Add(innerState.carry)

			outerCarry = append(outerCarry, product.mantissa[0])
			innerState.carry = product.RShift()
			innerState.cursor2++
			if innerState.cursor2 == len(b.mantissa) {
				innerState.cursor2 = b.quote
			}
		}
	}
}

func Uint8(n uint8) *Rat {
	if n == 0 {
		return &Rat{
			mantissa: []uint8{0},
		}
	} else {
		return &Rat{
			mantissa: []uint8{n, 0},
			quote:    1,
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
			return c
		}
	}

	c.mantissa = c.mantissa[0:1]
	c.quote = 0
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
