package vo

import "fmt"

type Filename string

const originalPath = "original"

func NewFilename(s string) Filename {
	return Filename(s)
}

func NewFilenameOriginal(s string) Filename {
	return Filename(fmt.Sprintf("%s:%s", originalPath, s))
}

func (f Filename) String() string {
	return string(f)
}
