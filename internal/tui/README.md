# package tui

`import "github.com/agentflare-ai/amux/internal/tui"`

Package tui provides terminal screen decoding and TUI XML encoding for amux.

The TUI decoder incrementally processes PTY output to build a terminal
screen model, which can be serialized to XML for LLM ingestion.

See spec §7.7 and §11.2.1 for TUI decoding requirements.

- `type Cell` — Cell represents a single terminal cell.
- `type Decoder` — Decoder decodes PTY output into a screen model.
- `type Screen` — Screen represents the current terminal screen state.
- `type Style` — Style represents cell styling.
- `type xmlLine` — xmlLine represents a line in XML output.
- `type xmlScreen` — xmlScreen represents the screen in XML output.

## type Cell

```go
type Cell struct {
	Char  rune
	Style Style
}
```

Cell represents a single terminal cell.

## type Decoder

```go
type Decoder struct {
	screen *Screen
}
```

Decoder decodes PTY output into a screen model.

### Functions returning Decoder

#### NewDecoder

```go
func NewDecoder(rows, cols int) *Decoder
```

NewDecoder creates a new decoder.


### Methods

#### Decoder.Reset

```go
func () Reset()
```

Reset resets the decoder.

#### Decoder.Screen

```go
func () Screen() *Screen
```

Screen returns the current screen state.

#### Decoder.Write

```go
func () Write(data []byte) (int, error)
```

Write processes PTY output.


## type Screen

```go
type Screen struct {
	Rows    int
	Cols    int
	Cells   [][]Cell
	CursorX int
	CursorY int
}
```

Screen represents the current terminal screen state.

### Functions returning Screen

#### NewScreen

```go
func NewScreen(rows, cols int) *Screen
```

NewScreen creates a new screen with the given dimensions.


### Methods

#### Screen.Clear

```go
func () Clear()
```

Clear clears the screen.

#### Screen.Resize

```go
func () Resize(rows, cols int)
```

Resize resizes the screen.

#### Screen.ToXML

```go
func () ToXML() ([]byte, error)
```

ToXML serializes the screen to XML.


## type Style

```go
type Style struct {
	Bold       bool
	Dim        bool
	Italic     bool
	Underline  bool
	Foreground string
	Background string
}
```

Style represents cell styling.

## type xmlLine

```go
type xmlLine struct {
	Row     int    `xml:"row,attr"`
	Content string `xml:",chardata"`
}
```

xmlLine represents a line in XML output.

## type xmlScreen

```go
type xmlScreen struct {
	XMLName xml.Name  `xml:"screen"`
	Rows    int       `xml:"rows,attr"`
	Cols    int       `xml:"cols,attr"`
	Lines   []xmlLine `xml:"line"`
}
```

xmlScreen represents the screen in XML output.

