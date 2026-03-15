package port

// ScriptWriter writes generated script content to the state directory.
type ScriptWriter interface {
	Write(scriptName string, content string) (string, error)
}
