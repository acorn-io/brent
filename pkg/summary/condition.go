package summary

import (
	"github.com/acorn-io/schemer/data"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func getRawConditions(obj data.Object) []data.Object {
	return obj.Slice("status", "conditions")
}

func getConditions(obj data.Object) (result []Condition) {
	for _, condition := range getRawConditions(obj) {
		result = append(result, Condition{Object: condition})
	}
	return
}

type Condition struct {
	data.Object
}

func (c Condition) Type() string {
	return c.String("type")
}

func (c Condition) Status() string {
	return c.String("status")
}

func (c Condition) Reason() string {
	return c.String("reason")
}

func (c Condition) Message() string {
	return c.String("message")
}

func (c Condition) Equals(other Condition) bool {
	return c.Type() == other.Type() &&
		c.Status() == other.Status() &&
		c.Reason() == other.Reason() &&
		c.Message() == other.Message()
}

func NormalizeConditions(runtimeObj runtime.Object) {
	var (
		obj           data.Object
		newConditions []map[string]interface{}
	)

	unstr, ok := runtimeObj.(*unstructured.Unstructured)
	if !ok {
		return
	}

	obj = unstr.Object
	for _, condition := range obj.Slice("status", "conditions") {
		var summary Summary
		for _, summarizer := range ConditionSummarizers {
			summary = summarizer(obj, []Condition{{Object: condition}}, summary)
		}
		condition.Set("error", summary.Error)
		condition.Set("transitioning", summary.Transitioning)

		if condition.String("lastUpdateTime") == "" {
			condition.Set("lastUpdateTime", condition.String("lastTransitionTime"))
		}
		newConditions = append(newConditions, condition)
	}

	if len(newConditions) > 0 {
		obj.SetNested(newConditions, "status", "conditions")
	}
}
