package service

import "testing"

func TestCategoryValidation(t *testing.T) {
	name, err := validateCategoryName("  Food  ")
	if err != nil || name != "Food" {
		t.Fatalf("name=%q err=%v", name, err)
	}
	if _, err := validateCategoryName("   "); err == nil {
		t.Fatal("blank category name was accepted")
	}
	if validCategoryType("transfer") || !validCategoryType(CategoryTypeIncome) || !validCategoryType(CategoryTypeExpense) {
		t.Fatal("category type validation is incorrect")
	}
}
