package modulediscoverer_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/stretchr/testify/assert"
)

// This should be re-written for ifaces.Column instead of smartvectors.Smartvector
type ModuleName string

type DisjointSetTest struct {
	parent map[smartvectors.SmartVector]smartvectors.SmartVector
	rank   map[smartvectors.SmartVector]int
}

func NewDisjointSetTest() *DisjointSetTest {
	return &DisjointSetTest{
		parent: make(map[smartvectors.SmartVector]smartvectors.SmartVector),
		rank:   make(map[smartvectors.SmartVector]int),
	}
}

func (ds *DisjointSetTest) Find(vec smartvectors.SmartVector) smartvectors.SmartVector {
	if _, exists := ds.parent[vec]; !exists {
		ds.parent[vec] = vec
		ds.rank[vec] = 0
	}
	if ds.parent[vec] != vec {
		ds.parent[vec] = ds.Find(ds.parent[vec])
	}
	return ds.parent[vec]
}

func (ds *DisjointSetTest) Union(vec1, vec2 smartvectors.SmartVector) {
	root1 := ds.Find(vec1)
	root2 := ds.Find(vec2)

	if root1 != root2 {
		if ds.rank[root1] > ds.rank[root2] {
			ds.parent[root2] = root1
		} else if ds.rank[root1] < ds.rank[root2] {
			ds.parent[root1] = root2
		} else {
			ds.parent[root2] = root1
			ds.rank[root1]++
		}
	}
}

type Module struct {
	moduleName ModuleName
	ds         *DisjointSetTest
}

type Discoverer struct {
	mutex           sync.Mutex
	modules         []*Module
	moduleNames     []ModuleName
	columnsToModule map[smartvectors.SmartVector]ModuleName
}

func NewDiscovererTest() *Discoverer {
	return &Discoverer{
		modules:         []*Module{},
		moduleNames:     []ModuleName{},
		columnsToModule: make(map[smartvectors.SmartVector]ModuleName),
	}
}

func (disc *Discoverer) ModuleList() []ModuleName {
	disc.mutex.Lock()
	defer disc.mutex.Unlock()
	return disc.moduleNames
}

func (disc *Discoverer) assignModule(moduleName ModuleName, vectors []smartvectors.SmartVector) {
	for _, vec := range vectors {
		disc.columnsToModule[vec] = moduleName
	}
}

func (disc *Discoverer) CreateModule(vectors []smartvectors.SmartVector) *Module {
	module := &Module{
		moduleName: ModuleName(fmt.Sprintf("Module_%d", len(disc.modules))),
		ds:         NewDisjointSetTest(),
	}

	for _, vec := range vectors {
		module.ds.parent[vec] = vec
		module.ds.rank[vec] = 0
	}

	// Union all vectors together in the module
	for i := 0; i < len(vectors); i++ {
		for j := i + 1; j < len(vectors); j++ {
			module.ds.Union(vectors[i], vectors[j])
		}
	}

	fmt.Println("Final parent map for module:", module.moduleName, module.ds.parent)
	disc.modules = append(disc.modules, module)
	return module
}

func HasOverlap(module *Module, vectors []smartvectors.SmartVector) bool {
	for _, vec := range vectors {
		fmt.Println("Checking vector:", vec, "against module:", module.moduleName)
		if _, exists := module.ds.parent[vec]; exists {
			fmt.Println("Overlap found between:", vec, "and module:", module.moduleName)
			return true
		} else {
			fmt.Println("Vector:", vec, "NOT found in module:", module.moduleName)
		}
	}
	return false
}

func TestUnion(t *testing.T) {
	ds := NewDisjointSetTest()
	vec1 := smartvectors.ForTest(1)
	vec2 := smartvectors.ForTest(2)
	vec3 := smartvectors.ForTest(3)
	ds.Union(vec1, vec2)
	assert.Equal(t, ds.Find(vec1), ds.Find(vec2), "vec1 and vec2 should have the same root")
	ds.Union(vec2, vec3)
	assert.Equal(t, ds.Find(vec1), ds.Find(vec3), "vec1, vec2, and vec3 should have the same root")
}

func TestCreateModule(t *testing.T) {
	disc := NewDiscovererTest()
	vectors := []smartvectors.SmartVector{smartvectors.ForTest(1), smartvectors.ForTest(2), smartvectors.ForTest(3)}
	module := disc.CreateModule(vectors)

	assert.NotNil(t, module, "Module should not be nil")
	assert.Equal(t, 1, len(disc.modules), "Discoverer should have one module")
	for _, vec := range vectors {
		assert.Equal(t, module.ds.Find(vec), module.ds.Find(vectors[0]), "All vectors should belong to the same set")
	}
}
func TestHasOverlap(t *testing.T) {
	disc := NewDiscovererTest()

	vec1 := smartvectors.ForTest(1)
	vec2 := smartvectors.ForTest(2)
	vec3 := smartvectors.ForTest(3)
	vec4 := smartvectors.ForTest(4)
	vec5 := smartvectors.ForTest(5)
	vec6 := smartvectors.ForTest(6)
	vec7 := smartvectors.ForTest(7)

	module1 := disc.CreateModule([]smartvectors.SmartVector{vec1, vec2})
	module2 := disc.CreateModule([]smartvectors.SmartVector{vec3, vec4})
	candidates := []smartvectors.SmartVector{vec2, vec5}

	assert.False(t, HasOverlap(module1, []smartvectors.SmartVector{vec6, vec7}), "module1 should NOT overlap with unrelated vectors")
	assert.True(t, HasOverlap(module1, candidates), "module1 should overlap with candidates")
	assert.False(t, HasOverlap(module2, candidates), "module2 should NOT overlap with candidates")
}
