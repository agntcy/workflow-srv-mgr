package k8s

type ChartValues struct {
	Agents []AgentValues `yaml:"agents"`
}

type AgentValues struct {
	Name               string            `yaml:"name"`
	Image              Image             `yaml:"image"`
	Labels             map[string]string `yaml:"labels,omitempty"`
	Env                []EnvVar          `yaml:"env"`
	SecretEnvs         []EnvVar          `yaml:"secretEnvs"`
	ExistingSecretName string            `yaml:"existingSecretName,omitempty"`
	VolumePath         string            `yaml:"volumePath,omitempty"`
	ExternalPort       int               `yaml:"externalPort"`
	InternalPort       int               `yaml:"internalPort"`
	Service            Service           `yaml:"service"`
	StatefulSet        StatefulSet       `yaml:"statefulset"`
}

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}

type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Service struct {
	Type        string            `yaml:"type,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type StatefulSet struct {
	Replicas       int               `yaml:"replicas,omitempty"`
	Labels         map[string]string `yaml:"labels,omitempty"`
	Annotations    map[string]string `yaml:"annotations,omitempty"`
	PodAnnotations map[string]string `yaml:"podAnnotations,omitempty"`
	Resources      Resources         `yaml:"resources,omitempty"`
	NodeSelector   map[string]string `yaml:"nodeSelector,omitempty"`
	Affinity       Affinity          `yaml:"affinity,omitempty"`
	Tolerations    []Toleration      `yaml:"tolerations,omitempty"`
}

type Resources struct {
	Requests map[string]string `yaml:"requests,omitempty"`
	Limits   map[string]string `yaml:"limits,omitempty"`
}

type Affinity struct {
	NodeAffinity NodeAffinity `yaml:"nodeAffinity,omitempty"`
}

type NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution RequiredDuringSchedulingIgnoredDuringExecution `yaml:"requiredDuringSchedulingIgnoredDuringExecution"`
}

type RequiredDuringSchedulingIgnoredDuringExecution struct {
	NodeSelectorTerms []NodeSelectorTerm `yaml:"nodeSelectorTerms"`
}

type NodeSelectorTerm struct {
	MatchExpressions []MatchExpression `yaml:"matchExpressions"`
}

type MatchExpression struct {
	Key      string   `yaml:"key"`
	Operator string   `yaml:"operator"`
	Values   []string `yaml:"values"`
}

type Toleration struct {
	Key      string `yaml:"key"`
	Operator string `yaml:"operator"`
	Effect   string `yaml:"effect"`
}
