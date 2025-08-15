package thirdpartyhosting

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// generateComposeFile creates a temporary docker-compose.yml file from the config
func generateComposeFile(config ComposeConfig) (string, error) {
	// Create a temporary directory for the compose file
	tempDir, err := ioutil.TempDir("", "docker-compose-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate the compose file content
	content, err := generateComposeContent(config)
	if err != nil {
		return "", fmt.Errorf("failed to generate compose content: %w", err)
	}

	// Write the content to a file
	composeFilePath := filepath.Join(tempDir, "docker-compose.yml")
	if err := ioutil.WriteFile(composeFilePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	return composeFilePath, nil
}

// generateComposeContent creates the content for a docker-compose.yml file
func generateComposeContent(config ComposeConfig) (string, error) {
	var sb strings.Builder

	// Write the version
	sb.WriteString("version: \"3.4\"\n\n")

	// Write the services section
	sb.WriteString("services:\n")
	for serviceName, serviceConfig := range config.Services {
		sb.WriteString(fmt.Sprintf("  %s:\n", serviceName))
		sb.WriteString(fmt.Sprintf("    image: %s:%s\n", serviceConfig.ImageName, serviceConfig.ImageTag))

		// Write restart policy if specified
		if serviceConfig.RestartPolicy != "" {
			sb.WriteString(fmt.Sprintf("    restart: %s\n", serviceConfig.RestartPolicy))
		}

		// Write port mappings if any
		if len(serviceConfig.ExposedPorts) > 0 {
			sb.WriteString("    ports:\n")
			for _, port := range serviceConfig.ExposedPorts {
				sb.WriteString(fmt.Sprintf("      - \"%d:%d/%s\"\n", port.HostPort, port.ContainerPort, port.Protocol))
			}
		}

		// Write volume mappings if any
		if len(serviceConfig.Volumes) > 0 {
			sb.WriteString("    volumes:\n")
			for _, volume := range serviceConfig.Volumes {
				sb.WriteString(fmt.Sprintf("      - %s:%s\n", volume.HostPath, volume.ContainerPath))
			}
		}

		// Write environment variables if any
		if len(serviceConfig.Environment) > 0 {
			sb.WriteString("    environment:\n")
			for key, value := range serviceConfig.Environment {
				sb.WriteString(fmt.Sprintf("      - %s=%s\n", key, value))
			}
		}

		// Write dependencies if any
		if len(serviceConfig.DependsOn) > 0 {
			sb.WriteString("    depends_on:\n")
			for _, dep := range serviceConfig.DependsOn {
				sb.WriteString(fmt.Sprintf("      - %s\n", dep))
			}
		}

		// Write resource limits if specified
		if serviceConfig.Resources.Memory != "" || serviceConfig.Resources.CPUShare != "" {
			sb.WriteString("    deploy:\n")
			sb.WriteString("      resources:\n")
			sb.WriteString("        limits:\n")
			if serviceConfig.Resources.Memory != "" {
				sb.WriteString(fmt.Sprintf("          memory: %s\n", serviceConfig.Resources.Memory))
			}
			if serviceConfig.Resources.CPUShare != "" {
				sb.WriteString(fmt.Sprintf("          cpus: %s\n", serviceConfig.Resources.CPUShare))
			}
		}
	}

	// Write the networks section if a network is specified
	if config.Network != "" {
		sb.WriteString("\nnetworks:\n")
		sb.WriteString(fmt.Sprintf("  %s:\n", config.Network))
		sb.WriteString("    driver: bridge\n")
	}

	return sb.String(), nil
}

// CleanupComposeFile removes the temporary docker-compose.yml file
func CleanupComposeFile(composeFilePath string) error {
	// Remove the parent directory and all its contents
	return os.RemoveAll(filepath.Dir(composeFilePath))
}
