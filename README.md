# NimsForest Docker

A Go package that provides a programmatic abstraction layer for hosting third-party applications using Docker and Docker Compose.

## Overview

This package simplifies the deployment and management of multi-service Docker applications by providing:

- **Dynamic Docker Compose generation** from Go configurations
- **Container lifecycle management** (start, stop, status monitoring)
- **Service orchestration** with dependency management
- **Log retrieval** and container inspection

## Features

- Programmatic docker-compose.yml generation
- Multi-service application support
- Port mapping and volume management
- Environment variable configuration
- Resource limits and restart policies
- Container status monitoring
- Log streaming capabilities

## Usage

```go
// Create a new Docker Compose provider
provider := NewDockerComposeProvider()

// Define service configuration
config := ComposeConfig{
    ProjectName: "my-app",
    Services: map[string]ServiceConfig{
        "web": {
            ImageName: "myapp",
            ImageTag:  "latest",
            ExposedPorts: []PortMapping{{
                HostPort:      8080,
                ContainerPort: 80,
                Protocol:      "tcp",
            }},
            Environment: map[string]string{
                "ENV": "production",
            },
        },
    },
}

// Initialize and start services
ctx := context.Background()
if err := provider.Initialize(ctx, config); err != nil {
    log.Fatal(err)
}

if err := provider.Start(ctx); err != nil {
    log.Fatal(err)
}
```

## Use Cases

- Platform-as-a-Service deployments
- Third-party application hosting
- Development environment automation
- Multi-service application orchestration

## Part of NimsForest

This package is part of the NimsForest ecosystem, providing Docker management capabilities for the platform's third-party hosting features.