package v1

import "knative.dev/backstage-plugins/backends/pkg/util"

func (gknn GroupKindNamespacedName) String() string {
	return util.GKNamespacedName(gknn.Group, gknn.Kind, gknn.Namespace, gknn.Name)
}
