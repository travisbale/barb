package test_test

import _ "embed"

//go:embed testdata/test_phishlet.yaml
var validPhishletYAML string

//go:embed testdata/miraged.yaml
var miragedConfig string

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
