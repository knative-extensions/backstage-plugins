package eventmesh

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "knative.dev/pkg/apis/duck/v1"
)

func ObjNameAndNamespace(obj metav1.ObjectMetaAccessor) string {
	return NameAndNamespace(obj.GetObjectMeta().GetNamespace(), obj.GetObjectMeta().GetName())
}

func RefNameAndNamespace(ref *v1.KReference) string {
	if ref == nil {
		return ""
	}
	return NameAndNamespace(ref.Namespace, ref.Name)
}

func NameAndNamespace(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func Deduplicate(arr []string) []string {
	m := make(map[string]bool)
	for _, v := range arr {
		m[v] = true
	}

	// create a new slice and add the keys to it
	var result []string
	for key := range m {
		result = append(result, key)
	}

	return result
}
