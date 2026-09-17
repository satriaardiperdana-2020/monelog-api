package service

import (
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestMoneyExactBoundsAndFormatting(t *testing.T) {
	for _, value := range []string{"0.01", "43500.00", "999999999999.99"} {
		money, err := ParseMoney(value)
		if err != nil || money.String() != value {
			t.Fatalf("ParseMoney(%q) = %q, %v", value, money.String(), err)
		}
		roundTrip, err := moneyFromNumeric(money.Numeric())
		if err != nil || roundTrip.String() != value {
			t.Fatalf("numeric round trip %q = %q, %v", value, roundTrip.String(), err)
		}
	}
	for _, value := range []string{"0.00", "00.01", "1", "1.0", "1.001", "-1.00", "1000000000000.00"} {
		if _, err := ParseMoney(value); err == nil {
			t.Fatalf("ParseMoney(%q) unexpectedly succeeded", value)
		}
	}
	if _, err := moneyFromNumeric(pgtype.Numeric{Int: big.NewInt(1), Exp: -3, Valid: true}); err == nil {
		t.Fatal("fractional cent was accepted")
	}
}

func TestFormatCentsSupportsNegativeDifference(t *testing.T) {
	if got := formatCents(-43500); got != "-435.00" {
		t.Fatalf("formatCents(-43500) = %q", got)
	}
}
