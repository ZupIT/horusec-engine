package engine

type Advisory interface {
	GetID() string
	GetName() string
	GetDescription() string
}
