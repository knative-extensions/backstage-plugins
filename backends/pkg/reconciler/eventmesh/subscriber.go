package eventmesh

type Subscriber struct {
	Group                string            `json:"group,omitempty"`
	Version              string            `json:"version"`
	Kind                 string            `json:"kind"`
	Namespace            string            `json:"namespace"`
	Name                 string            `json:"name"`
	UID                  string            `json:"uid"`
	Labels               map[string]string `json:"labels,omitempty"`
	Annotations          map[string]string `json:"annotations,omitempty"`
	SubscribedEventTypes []string          `json:"subscribedEventTypes,omitempty"`
	BackstageId          string            `json:"backstageId"`
}
