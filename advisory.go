package engine

import "regexp"

type Advisory interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetRules() []*regexp.Regexp
}
