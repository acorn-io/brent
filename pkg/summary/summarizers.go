package summary

import (
	"github.com/acorn-io/schemer/data"
)

var (
	Summarizers          []Summarizer
	ConditionSummarizers []Summarizer
)

type Summarizer func(obj data.Object, conditions []Condition, summary Summary) Summary
