package util

import (
	"fmt"

	v1 "knative.dev/pkg/apis/duck/v1"
)

// NamespacedRefName returns the namespaced name of the given reference.
// If the reference is nil, it returns an empty string.
// It returns the namespaced name in the format "namespace/name".
func NamespacedRefName(ref *v1.KReference) string {
	if ref == nil {
		return ""
	}
	return NamespacedName(ref.Namespace, ref.Name)
}

// NamespacedName returns the namespaced name in the format "namespace/name".
func NamespacedName(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func ToStrPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
