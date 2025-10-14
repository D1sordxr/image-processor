package vo

import "fmt"

type Filename string

const (
	originalFilename  = "original"
	processedFilename = "processed"
)

func NewFilename(s string) Filename {
	return Filename(s)
}

func NewFilenameOriginal(s string) Filename {
	return Filename(fmt.Sprintf("%s:%s", originalFilename, s))
}

func NewFilenameProcessed(s string) Filename {
	return Filename(fmt.Sprintf("%s:%s", processedFilename, s))
}

func (f Filename) String() string {
	return string(f)
}
