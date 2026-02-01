package tools

// boolPtr returns a pointer to a bool value.
// Used for optional *bool fields in mcp.ToolAnnotations.
func boolPtr(b bool) *bool { return &b }
