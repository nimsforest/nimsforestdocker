package thirdpartyhosting

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDockerProvider implements the DockerProvider interface for testing
type MockDockerProvider struct {
	mock.Mock
}

func (m *MockDockerProvider) Initialize(ctx context.Context, config ComposeConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockDockerProvider) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDockerProvider) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDockerProvider) Status(ctx context.Context) (map[string]string, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockDockerProvider) GetLogs(ctx context.Context, serviceName string) (io.Reader, error) {
	args := m.Called(ctx, serviceName)
	return args.Get(0).(io.Reader), args.Error(1)
}

func (m *MockDockerProvider) GetContainerID(serviceName string) string {
	args := m.Called(serviceName)
	return args.String(0)
}

func (m *MockDockerProvider) GetServices() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestDockerProviderInitialize(t *testing.T) {
	mockProvider := new(MockDockerProvider)
	ctx := context.Background()

	config := ComposeConfig{
		ProjectName: "test-project",
		Network:     "test-network",
		Services: map[string]ServiceConfig{
			"app": {
				ImageName: "test-image",
				ImageTag:  "latest",
				ExposedPorts: []PortMapping{
					{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
				},
				Environment: map[string]string{
					"ENV_VAR": "value",
				},
			},
		},
	}

	mockProvider.On("Initialize", ctx, config).Return(nil)

	err := mockProvider.Initialize(ctx, config)

	assert.NoError(t, err)
	mockProvider.AssertExpectations(t)
}

func TestDockerProviderStartAndStop(t *testing.T) {
	mockProvider := new(MockDockerProvider)
	ctx := context.Background()

	mockProvider.On("Start", ctx).Return(nil)
	mockProvider.On("Stop", ctx).Return(nil)

	err := mockProvider.Start(ctx)
	assert.NoError(t, err)

	err = mockProvider.Stop(ctx)
	assert.NoError(t, err)

	mockProvider.AssertExpectations(t)
}

func TestDockerProviderStatus(t *testing.T) {
	mockProvider := new(MockDockerProvider)
	ctx := context.Background()

	expectedStatus := map[string]string{
		"app": "running",
		"db":  "running",
	}

	mockProvider.On("Status", ctx).Return(expectedStatus, nil)

	status, err := mockProvider.Status(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
	mockProvider.AssertExpectations(t)
}

func TestDockerProviderGetLogs(t *testing.T) {
	mockProvider := new(MockDockerProvider)
	ctx := context.Background()
	serviceName := "app"

	expectedLogs := "Container logs output"
	mockProvider.On("GetLogs", ctx, serviceName).Return(strings.NewReader(expectedLogs), nil)

	logs, err := mockProvider.GetLogs(ctx, serviceName)

	assert.NoError(t, err)

	// Read the logs
	logsBytes, err := io.ReadAll(logs)
	assert.NoError(t, err)
	assert.Equal(t, expectedLogs, string(logsBytes))

	mockProvider.AssertExpectations(t)
}

func TestDockerProviderGetContainerID(t *testing.T) {
	mockProvider := new(MockDockerProvider)
	serviceName := "app"

	expectedID := "container123"
	mockProvider.On("GetContainerID", serviceName).Return(expectedID)

	id := mockProvider.GetContainerID(serviceName)

	assert.Equal(t, expectedID, id)
	mockProvider.AssertExpectations(t)
}

func TestDockerProviderGetServices(t *testing.T) {
	mockProvider := new(MockDockerProvider)

	expectedServices := []string{"app", "db"}
	mockProvider.On("GetServices").Return(expectedServices)

	services := mockProvider.GetServices()

	assert.Equal(t, expectedServices, services)
	mockProvider.AssertExpectations(t)
}

func TestServiceConfig(t *testing.T) {
	// Test creating a valid service config
	serviceConfig := ServiceConfig{
		ImageName: "nginx",
		ImageTag:  "latest",
		ExposedPorts: []PortMapping{
			{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
		},
		Environment: map[string]string{
			"ENV_VAR1": "value1",
			"ENV_VAR2": "value2",
		},
		Volumes: []VolumeMapping{
			{HostPath: "/host/path", ContainerPath: "/container/path"},
		},
		DependsOn:     []string{"db"},
		RestartPolicy: "always",
		Resources: ResourceLimits{
			Memory:   "512m",
			CPUShare: "0.5",
		},
	}

	assert.Equal(t, "nginx", serviceConfig.ImageName)
	assert.Equal(t, "latest", serviceConfig.ImageTag)
	assert.Len(t, serviceConfig.ExposedPorts, 1)
	assert.Equal(t, 8080, serviceConfig.ExposedPorts[0].HostPort)
	assert.Equal(t, 80, serviceConfig.ExposedPorts[0].ContainerPort)
	assert.Equal(t, "tcp", serviceConfig.ExposedPorts[0].Protocol)
	assert.Len(t, serviceConfig.Environment, 2)
	assert.Equal(t, "value1", serviceConfig.Environment["ENV_VAR1"])
	assert.Equal(t, "value2", serviceConfig.Environment["ENV_VAR2"])
	assert.Len(t, serviceConfig.Volumes, 1)
	assert.Equal(t, "/host/path", serviceConfig.Volumes[0].HostPath)
	assert.Equal(t, "/container/path", serviceConfig.Volumes[0].ContainerPath)
	assert.Len(t, serviceConfig.DependsOn, 1)
	assert.Equal(t, "db", serviceConfig.DependsOn[0])
	assert.Equal(t, "always", serviceConfig.RestartPolicy)
	assert.Equal(t, "512m", serviceConfig.Resources.Memory)
	assert.Equal(t, "0.5", serviceConfig.Resources.CPUShare)
}

func TestComposeConfig(t *testing.T) {
	// Test creating a valid compose config
	composeConfig := ComposeConfig{
		ProjectName: "test-project",
		Network:     "test-network",
		EnvFile:     ".env",
		Services: map[string]ServiceConfig{
			"app": {
				ImageName: "app-image",
				ImageTag:  "latest",
			},
			"db": {
				ImageName: "postgres",
				ImageTag:  "13",
			},
		},
	}

	assert.Equal(t, "test-project", composeConfig.ProjectName)
	assert.Equal(t, "test-network", composeConfig.Network)
	assert.Equal(t, ".env", composeConfig.EnvFile)
	assert.Len(t, composeConfig.Services, 2)
	assert.Contains(t, composeConfig.Services, "app")
	assert.Contains(t, composeConfig.Services, "db")
	assert.Equal(t, "app-image", composeConfig.Services["app"].ImageName)
	assert.Equal(t, "latest", composeConfig.Services["app"].ImageTag)
	assert.Equal(t, "postgres", composeConfig.Services["db"].ImageName)
	assert.Equal(t, "13", composeConfig.Services["db"].ImageTag)
}
