package flag

var (
	Host           string
	Bind           string
	Port           int
	Token          string
	Version        string
	Mode           string
	ActionsProfile string

	MaxInlineAttachmentBytes int

	Insecure      bool
	ReadOnly      bool
	Debug         bool
	AllowedTools  map[string]struct{}
	AllowedScopes map[string]struct{}
)
