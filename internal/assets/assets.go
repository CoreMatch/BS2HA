package assets

import _ "embed"

// HABaselineSQL is the embedded HRPAuth core baseline SQL template.
//
//go:embed ha-baseline.sql
var HABaselineSQL []byte

// YGGBaselineSQL is the embedded HRPAuth Yggdrasil baseline SQL template.
//
//go:embed ygg-baseline.sql
var YGGBaselineSQL []byte
