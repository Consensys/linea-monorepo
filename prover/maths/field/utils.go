package field

import "math"

// ToInt safely converts Element to int with optimized checks
func ToInt(e *Element) int {
    // First check if the element fits in uint64 to avoid invalid value extraction
    if !e.IsUint64() {
        panic("element exceeds uint64 range")
    }
    
    // Now safely get the value after validation
    n := e.Uint64()
    
    // Check against platform-specific int max using bitwise comparison
    if n > math.MaxInt {
        panic("element exceeds int range")
    }
    
    return int(n)
}
