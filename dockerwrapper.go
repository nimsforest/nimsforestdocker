package thirdpartyhosting

import (
	"context"
	"io"
)

// ServiceConfig contains configuration for a single Docker service
type ServiceConfig struct {
	// Basic configuration
	ImageName    string
	ImageTag     string // e.g., "stable" for Fider
	ExposedPorts []PortMapping
	Environment  map[string]string
	Volumes      []VolumeMapping

	// Dependencies
	DependsOn []string // e.g., Fider depends on "db"

	// Restart policy
	RestartPolicy string // e.g., "always"

	// Resource constraints
	Resources ResourceLimits
}

// PortMapping defines how ports are mapped from host to container
type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string // "tcp" or "udp"
}

// VolumeMapping defines how volumes are mapped
type VolumeMapping struct {
	HostPath      string // e.g., "/var/fider/pg_data"
	ContainerPath string // e.g., "/var/lib/postgresql/data"
}

// ResourceLimits defines container resource constraints
type ResourceLimits struct {
	Memory   string // e.g., "512m"
	CPUShare string // e.g., "0.5"
}

// ComposeConfig represents the configuration for multiple Docker services
type ComposeConfig struct {
	Services map[string]ServiceConfig
	Network  string

	// Global settings
	ProjectName string // Name for the compose project
	EnvFile     string // Path to .env file if used
}

// DockerProvider defines the interface for Docker-based service hosting
type DockerProvider interface {
	// Initialize sets up the Docker environment and validates the configuration
	// ComposeConfig contains settings for all services that need to be run together
	Initialize(ctx context.Context, config ComposeConfig) error

	// Start creates and starts all Docker containers defined in the compose configuration
	Start(ctx context.Context) error

	// Stop gracefully stops and removes all Docker containers
	Stop(ctx context.Context) error

	// Status returns the current status of all Docker containers
	// Returns a map of service names to their status: "running", "stopped", "error", "not_found"
	Status(ctx context.Context) (map[string]string, error)

	// GetLogs retrieves Docker container logs for a specific service
	// Returns an io.Reader for streaming the container logs
	GetLogs(ctx context.Context, serviceName string) (io.Reader, error)

	// GetContainerID returns the Docker container ID for a specific service
	GetContainerID(serviceName string) string

	// GetServices returns all service names currently managed by this provider
	GetServices() []string
}
