package util

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

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

// GKNamespacedName returns the namespaced name in the format "group/kind/namespace/name".
func GKNamespacedName(group, kind, namespace, name string) string {
	return fmt.Sprintf("%s/%s/%s/%s", group, kind, namespace, name)
}

// APIVersionToGroup returns the group part of the API version.
// For example, "apps/v1" returns "apps".
func APIVersionToGroup(apiVersion string) string {
	// The group part is before the first "/".
	// For example, "apps/v1" returns "apps".
	// if there is no "/", the whole string is the group.
	if !strings.Contains(apiVersion, "/") {
		return apiVersion
	}

	return apiVersion[:strings.Index(apiVersion, "/")]
}

func ToStrPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func GVRFromUnstructured(u *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	group, err := groupFromUnstructured(u)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	version, err := versionFromUnstructured(u)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	resource, err := resourceFromUnstructured(u)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}, nil
}

func groupFromUnstructured(u *unstructured.Unstructured) (string, error) {
	content := u.UnstructuredContent()
	group, found, err := unstructured.NestedString(content, "spec", "group")
	if !found || err != nil {
		return "", fmt.Errorf("can't find source kind from source CRD: %w", err)
	}

	return group, nil
}

func versionFromUnstructured(u *unstructured.Unstructured) (string, error) {
	content := u.UnstructuredContent()
	var version string
	versions, found, err := unstructured.NestedSlice(content, "spec", "versions")
	if !found || err != nil || len(versions) == 0 {
		version, found, err = unstructured.NestedString(content, "spec", "version")
		if !found || err != nil {
			return "", fmt.Errorf("can't find source version from source CRD: %w", err)
		}
	} else {
		for _, v := range versions {
			if vmap, ok := v.(map[string]interface{}); ok {
				if vmap["served"] == true {
					version = vmap["name"].(string)
					break
				}
			}
		}
	}

	if version == "" {
		return "", fmt.Errorf("can't find source version from source CRD: %w", err)
	}

	return version, nil
}

func resourceFromUnstructured(u *unstructured.Unstructured) (string, error) {
	content := u.UnstructuredContent()
	resource, found, err := unstructured.NestedString(content, "spec", "names", "plural")
	if !found || err != nil {
		return "", fmt.Errorf("can't find source resource from source CRD: %w", err)
	}

	return resource, nil
}
