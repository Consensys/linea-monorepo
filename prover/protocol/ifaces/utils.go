package ifaces

func AreAllBase(inp []Column) bool {
	for _, v := range inp {
		if !v.IsBase() {
			return false
		}
	}
	return true
}
