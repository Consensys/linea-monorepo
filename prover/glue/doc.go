// Package glue calls complex gnark circuits.
//
// Wizard allows to define IOPs for some data. In case where it would be too
// difficult to hand-craft (or to compile from the ZKEVM arithmetization
// specification) the IOP for complex relations (precompiles mostly), we instead
// define the relations as gnark circuit, obtain it as PLONK IOP and check it in
// the Wizard.
//
// This package implements all the boilerplate, e.g. extracting witness data
// from the execution trace, converting it to expected encoding in gnark,
// perform sanity checks etc. To include the glued circuits in ZKEVM, it should
// be sufficient to call `Register...` methods after calling the main `Define`
// function.
//
// Currently implemented glues are:
//   - [RegisterECDSA] - verify multiple ECDSA circuits.
package glue
