package im

type ElementType int

const (
	TextElement ElementType = iota + 1
	AtElement
	SharpElement
	LinkElement
	ImgElement
	AudioElement
	VideoElement
	FileElement
	StrongElement
	EmElement
	InsElement
	DelElement
	SplElement
	CodeElement
	SupElement
	SubElement
	BrElement
	QuoteElement
	AuthorElement
	MessageElement
)

type Element struct {
	Type ElementType
	Data any
}

func To[T any](el *Element) T {
	var v T
	if e, ok := el.Data.(T); ok {
		return e
	}
	return v
}

func n() {
	i := To[string](&Element{})
	_ = i
}
