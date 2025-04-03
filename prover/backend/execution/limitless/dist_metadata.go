package limitless

import (
	"fmt"
	"strings"
	"sync"
)

// ModuleSegmentMap maps `ReqID-ModuleID-SegID-SegGlobalID` to a boolean
// The boolean signifies if the proof associated with the segment is completed
// by the worker or not
type ModuleSegmentMap map[string]bool

// DistMetadata holds the registry map
type DistMetadata struct {
	Registry ModuleSegmentMap
}

// Singleton instance and sync.Once for initialization
var (
	metadataInstance *DistMetadata
	onceMetadata     sync.Once
)

// InitDistMetadata initializes the DistMetadata singleton instance and stores it on memory
// To be called during the setup phase
func InitDistMetadata() *DistMetadata {
	onceMetadata.Do(func() {
		metadataInstance = &DistMetadata{
			Registry: make(ModuleSegmentMap),
		}
	})
	return metadataInstance
}

func GetDistMetada() *DistMetadata {
	return metadataInstance
}

// AddKey adds a new key to the map with a default value of false
func (d *DistMetadata) AddKey(reqID, module string, segID, segGlobalID int) {
	key := fmt.Sprintf("%s-MOD%s-SEG%d-%d", reqID, module, segID, segGlobalID)
	d.Registry[key] = false
}

// SetKeyTrue sets the value of a specific key to true
func (d *DistMetadata) SetKeyTrue(reqID, module string, segID, segGlobalID int) {
	key := fmt.Sprintf("%s-%s-%d-%d", reqID, module, segID, segGlobalID)
	d.Registry[key] = true
}

// AreAllTrueForReqID checks if all values for a given reqID are true
func (d *DistMetadata) AreAllTrueForReqID(reqID string) bool {
	for key, value := range d.Registry {
		if strings.HasPrefix(key, reqID) && !value {
			return false // Found a false value for this reqID
		}
	}
	return true // All values are true (or no keys exist for this reqID)
}

// DeleteByReqID deletes all keys corresponding to the given reqID
func (d *DistMetadata) DeleteByReqID(reqID string) {
	for key := range d.Registry {
		if strings.HasPrefix(key, reqID) {
			delete(d.Registry, key)
		}
	}
}
