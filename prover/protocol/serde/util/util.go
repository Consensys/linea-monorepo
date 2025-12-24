package util

// --- REPRODUCTION TYPES ---
// These mimic the structure of symbolic.Variable (Container)
// and the Interface holding a nil pointer.

type ReproInterface interface {
	Foo()
}

type ReproConcrete struct {
	Val int
}

func (r *ReproConcrete) Foo() {}

type ReproContainer struct {
	Name string
	// This is the trigger: An interface that will hold a nil pointer
	Meta ReproInterface
}

// Global variable to ensure the exact same data structure is used in both tests
func GetReproObject() ReproContainer {
	var nilPtr *ReproConcrete = nil // The pointer is nil...
	return ReproContainer{
		Name: "CrashTest",
		Meta: nilPtr, // ...but the interface is NOT nil (it is a Typed Nil)
	}
}

// --- RECURSIVE TYPES ---

// RecursiveNode points to itself.
// Graph: A -> A (Self Loop)
type RecursiveNode struct {
	Name string
	Next *RecursiveNode
}

// GetRecursiveObject creates a tight loop:
// Root -> Root
func GetRecursiveObject() *RecursiveNode {
	n := &RecursiveNode{Name: "TheLoop"}
	n.Next = n // CRITICAL: Point to itself
	return n
}

// --- MOCK TYPES (Mirroring wizard/recursion structures) ---

// 1. MockIOP mimics wizard.CompiledIOP
type MockIOP struct {
	Name string
	// This mimics comp.SubProvers (Slice of Interfaces)
	SubProvers []MockAction
}

// 2. MockAction mimics wizard.ProverAction (Interface)
type MockAction interface {
	Run()
}

// 3. MockRecursion mimics recursion.Recursion
type MockRecursion struct {
	Name string
	// Recursion holds the Plonk Context
	PlonkCtx MockPlonkCtx
	// Recursion holds the Input IOP
	InputIOP *MockIOP
}

// 4. MockPlonkCtx mimics plonkinternal.PlonkInWizardProverAction
type MockPlonkCtx struct {
	// PlonkCtx holds a reference BACK to the IOP it verifies.
	// This creates the CYCLE: IOP -> Action -> Recursion -> PlonkCtx -> IOP
	CheckingIOP *MockIOP
}

// 5. AssignVortexUAlpha mimics the struct used in DefineRecursionOf
type AssignVortexUAlpha struct {
	// This holds the pointer to the Recursion struct
	Ctx *MockRecursion
}

func (a AssignVortexUAlpha) Run() {}

// --- LOGIC REPRODUCTION ---

func GetReproLogicObject() *MockIOP {
	// 1. Create the Input IOP (wiop)
	wiop := &MockIOP{Name: "Input-WIOP"}

	// 2. Create the Result IOP (comp/rec)
	// This is the object we are serializing
	comp := &MockIOP{Name: "Result-IOP"}

	// 3. Create the Plonk Context (Logic from DefineRecursionOf)
	// plonkinternal.PlonkCheck(comp, ...) -> links back to comp
	plonkCtx := MockPlonkCtx{
		CheckingIOP: comp, // <--- CYCLE FORMED HERE
	}

	// 4. Create the Recursion Object
	rec := &MockRecursion{
		Name:     "Recursion-Struct",
		PlonkCtx: plonkCtx,
		InputIOP: wiop,
	}

	// 5. Register the Prover Action (Logic from DefineRecursionOf)
	// comp.RegisterProverAction(1, AssignVortexUAlpha{Ctxs: rec})
	action := AssignVortexUAlpha{
		Ctx: rec,
	}

	// Add to the slice of interfaces
	comp.SubProvers = []MockAction{action}

	return comp
}

// --- MOCK STRUCTURES (Exported so Reflection works) ---

// 1. The shared object (Simulating the compiled IOP/Expression)
type SharedLeaf struct {
	ID   int
	Name string
}

// 2. The container (Simulating the recursion object)
type RootContainer struct {
	// A direct typed pointer to the leaf
	DirectPtr *SharedLeaf

	// A map holding the SAME leaf inside an interface (Simulating ExtraData)
	ExtraData map[string]interface{}
}

// Helper to create the object graph
func GetDebugObject() *RootContainer {
	// Create ONE instance of the leaf
	leaf := &SharedLeaf{ID: 101, Name: "TheSharedOne"}

	return &RootContainer{
		DirectPtr: leaf,
		ExtraData: map[string]interface{}{
			// This interface{} holds a pointer to the SAME leaf
			"aliased_ref": leaf,
		},
	}
}

// --- COMPLEX MOCK STRUCTURES ---

type Parent struct {
	Name     string
	Children []*Child
}

type Child struct {
	ID     int
	Parent *Parent // Cycle back to parent
}

// WrapperC struct often found in ExtraData (e.g. Symbolic Variable)
type WrapperC struct {
	Ref *Child
}

type RootComplex struct {
	MainParent *Parent
	// Map holds interface{} which holds Wrapper (struct), which holds *Child
	ExtraData map[string]interface{}
}

func GetComplexReproObject() *RootComplex {
	parent := &Parent{Name: "TheParent"}
	child := &Child{ID: 1, Parent: parent}
	parent.Children = append(parent.Children, child)

	// Root holds Parent directly
	// Root also holds a Wrapper in ExtraData, which points to the SAME child
	// The child points back to the SAME parent.
	return &RootComplex{
		MainParent: parent,
		ExtraData: map[string]interface{}{
			"wrapper_key": WrapperC{Ref: child},
		},
	}
}

// --- MOCK STRUCTURES (Mimicking the Recursion Graph) ---

// 1. MockIOP: The Root Object (Equivalent to CompiledIOP)
type MockIOP1 struct {
	ID int
	// SubProvers is the "Trap". It is a slice of interfaces.
	// The serializer must unwrap this interface to find the concrete pointer.
	SubProvers []any
}

// 2. MockAction: The Concrete Object (Equivalent to AssignVortexUAlpha)
// This sits inside the interface.
type MockAction1 struct {
	Name string
	Ctx  *MockRecursionCtx1
}

// 3. MockRecursionCtx: Intermediate context (Equivalent to Recursion)
type MockRecursionCtx1 struct {
	PlonkCtx *MockPlonkCtx1
}

// 4. MockPlonkCtx: The Deepest Context (Equivalent to Plonk)
type MockPlonkCtx1 struct {
	// THE CYCLE: This pointer MUST point back to the original MockIOP (ID: 1337)
	// If Aliasing fails, this will point to a COPY of MockIOP.
	CheckingIOP *MockIOP1
}

// GetReproLogicObject builds the circular graph:
// Root -> Interface -> Action -> Rec -> Plonk -> Root
func GetReproLogicObject1() *MockIOP1 {
	// A. Create the Root
	root := &MockIOP1{ID: 1337}

	// B. Create the Deepest Link (Plonk) pointing back to Root
	plonk := &MockPlonkCtx1{CheckingIOP: root}

	// C. Create the Chain upwards
	rec := &MockRecursionCtx1{PlonkCtx: plonk}
	action := &MockAction1{Name: "RecursionAction", Ctx: rec}

	// D. Embed the Action into the Root's Interface Slice
	// This completes the setup. Root holds Action (via interface), Action holds Root (via pointers).
	root.SubProvers = []any{action}

	return root
}

type FakeCommittedMatrix struct {
	Limbs [][]uint64
}

type Wrapper struct {
	Committed any
}
