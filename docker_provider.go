package thirdpartyhosting

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// DockerComposeProvider implements the DockerProvider interface using docker-compose
type DockerComposeProvider struct {
	config      ComposeConfig
	initialized bool
	containers  map[string]string // service name -> container ID
	mu          sync.RWMutex
}

// NewDockerComposeProvider creates a new Docker Compose provider
func NewDockerComposeProvider() *DockerComposeProvider {
	return &DockerComposeProvider{
		containers: make(map[string]string),
	}
}

// Initialize sets up the Docker environment and validates the configuration
func (p *DockerComposeProvider) Initialize(ctx context.Context, config ComposeConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config
	p.initialized = true
	return nil
}

// Start creates and starts all Docker containers defined in the compose configuration
func (p *DockerComposeProvider) Start(ctx context.Context) error {
	p.mu.RLock()
	if !p.initialized {
		p.mu.RUnlock()
		return fmt.Errorf("provider not initialized")
	}
	config := p.config
	p.mu.RUnlock()

	// Generate docker-compose.yml file
	composeFile, err := generateComposeFile(config)
	if err != nil {
		return fmt.Errorf("failed to generate compose file: %w", err)
	}

	// Run docker-compose up
	cmd := exec.CommandContext(ctx, "docker-compose", "-p", config.ProjectName, "-f", composeFile, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start containers: %s, error: %w", string(output), err)
	}

	// Update container IDs
	return p.updateContainerIDs(ctx)
}

// Stop gracefully stops and removes all Docker containers
func (p *DockerComposeProvider) Stop(ctx context.Context) error {
	p.mu.RLock()
	if !p.initialized {
		p.mu.RUnlock()
		return fmt.Errorf("provider not initialized")
	}
	config := p.config
	p.mu.RUnlock()

	// Generate docker-compose.yml file
	composeFile, err := generateComposeFile(config)
	if err != nil {
		return fmt.Errorf("failed to generate compose file: %w", err)
	}

	// Run docker-compose down
	cmd := exec.CommandContext(ctx, "docker-compose", "-p", config.ProjectName, "-f", composeFile, "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop containers: %s, error: %w", string(output), err)
	}

	p.mu.Lock()
	p.containers = make(map[string]string)
	p.mu.Unlock()

	return nil
}

// Status returns the current status of all Docker containers
func (p *DockerComposeProvider) Status(ctx context.Context) (map[string]string, error) {
	p.mu.RLock()
	if !p.initialized {
		p.mu.RUnlock()
		return nil, fmt.Errorf("provider not initialized")
	}
	config := p.config
	p.mu.RUnlock()

	// Update container IDs first
	if err := p.updateContainerIDs(ctx); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	statuses := make(map[string]string)
	for service := range config.Services {
		containerID, exists := p.containers[service]
		if !exists {
			statuses[service] = "not_found"
			continue
		}

		cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Status}}", containerID)
		output, err := cmd.CombinedOutput()
		if err != nil {
			statuses[service] = "error"
			continue
		}

		status := strings.TrimSpace(string(output))
		statuses[service] = status
	}

	return statuses, nil
}

// GetLogs retrieves Docker container logs for a specific service
func (p *DockerComposeProvider) GetLogs(ctx context.Context, serviceName string) (io.Reader, error) {
	p.mu.RLock()
	if !p.initialized {
		p.mu.RUnlock()
		return nil, fmt.Errorf("provider not initialized")
	}
	config := p.config
	p.mu.RUnlock()

	// Check if service exists
	if _, exists := config.Services[serviceName]; !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Update container IDs first
	if err := p.updateContainerIDs(ctx); err != nil {
		return nil, err
	}

	p.mu.RLock()
	containerID, exists := p.containers[serviceName]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container for service %s not found", serviceName)
	}

	cmd := exec.CommandContext(ctx, "docker", "logs", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return bytes.NewReader(output), nil
}

// GetContainerID returns the Docker container ID for a specific service
func (p *DockerComposeProvider) GetContainerID(serviceName string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.containers[serviceName]
}

// GetServices returns all service names currently managed by this provider
func (p *DockerComposeProvider) GetServices() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil
	}

	services := make([]string, 0, len(p.config.Services))
	for service := range p.config.Services {
		services = append(services, service)
	}

	return services
}

// updateContainerIDs refreshes the container IDs for all services
func (p *DockerComposeProvider) updateContainerIDs(ctx context.Context) error {
	p.mu.RLock()
	config := p.config
	p.mu.RUnlock()

	containers := make(map[string]string)
	for service := range config.Services {
		cmd := exec.CommandContext(
			ctx,
			"docker-compose",
			"-p", config.ProjectName,
			"ps", "-q", service,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			continue // Skip if service not running
		}

		containerID := strings.TrimSpace(string(output))
		if containerID != "" {
			containers[service] = containerID
		}
	}

	p.mu.Lock()
	p.containers = containers
	p.mu.Unlock()

	return nil
}
