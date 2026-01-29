# package tui

`import "github.com/agentflare-ai/amux/internal/tui"`

Package tui handles terminal screen decoding and XML encoding.

- `func stripANSI(str string) string`
- `type Screen` — Screen represents the decoded state of a terminal screen.

### Functions

#### stripANSI

```go
func stripANSI(str string) string
```


## type Screen

```go
type Screen struct {
	Rows []string `xml:"row"`
}
```

Screen represents the decoded state of a terminal screen.

### Functions returning Screen

#### DecodeScreen

```go
func DecodeScreen(data []byte) *Screen
```

DecodeScreen decodes raw PTY output into a Screen model.
For Phase 5, this is a basic implementation that splits by newline
and strips some common ANSI sequences (simplified).


### Methods

#### Screen.EncodeXML

```go
func () EncodeXML() ([]byte, error)
```

EncodeXML encodes the screen state to XML format.


