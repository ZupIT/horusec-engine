package engine

type MatchType int
type UnitType int

const (
	Regular MatchType = iota
	NotMatch
	OrMatch
	AndMatch

	ProgramTextUnit UnitType = iota
)

type Rule interface {
	IsFor() UnitType // Indicates which kind of program unit this rules can be ran on
	Type() MatchType // Indicates which type of match this rule uses
}
