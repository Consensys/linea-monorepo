package serialization

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

func init() {
	// This registers all the types that we may need to deserialize
	// as interfaces implementations. Cases of interfaces that are relevant
	// are:
	//
	// 		- symbolic.Operator
	//      - symbolic.Metadata
	//      - ifaces.Column
	//  	- ifaces.Query
	//
	RegisterImplementation(ifaces.ColID(""))
	RegisterImplementation(ifaces.QueryID(""))
	RegisterImplementation(coin.Name(""))
	RegisterImplementation(column.Natural{})
	RegisterImplementation(column.Shifted{})
	RegisterImplementation(verifiercol.ConstCol{})
	RegisterImplementation(verifiercol.FromYs{})
	RegisterImplementation(verifiercol.FromAccessors{})
	RegisterImplementation(verifiercol.ExpandedVerifCol{})
	RegisterImplementation(verifiercol.RepeatedAccessor{})
	RegisterImplementation(query.FixedPermutation{})
	RegisterImplementation(query.GlobalConstraint{})
	RegisterImplementation(query.Inclusion{})
	RegisterImplementation(query.InnerProduct{})
	RegisterImplementation(query.LocalConstraint{})
	RegisterImplementation(query.LocalOpening{})
	RegisterImplementation(query.MiMC{})
	RegisterImplementation(query.Permutation{})
	RegisterImplementation(query.Range{})
	RegisterImplementation(query.UnivariateEval{})
	RegisterImplementation(symbolic.Variable{})
	RegisterImplementation(symbolic.Constant{})
	RegisterImplementation(symbolic.Product{})
	RegisterImplementation(symbolic.LinComb{})
	RegisterImplementation(symbolic.PolyEval{})
	RegisterImplementation(coin.Info{})
	RegisterImplementation(accessors.FromCoinAccessor{})
	RegisterImplementation(accessors.FromConstAccessor{})
	RegisterImplementation(accessors.FromExprAccessor{})
	RegisterImplementation(accessors.FromIntVecCoinPositionAccessor{})
	RegisterImplementation(accessors.FromLocalOpeningYAccessor{})
	RegisterImplementation(accessors.FromPublicColumn{})
	RegisterImplementation(accessors.FromUnivXAccessor{})
	RegisterImplementation(variables.X{})
	RegisterImplementation(variables.PeriodicSample{})
	RegisterImplementation(symbolic.StringVar(""))
}

// In order to save some space, we trim the prefix of the package path as this
// is repetitive.
const pkgPathPrefixToRemove = "github.com/consensys/linea-monorepo/prover"

// implementationRegistry maps a string representing a string of the form
// `path/to/package#ImplementingStruct` to a [reflect.Type]
// of the struct `ImplementingStruct` where the struct can be anything we would
// like to potentially unmarshal.
var implementationRegistry = collection.NewMapping[string, reflect.Type]()

// RegisterImplementation registers the type of the provided instance. This is
// needed if the caller of the package wants to deserialize into an interface
// type. When that happens, the deserializer needs to know which type to
// concretely deserialize into. This is achieved by looking up the interface
// implementation within the registry.
//
// If the provided instance is a pointer type, then the registry will only store
// the base type of the instance.
//
// If the provided type is an interface or a pointer to interface, then the
// function will refuse.
//
// If the provided type was already registered, then this function is a no-op.
func RegisterImplementation(instance any) {

	if instance == nil {
		// If nil is provided, we cannot neither derive an underlying type nor
		// use the result of the constructor as a receiver to deserialize an
		// instance.
		utils.Panic("The constructor returned nil")
	}

	// Using a loop here ensures that all the required levels of indirections are
	// done so that we are sure we don't register a pointer-type.
	typeOfInstance := reflect.TypeOf(instance)
	for typeOfInstance.Kind() == reflect.Pointer {
		typeOfInstance = typeOfInstance.Elem()
	}

	if len(typeOfInstance.Name()) == 0 {
		utils.Panic("unsupported type of instance: %T", instance)
	}

	registeredTypeName := getPkgPathAndTypeName(instance)

	if implementationRegistry.Exists(registeredTypeName) {
		return
	}

	implementationRegistry.InsertNew(registeredTypeName, reflect.TypeOf(instance))
}

