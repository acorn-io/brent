package common

import (
	"slices"
	"strings"

	"github.com/acorn-io/brent/pkg/accesscontrol"
	"github.com/acorn-io/brent/pkg/attributes"
	"github.com/acorn-io/brent/pkg/schema"
	"github.com/acorn-io/brent/pkg/stores/proxy"
	"github.com/acorn-io/brent/pkg/summary"
	"github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer/data"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	schema2 "k8s.io/apimachinery/pkg/runtime/schema"
)

func DefaultTemplate(clientGetter proxy.ClientGetter,
	asl accesscontrol.AccessSetLookup) schema.Template {
	return schema.Template{
		Store:     proxy.NewProxyStore(clientGetter, asl),
		Formatter: formatter,
	}
}

func selfLink(gvr schema2.GroupVersionResource, meta metav1.Object) (prefix string) {
	buf := &strings.Builder{}
	if gvr.Group == "" {
		buf.WriteString("/api/v1/")
	} else {
		buf.WriteString("/apis/")
		buf.WriteString(gvr.Group)
		buf.WriteString("/")
		buf.WriteString(gvr.Version)
		buf.WriteString("/")
	}
	if meta.GetNamespace() != "" {
		buf.WriteString("namespaces/")
		buf.WriteString(meta.GetNamespace())
		buf.WriteString("/")
	}
	buf.WriteString(gvr.Resource)
	buf.WriteString("/")
	buf.WriteString(meta.GetName())
	return buf.String()
}

func formatter(request *types.APIRequest, resource *types.RawResource) {
	if resource.Schema == nil {
		return
	}

	gvr := attributes.GVR(resource.Schema)
	if gvr.Version == "" {
		return
	}

	meta, err := meta.Accessor(resource.APIObject.Object)
	if err != nil {
		return
	}
	selfLink := selfLink(gvr, meta)

	u := request.URLBuilder.RelativeToRoot(selfLink)
	resource.Links["view"] = u

	if _, ok := resource.Links["update"]; !ok && slices.Contains(resource.Schema.CollectionMethods, "PUT") {
		resource.Links["update"] = u
	}

	if _, ok := resource.Links["update"]; !ok && slices.Contains(resource.Schema.ResourceMethods, "blocked-PUT") {
		resource.Links["update"] = "blocked"
	}

	if _, ok := resource.Links["remove"]; !ok && slices.Contains(resource.Schema.ResourceMethods, "blocked-DELETE") {
		resource.Links["remove"] = "blocked"
	}

	if unstr, ok := resource.APIObject.Object.(*unstructured.Unstructured); ok {
		s := summary.Summarized(unstr)
		data.PutValue(unstr.Object, map[string]interface{}{
			"name":          s.State,
			"error":         s.Error,
			"transitioning": s.Transitioning,
			"message":       strings.Join(s.Message, ":"),
		}, "metadata", "state")

		summary.NormalizeConditions(unstr)

		includeFields(request, unstr)
		excludeManagedFields(request, unstr)
		excludeFields(request, unstr)
		excludeValues(request, unstr)
	}


}

func includeFields(request *types.APIRequest, unstr *unstructured.Unstructured) {
	if fields, ok := request.Query["include"]; ok {
		newObj := map[string]interface{}{}
		for _, f := range fields {
			fieldParts := strings.Split(f, ".")
			if val, ok := data.GetValue(unstr.Object, fieldParts...); ok {
				data.PutValue(newObj, val, fieldParts...)
			}
		}
		unstr.Object = newObj
	}
}

func excludeManagedFields(request *types.APIRequest, unstr *unstructured.Unstructured) {
	data.RemoveValue(unstr.Object, "metadata", "managedFields")
}

func excludeFields(request *types.APIRequest, unstr *unstructured.Unstructured) {
	if fields, ok := request.Query["exclude"]; ok {
		for _, f := range fields {
			fieldParts := strings.Split(f, ".")
			data.RemoveValue(unstr.Object, fieldParts...)
		}
	}
}

func excludeValues(request *types.APIRequest, unstr *unstructured.Unstructured) {
	if values, ok := request.Query["excludeValues"]; ok {
		for _, f := range values {
			fieldParts := strings.Split(f, ".")
			fieldValues := data.GetValueN(unstr.Object, fieldParts...)
			if obj, ok := fieldValues.(map[string]interface{}); ok {
				for k := range obj {
					data.PutValue(unstr.Object, "", append(fieldParts, k)...)
				}
			}
		}
	}
}
