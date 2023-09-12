package hclhelpers

import "github.com/hashicorp/hcl/v2"

func ExpressionsEqual(expr1, expr2 hcl.Expression) bool {
	if expr1 == nil && expr2 == nil {
		return true
	}

	if expr1 == nil || expr2 == nil {
		return false
	}

	if len(expr1.Variables()) != len(expr2.Variables()) {
		return false
	}

	for i, v := range expr1.Variables() {
		v2 := expr2.Variables()[i]

		if v.RootName() != v2.RootName() {
			return false
		}
	}

	return true
}
