package engine

type File interface {
	Content() string
	FindLineAndColumn(findingIndex int) (line int, column int)
}
