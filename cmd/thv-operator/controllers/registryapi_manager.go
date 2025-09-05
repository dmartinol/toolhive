package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
)

const (
	// RegistryAPIPort is the port the Registry API listens on
	RegistryAPIPort = 8080
	
	// RegistryAPIImage is the container image for the Registry API
	RegistryAPIImage = "thv-registry-api:latest"
	
	// RegistryAPIServicePort is the port exposed by the Service
	RegistryAPIServicePort = 80
)

// RegistryAPIManager manages Registry API deployments for MCPRegistry resources
type RegistryAPIManager struct {
	client client.Client
	scheme *runtime.Scheme
}

// NewRegistryAPIManager creates a new Registry API manager
func NewRegistryAPIManager(client client.Client, scheme *runtime.Scheme) *RegistryAPIManager {
	return &RegistryAPIManager{
		client: client,
		scheme: scheme,
	}
}

// ReconcileRegistryAPI ensures the Registry API deployment and service exist for the given registry
func (m *RegistryAPIManager) ReconcileRegistryAPI(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) (string, error) {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name, "component", "registry-api")
	
	// Create or update the Deployment
	deployment := m.buildDeployment(registry)
	if err := controllerutil.SetControllerReference(registry, deployment, m.scheme); err != nil {
		return "", fmt.Errorf("failed to set controller reference on deployment: %w", err)
	}
	
	if err := m.reconcileDeployment(ctx, deployment); err != nil {
		return "", fmt.Errorf("failed to reconcile deployment: %w", err)
	}
	
	// Create or update the Service
	service := m.buildService(registry)
	if err := controllerutil.SetControllerReference(registry, service, m.scheme); err != nil {
		return "", fmt.Errorf("failed to set controller reference on service: %w", err)
	}
	
	if err := m.reconcileService(ctx, service); err != nil {
		return "", fmt.Errorf("failed to reconcile service: %w", err)
	}
	
	// Generate the API endpoint URL
	apiEndpoint := m.buildAPIEndpoint(service)
	logger.Info("Registry API reconciled successfully", "endpoint", apiEndpoint)
	
	return apiEndpoint, nil
}

// buildDeployment creates a Deployment for the Registry API
func (m *RegistryAPIManager) buildDeployment(registry *mcpv1alpha1.MCPRegistry) *appsv1.Deployment {
	labels := m.buildLabels(registry)
	replicas := int32(1)
	
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.buildResourceName(registry, "api"),
			Namespace: registry.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "registry-api",
							Image: RegistryAPIImage,
							Args: []string{
								fmt.Sprintf("--port=%d", RegistryAPIPort),
								fmt.Sprintf("--registry-name=%s", registry.Name),
								fmt.Sprintf("--registry-namespace=%s", registry.Namespace),
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: RegistryAPIPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/readiness",
										Port: intstr.FromInt(RegistryAPIPort),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(RegistryAPIPort),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       20,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    mustParseQuantity("100m"),
									corev1.ResourceMemory: mustParseQuantity("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    mustParseQuantity("500m"),
									corev1.ResourceMemory: mustParseQuantity("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}

// buildService creates a Service for the Registry API
func (m *RegistryAPIManager) buildService(registry *mcpv1alpha1.MCPRegistry) *corev1.Service {
	labels := m.buildLabels(registry)
	
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.buildResourceName(registry, "api"),
			Namespace: registry.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       RegistryAPIServicePort,
					TargetPort: intstr.FromInt(RegistryAPIPort),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// reconcileDeployment ensures the Deployment exists and is up to date
func (m *RegistryAPIManager) reconcileDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	existing := &appsv1.Deployment{}
	err := m.client.Get(ctx, types.NamespacedName{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
	}, existing)
	
	if err != nil {
		if errors.IsNotFound(err) {
			// Create the deployment
			return m.client.Create(ctx, deployment)
		}
		return err
	}
	
	// Update the existing deployment
	existing.Spec = deployment.Spec
	return m.client.Update(ctx, existing)
}

// reconcileService ensures the Service exists and is up to date
func (m *RegistryAPIManager) reconcileService(ctx context.Context, service *corev1.Service) error {
	existing := &corev1.Service{}
	err := m.client.Get(ctx, types.NamespacedName{
		Name:      service.Name,
		Namespace: service.Namespace,
	}, existing)
	
	if err != nil {
		if errors.IsNotFound(err) {
			// Create the service
			return m.client.Create(ctx, service)
		}
		return err
	}
	
	// Update the existing service (preserve ClusterIP)
	existing.Spec.Selector = service.Spec.Selector
	existing.Spec.Ports = service.Spec.Ports
	existing.Spec.Type = service.Spec.Type
	return m.client.Update(ctx, existing)
}

// buildLabels creates standard labels for Registry API resources
func (m *RegistryAPIManager) buildLabels(registry *mcpv1alpha1.MCPRegistry) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "registry-api",
		"app.kubernetes.io/instance":   registry.Name,
		"app.kubernetes.io/component":  "api",
		"app.kubernetes.io/part-of":    "toolhive",
		"app.kubernetes.io/managed-by": "thv-operator",
		"toolhive.stacklok.dev/registry": registry.Name,
	}
}

// buildResourceName creates a consistent resource name for Registry API components
func (m *RegistryAPIManager) buildResourceName(registry *mcpv1alpha1.MCPRegistry, component string) string {
	return fmt.Sprintf("%s-%s", registry.Name, component)
}

// buildAPIEndpoint constructs the API endpoint URL for the service
func (m *RegistryAPIManager) buildAPIEndpoint(service *corev1.Service) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", 
		service.Name, 
		service.Namespace, 
		RegistryAPIServicePort)
}

// DeleteRegistryAPI removes the Registry API deployment and service
func (m *RegistryAPIManager) DeleteRegistryAPI(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) error {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name, "component", "registry-api")
	
	// Delete deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.buildResourceName(registry, "api"),
			Namespace: registry.Namespace,
		},
	}
	if err := m.client.Delete(ctx, deployment); err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to delete Registry API deployment")
		return err
	}
	
	// Delete service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.buildResourceName(registry, "api"),
			Namespace: registry.Namespace,
		},
	}
	if err := m.client.Delete(ctx, service); err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to delete Registry API service")
		return err
	}
	
	logger.Info("Registry API resources deleted successfully")
	return nil
}

// mustParseQuantity is a helper function that panics on parse errors
func mustParseQuantity(s string) resource.Quantity {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse quantity %s: %v", s, err))
	}
	return q
}