package service

import "testing"

func TestTemplateInputValidation(t *testing.T) {
	input := TemplateInput{CategoryID: 1, Type: CategoryTypeExpense, Name: " Lunch ", Amount: "43500.00", Title: " Office lunch "}
	validated, err := validateTemplateInput(1, input, false)
	if err != nil || validated.name != "Lunch" || validated.title != "Office lunch" || validated.amount.String() != "43500.00" {
		t.Fatalf("validated=%+v err=%v", validated, err)
	}
	for _, invalid := range []TemplateInput{
		{CategoryID: 0, Type: CategoryTypeExpense, Name: "Lunch", Amount: "1.00", Title: "Lunch"},
		{CategoryID: 1, Type: "transfer", Name: "Lunch", Amount: "1.00", Title: "Lunch"},
		{CategoryID: 1, Type: CategoryTypeExpense, Name: " ", Amount: "1.00", Title: "Lunch"},
		{CategoryID: 1, Type: CategoryTypeExpense, Name: "Lunch", Amount: "0.00", Title: "Lunch"},
		{CategoryID: 1, Type: CategoryTypeExpense, Name: "Lunch", Amount: "1.00", Title: " "},
	} {
		if _, err := validateTemplateInput(1, invalid, false); err == nil {
			t.Fatalf("invalid input accepted: %+v", invalid)
		}
	}
}
