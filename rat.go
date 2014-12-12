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

var Zero = &Rat{[]byte{0}, 0, 0}

func (a *Rat) Eq(b *Rat) bool {
	a.normalize()
	b.normalize()
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

	a.normalize()

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

	return b
}

func Unmarshal(b []byte) (*Rat, error) {
	if len(b) <= 0x11 {
		return &Rat{
			radix:    int(b[0] & 0x0f),
			quote:    int((b[0] >> 4) & 0x0f),
			mantissa: b[1:],
		}, nil
	} else if len(b) < 0x102 {
		return &Rat{
			radix:    int(b[0]),
			quote:    int(b[1]),
			mantissa: b[2:],
		}, nil
	} else if len(b) < 0x10004 {
		return &Rat{
			radix:    int(binary.LittleEndian.Uint16(b[0:2])),
			quote:    int(binary.LittleEndian.Uint16(b[2:4])),
			mantissa: b[4:],
		}, nil
	} else if len(b) < 0x100000008 {
		return &Rat{
			radix:    int(binary.LittleEndian.Uint32(b[0:4])),
			quote:    int(binary.LittleEndian.Uint32(b[4:8])),
			mantissa: b[8:],
		}, nil
	} else {
		return &Rat{
			radix:    int(binary.LittleEndian.Uint64(b[0:8])),
			quote:    int(binary.LittleEndian.Uint64(b[8:16])),
			mantissa: b[16:],
		}, nil
	}
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

	// Now b.radix >= a.radix
	if a.radix < b.radix {
		r := b.radix - a.radix
		m := make([]uint8, len(a.mantissa)+r)
		copy(m[r:], a.mantissa)
		a = &Rat{
			mantissa: m,
			quote:    a.quote + r,
			radix:    a.radix + r,
		}
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

func (a *Rat) normalize() *Rat {
	for i := 0; i < a.radix && a.mantissa[i] == 0; i++ {
		a.mantissa = a.mantissa[1:]
		a.quote -= 1
		a.radix -= 1
	}

	quotelen := len(a.mantissa) - a.quote

	for i := a.quote - 1; i >= 0; i-- {
		if a.mantissa[i] == a.mantissa[i+quotelen] {
			a.quote -= 1
			a.mantissa = a.mantissa[0 : len(a.mantissa)-1]
		} else {
			break
		}
	}

	for chunklen := 1; chunklen < quotelen; chunklen++ {
		// Only iterate over factors of quotelen.
		if quotelen%chunklen != 0 {
			continue
		}

		nchunks := quotelen / chunklen
		firstChunk := a.mantissa[a.quote : a.quote+chunklen]

		allEqual := true

		for i := 1; i < nchunks; i++ {
			if !bytes.Equal(firstChunk, a.mantissa[a.quote+i*chunklen:a.quote+(i+1)*chunklen]) {
				allEqual = false
				break
			}
		}

		if allEqual {
			a.mantissa = a.mantissa[:a.quote+chunklen]
			break
		}
	}

	return a
}

// TODO: Normalized?
func (a *Rat) RShift() *Rat {
	if a.quote == 0 {
		m := make([]uint8, len(a.mantissa))
		copy(m[1:], a.mantissa)
		m[0] = a.mantissa[len(a.mantissa)-1]
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

func (a *Rat) Div256() *Rat {
	if a.mantissa[0] == 0 && a.quote > 0 {
		return &Rat{
			mantissa: a.mantissa[1:],
			radix:    a.radix,
			quote:    a.quote - 1,
		}
	} else {
		radix := a.radix + 1
		if radix == len(a.mantissa) {
			radix = a.quote
		}

		return &Rat{
			mantissa: a.mantissa,
			radix:    radix,
			quote:    a.quote,
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
				return (&Rat{
					mantissa: outerMantissa,
					quote:    i,
					radix:    a.radix + b.radix,
				}).normalize()
			}
		}

		outerTable = append(outerTable, outerState)

		firstSum := Uint(uint64(a.mantissa[outerState.cursor1]) * uint64(b.mantissa[outerState.cursor2])).Add(outerState.carry)

		outerMantissa = append(outerMantissa, firstSum.mantissa[0])

		innerState := mulstate{outerState.cursor1, 1, firstSum.RShift()}
		innerTable := make([]mulstate, 0)
		outerCarry := make([]uint8, 0)

		if innerState.cursor2 == len(b.mantissa) {
			innerState.cursor2 = b.quote
		}

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

// FindMod(x, z) = y such that (x*y) mod 256 = z
// Only valid for odd values of x.
func FindMod(x uint8, z uint8) uint8 {
	for y := 0; y < 256; y++ {
		if uint8((uint16(x)*uint16(y))&0xff) == z {
			return uint8(y)
		}
	}

	return 0
}

// TODO: Normalized?
func (a *Rat) Div(b *Rat) *Rat {
	for {
		if b.radix > 0 {
			a = a.Mul(Uint(256))
			b = b.Mul(Uint(256))
		} else if b.mantissa[0] == 0 {
			a = a.Div256()
			b = b.Div256()
		} else if b.mantissa[0]&1 == 0 {
			a = a.Mul(Uint8(2))
			b = b.Mul(Uint8(2))
		} else {
			break
		}
	}

	radix := a.radix + b.radix

	a = &Rat{
		mantissa: a.mantissa,
		quote:    a.quote,
		radix:    0,
	}
	b = &Rat{
		mantissa: b.mantissa,
		quote:    b.quote,
		radix:    0,
	}

	// First question:
	// let x = first digit of b
	// let z = first digit of a
	// find y such that x*y mod 256 = z
	divisor := b
	result := make([]uint8, 0)

	type divState struct {
		dividend *Rat
		quotient uint8
	}

	history := make([]divState, 0)
	dividend := a

	for {
		quotient := FindMod(divisor.mantissa[0], dividend.mantissa[0])

		for i := range history {
			if history[i].dividend.Eq(dividend) && history[i].quotient == quotient {
				return &Rat{
					mantissa: result,
					quote:    i,
					radix:    radix,
				}
			}
		}

		history = append(history, divState{dividend, quotient})
		result = append(result, quotient)

		dividend = dividend.Sub(Uint8(quotient).Mul(divisor)).RShift()
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
	if n == 0 {
		return Zero
	}

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

	panic("It should only be possible to get here if n is 0, but it's not.")
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

func Ratio(n int64, d int64) *Rat {
	return Int(n).Div(Int(d))
}

func (a *Rat) Negate() *Rat {
	return a.Complement().Add(Uint8(1))
}

func (a *Rat) Sub(b *Rat) *Rat {
	return a.Add(b.Negate())
}
