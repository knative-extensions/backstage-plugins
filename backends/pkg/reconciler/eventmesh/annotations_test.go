package eventmesh

import (
	"reflect"
	"testing"
)

func TestFilterAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		want        map[string]string
	}{
		{
			name:        "nil annotations",
			annotations: nil,
			want:        nil,
		},
		{
			name:        "empty annotations",
			annotations: map[string]string{},
			want:        nil,
		},
		{
			name:        "excluded annotation",
			annotations: map[string]string{"a": "b", "kubectl.kubernetes.io/last-applied-configuration": "foo"},
			want:        map[string]string{"a": "b"},
		},
		{
			name:        "all excluded",
			annotations: map[string]string{"kubectl.kubernetes.io/last-applied-configuration": "foo"},
			want:        nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterAnnotations(tt.annotations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}
