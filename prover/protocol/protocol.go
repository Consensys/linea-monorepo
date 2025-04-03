package protocol

import "github.com/consensys/linea-monorepo/prover/protocol/wizard"

/*
Main object responsible for building an IOP protocol
It is exposed to the user directly, and the user should
use it to specify his protocol
*/
type Builder = wizard.Builder

/*
Function to specify the definition of an IOP
*/
type DefineFunc = wizard.DefineFunc

/*
Compile an IOP from a protocol definition
*/
var Compile = wizard.Compile

/*
Run the prover of the compiled protocol
*/
var Prover = wizard.Prove
