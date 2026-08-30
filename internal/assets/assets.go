package assets

import _ "embed"

// HAHashBaselineSQL is the embedded HRPAuth baseline SQL template.
//
//go:embed ha-baseline.sql
var HABaselineSQL []byte
