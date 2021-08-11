package types

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Cents is a wrapper for all values expressed as cents
type Cents int64

func (c Cents) ToMillicents() Millicents {
	return Millicents(int64(c) * int64(1000))
}

func (c *Cents) Scan(value interface{}) error {
	*c = Cents(value.(int64))
	return nil
}

func (c Cents) ToString() string {
	return fmt.Sprintf("%s.%s", c.fmtDollars(), c.fmtCents())
}

func (c Cents) fmtDollars() string {
	str := strconv.Itoa(int(c))

	if len(str) < 3 {
		return "0"
	}

	return addCommas(str[0 : len(str)-2])
}

func (c Cents) fmtCents() string {
	str := strconv.Itoa(int(c))

	if len(str) < 2 {
		return "0" + str
	}

	return str[len(str)-2 : len(str)]
}

func addCommas(s string) string {
	startOffset := 0
	var buff bytes.Buffer

	l := len(s)

	commaIndex := 3 - ((l - startOffset) % 3)

	if commaIndex == 3 {
		commaIndex = 0
	}

	for i := startOffset; i < l; i++ {
		if commaIndex == 3 {
			buff.WriteRune(',')
			commaIndex = 0
		}
		commaIndex++

		buff.WriteByte(s[i])
	}

	return buff.String()
}

// Wrapper for all millicent values
type Millicents int64

func (m Millicents) ToCents() Cents {
	cents := Cents(m / 1000)

	if m > 0 {
		if m%1000 >= 500 {
			cents = cents + 1
		}
	} else {
		if m%1000 <= -500 {
			cents = cents - 1
		}
	}

	return cents
}

func USDToMillicents(usd string) (Millicents, error) {
	if strings.HasPrefix(usd, "$") {
		usd = usd[1 : len(usd)-1]
	}

	// split "12.34" => []string{"12", "34}
	pieces := strings.Split(usd, ".")

	// sanitize
	if len(pieces) != 2 {
		return Millicents(0), fmt.Errorf("USD string improperly formatted")
	}
	if len(pieces[0]) < 1 {
		return Millicents(0), fmt.Errorf("Cents portion on USD has insufficient digits")
	}
	if len(pieces[1]) > 2 || len(pieces[1]) == 0 {
		return Millicents(0), fmt.Errorf("Cents portion of USD incorrectly formatted")
	}
	if len(pieces[1]) == 1 {
		pieces[1] = pieces[1] + "0"
	}

	// refs
	ds := pieces[0]
	cs := pieces[1]

	// uints
	di, err := strconv.ParseInt(ds, 10, 64)
	if err != nil {
		return Millicents(0), err
	}
	ci, err := strconv.ParseInt(cs, 10, 64)
	if err != nil {
		return Millicents(0), err
	}

	mc := (di * 100000) + (ci * 1000)
	return Millicents(mc), nil
}

func USDToCents(usd string) (Cents, error) {
	millicents, err := USDToMillicents(usd)
	if err != nil {
		return 0, err
	}

	return millicents.ToCents(), nil
}