// returns a reflect.Type registered in the registry for the provided type-string
// the function will modify the provided string in case the string represents a
// pointer type and will add the levels of indirections to the returned
// reflect.Type.
func findRegisteredImplementation(pkgTypeName string) (reflect.Type, error) {

	// The typename may contain indirections. In that case, our goal is to
	// resolve the name of the concrete underlying type since this is the one
	// that is registered. The format of the registered name is a#b and the
	// format of the caller's pkgTypeName is a#b#n
	split := strings.Split(pkgTypeName, "#")
	if len(split) != 3 {
		return nil, fmt.Errorf("provided type `%v` does not have the format `a#b#n`", pkgTypeName)
	}

	var (
		registeredName     = split[0] + "#" + split[1]
		nbIndirection, err = strconv.ParseInt(split[2], 10, 64)
	)

	if err != nil {
		return nil, fmt.Errorf("could not parse the number of indirection from the string `%v` : %w", pkgTypeName, err)
	}

	// We need an explicit check because the getter panics. Also it makes for a
	// nicer error message.
	if !implementationRegistry.Exists(registeredName) {
		return nil, fmt.Errorf("unregistered type %s", registeredName)
	}

	foundType := implementationRegistry.MustGet(registeredName)

	// This readds the levels of indirection that the caller requested
	for i := int64(0); i < nbIndirection; i++ {
		foundType = reflect.PointerTo(foundType)
	}

	// This sanity-checks an invariant of the function. No matter what is
	// returned, it must return a type whose pkgPathTypeName matches the
	// requested one.
	if getPkgPathAndTypeNameIndirect(foundType) != pkgTypeName {
		utils.Panic("caller requested `%v` and got `%v`", pkgTypeName, getPkgPathAndTypeNameIndirect(foundType))
	}

	return foundType, nil
}

// Returns the full `<Type.PkgPath>#<Type.Name>#<nbIndirection>` of a type.
// Caller can either provide an instance of the desired type or a reflect.Type
// of it.
//
// The function only supports concrete named types or pointers to them. If the
// caller provides an interface, an anonymous (aside from pointers) type or
// pointers to them. The function will panic.
//
// This is used for naming the types that we would want to resolve. But this is
// not what is concretely registered. The reason for the difference is that
// we don't want to force the user to register every possible types AND their
// pointers.
func getPkgPathAndTypeNameIndirect(x any) string {

	refType := reflect.TypeOf(x)

	// If provided a reflect.Type, don't use the TypeOf of that. Instead directly
	// use the provided Type.
	if xAsRefType, ok := x.(reflect.Type); ok {
		refType = xAsRefType
	}

	nbIndirection := 0
	for refType.Kind() == reflect.Pointer {
		nbIndirection++
		refType = refType.Elem()
	}

	var (
		pkgPath  = refType.PkgPath()
		typeName = refType.Name()
	)

	if len(typeName) == 0 {
		utils.Panic("got an untyped parameter `(%T)(%v)`; this is not supported", x, x)
	}

	// The parenthesis are needed to ensure that the returned string is parseable
	return strings.TrimPrefix(pkgPath, pkgPathPrefixToRemove) + "#" + typeName + "#" + strconv.Itoa(nbIndirection)
}

// Returns the name in the format <pkg>#<type>. Used to derive the type key in
// the register. It has the following restrictions:
//   - Requires that provided type is concrete
//   - If it's a [reflect.Type] this is fine.
func getPkgPathAndTypeName(x any) string {

	refType := reflect.TypeOf(x)

	// If provided a reflect.Type, don't use the TypeOf of that. Instead directly
	// use the provided Type.
	if xAsRefType, ok := x.(reflect.Type); ok {
		refType = xAsRefType
	}

	var (
		pkgPath  = refType.PkgPath()
		typeName = refType.Name()
	)

	if len(typeName) == 0 {
		utils.Panic("got an untyped parameter `(%T)(%v)`; this is not supported", x, x)
	}

	return strings.TrimPrefix(pkgPath, pkgPathPrefixToRemove) + "#" + typeName

}
