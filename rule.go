package engine

type UnitType int

const (
	ProgramTextUnit UnitType = iota
)

type Rule interface {
	IsFor(UnitType) bool // Indicates which kind of program unit this rules can be ran on
}
