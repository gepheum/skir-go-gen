package e2e_test

import (
	enums_skir "e2e-test/skirout/enums"
)

func Foo() {
	j := enums_skir.JsonValue_unknown()
	j.IsUnknown()
}
