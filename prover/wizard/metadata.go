package wizard

import (
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type metadata struct {
	id         id
	name       string
	doc        string
	declaredBy []string
	tags       []string
	scope      scope
}

// scope is a tree data-structure where each node can refer to is parent (thus,
// the tree is only walkable from children to parent).
type scope struct {
	// parent indicates the parents of the current node in the spantree
	parent *scope
	// selfTags indicates the tags added specifically to the current node
	tags []string
	// name is an identifier for the current node of the scope tree
	name string
}

// allTags returns the list of the tags of present in the current metadata
func (m *metadata) listTags() []string {

	var (
		tagSet  = m.scope.getFullTagSet()
		tagList = make([]string, 0, len(tagSet))
	)

	for _, tag := range m.tags {
		tagSet[tag] = struct{}{}
	}

	for tag := range tagSet {
		tagList = append(tagList, tag)
	}

	slices.Sort(tagList)
	return tagList
}

// enterChilder creates and returns a new child node for s
func (api *API) ChildScope(name string) *API {
	newAPI := *api
	newAPI.currScope = scope{
		parent: &api.currScope,
		name:   name,
	}
	return &newAPI
}

// addTags appends new tags to s
func (api *API) AddTags(tags ...string) *API {
	api.currScope.tags = append(api.currScope.tags, tags...)
	return api
}

// getAllTags returns the set of all the tags in s and its parents deduplicated.
func (s *scope) getFullTagSet() map[string]struct{} {
	tagSet := map[string]struct{}{}
	for ss := s; ss != nil; ss = ss.parent {
		for _, tag := range ss.tags {
			tagSet[tag] = struct{}{}
		}
	}
	return tagSet
}

// getFullScope returns a string repr of the scope in a directory like format
func (s *scope) getFullScope() string {
	nameList := []string{}
	for ss := s; ss != nil; ss = ss.parent {
		nameList = append(nameList, ss.name)
	}
	slices.Reverse(nameList)
	return path.Join(nameList...)
}

// newMetadata initializes a [metadata] reusing the current scope of the API.
// skip is used to set the traceback corresponding to the creation of the object.
//
// The traceback will always contains at most 3 frames.
func (api *API) newMetadata() *metadata {

	frames := getTraceBackFrames(1, 20)
	frames = trimFramesFromCurrDir(frames)

	return &metadata{
		id:         api.newID(),
		scope:      api.currScope,
		declaredBy: formatFrames(frames),
	}
}

// nameOfDefault construct a string repr from the user provided metadata or
// from the parameters if the name is missing. The default name is derived
// from the type of the object and the value of its attribute id. If the id
// is missing or obj is not a struct, the function panics.
func (met *metadata) nameOrDefault(obj any) string {

	if len(met.name) > 0 {
		return met.name
	}

	var (
		valueOfObj = reflect.ValueOf(obj)
		id         = int(valueOfObj.FieldByName("id").Uint())
		typeName   = reflect.TypeOf(obj).Name()
	)

	return strings.ToLower(typeName) + "-" + strconv.Itoa(id)
}

func (met *metadata) explain(obj any) string {

	var (
		b           = &strings.Builder{}
		tracesBacks = met.declaredBy
		doc         = met.doc
		tags        = met.listTags()
	)

	b.WriteString("------------------------------------------------------------\n")
	b.WriteString(met.scope.getFullScope() + "/" + met.nameOrDefault(obj))
	b.WriteString("\n")

	b.WriteString("\ttype:\n")
	b.WriteString("\t\t")
	b.WriteString(reflect.TypeOf(obj).String())
	b.WriteString("\n")

	b.WriteString("\tdeclaredAt:\n")
	for i := range tracesBacks {
		b.WriteString("\t\t")
		b.WriteString(tracesBacks[i])
		b.WriteString("\n")
	}

	b.WriteString("\tdocumentation:\n")
	doc = strings.ReplaceAll(doc, "\n", "\n\t\t")
	b.WriteString("\t\t")
	b.WriteString(doc)
	b.WriteString("\n")

	b.WriteString("\ttags:\n")
	for i := range tags {
		b.WriteString("\t\t")
		b.WriteString(tags[i])
		b.WriteString("\n")
	}

	b.WriteString("\n")

	return b.String()
}
