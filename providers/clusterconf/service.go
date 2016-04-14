package clusterconf

// Service is configuration for a service.
type Service struct {
	ID           string                  `json:"id"`
	Dataset      string                  `json:"dataset"`
	HealthChecks map[string]*HealthCheck `json:"healthCheck"`
	Limits       *ResourceLimits         `json:"limits"`
	Env          map[string]string       `json:"env"`
}

// ResourceLimits is configuration for resource upper bounds.
type ResourceLimits struct {
	CPU       int   `json:"cpu"`
	Memory    int64 `json:"memory"`
	Processes int   `json:"processes"`
}

// HealthCheck is configuration for performing a health check.
type HealthCheck struct {
	ID               string   `json:"id"`
	ProtocolProvider string   `json:"protocolProvider"`
	Parameters       []string `json:"parameters"`
}
