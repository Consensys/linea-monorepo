package template

import _ "embed"

//go:embed doc.go.tmpl
var DocTemplate string

//go:embed cgo.go.tmpl
var CgoTemplate string

//go:embed fr.go.tmpl
var FrTemplate string

//go:embed fr_stub.go.tmpl
var FrStubTemplate string

//go:embed fr_test.go.tmpl
var FrTestTemplate string

//go:embed fft.go.tmpl
var FFTTemplate string

//go:embed fft_stub.go.tmpl
var FFTStubTemplate string

//go:embed fft_test.go.tmpl
var FFTTestTemplate string

//go:embed msm.go.tmpl
var MSMTemplate string

//go:embed msm_stub.go.tmpl
var MSMStubTemplate string

//go:embed msm_test.go.tmpl
var MSMTestTemplate string
