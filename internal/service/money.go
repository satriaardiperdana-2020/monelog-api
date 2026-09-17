package service

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

const maxMoneyCents int64 = 99999999999999

var moneyPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,11})\.([0-9]{2})$`)

// Money stores IDR amounts as integer cents. It never uses binary floating point.
type Money struct{ cents int64 }

func ParseMoney(value string) (Money, error) {
	matches := moneyPattern.FindStringSubmatch(value)
	if matches == nil {
		return Money{}, errors.New("invalid money")
	}
	whole, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return Money{}, errors.New("invalid money")
	}
	fraction, _ := strconv.ParseInt(matches[2], 10, 64)
	cents := whole*100 + fraction
	if cents < 1 || cents > maxMoneyCents {
		return Money{}, errors.New("money outside range")
	}
	return Money{cents: cents}, nil
}

func moneyFromNumeric(value pgtype.Numeric) (Money, error) {
	if !value.Valid || value.NaN || value.InfinityModifier != pgtype.Finite || value.Int == nil {
		return Money{}, errors.New("invalid numeric money")
	}
	n := new(big.Int).Set(value.Int)
	shift := int(value.Exp) + 2
	if shift > 0 {
		n.Mul(n, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(shift)), nil))
	} else if shift < 0 {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-shift)), nil)
		quotient, remainder := new(big.Int), new(big.Int)
		quotient.QuoRem(n, divisor, remainder)
		if remainder.Sign() != 0 {
			return Money{}, errors.New("numeric has fractional cents")
		}
		n = quotient
	}
	if !n.IsInt64() {
		return Money{}, errors.New("numeric money overflow")
	}
	cents := n.Int64()
	if cents < 0 || cents > maxMoneyCents {
		return Money{}, errors.New("numeric money outside range")
	}
	return Money{cents: cents}, nil
}

func (m Money) Numeric() pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(m.cents), Exp: -2, Valid: true}
}

func (m Money) String() string { return formatCents(m.cents) }
func (m Money) Cents() int64   { return m.cents }

func formatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func normalizeTitle(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) < 1 || len(value) > 200 {
		return "", errors.New("invalid title")
	}
	return value, nil
}
