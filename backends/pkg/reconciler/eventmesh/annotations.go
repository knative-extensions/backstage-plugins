package eventmesh

var ExcludedAnnotations = map[string]struct{}{
	"kubectl.kubernetes.io/last-applied-configuration": {},
}

func FilterAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}

	ret := make(map[string]string)
	for k, v := range annotations {
		if _, ok := ExcludedAnnotations[k]; ok {
			continue
		}
		ret[k] = v
	}

	if len(ret) == 0 {
		return nil
	}

	return ret
}
