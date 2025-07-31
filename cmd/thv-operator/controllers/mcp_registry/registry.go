package mcp_registry

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	controllertypes "github.com/stacklok/toolhive/cmd/thv-operator/controllers/types"
)

// CreateRegistryResources creates the registry deployment and service
func CreateRegistryResources(ctx context.Context, client client.Client, scheme *runtime.Scheme, owner metav1.Object, registryName, namespace, mongoSvcName string) error {
	registryLabels := map[string]string{controllertypes.AppLabel: registryName, controllertypes.ComponentLabel: controllertypes.RegistryComponent}

	// Create registry deployment
	registryDeploymentName := registryName + "-registry"
	registryDeployment := &appsv1.Deployment{}
	if err := client.Get(ctx, types.NamespacedName{Name: registryDeploymentName, Namespace: namespace}, registryDeployment); apierrors.IsNotFound(err) {
		registryDeployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registryDeploymentName,
				Namespace: namespace,
				Labels:    registryLabels,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{MatchLabels: registryLabels},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: registryLabels},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "mcp-registry",
							Image: controllertypes.RegistryImage,
							Env: []corev1.EnvVar{
								{Name: "MCP_REGISTRY_DATABASE_TYPE", Value: controllertypes.MongoDBDatabaseType},
								{Name: "MCP_REGISTRY_DATABASE_URL", Value: fmt.Sprintf("mongodb://%s:%s", mongoSvcName, controllertypes.MongoDBPort)},
								{Name: "MCP_REGISTRY_DATABASE_NAME", Value: controllertypes.MongoDBDatabaseName},
								{Name: "MCP_REGISTRY_COLLECTION_NAME", Value: controllertypes.MongoDBCollectionName},
								{Name: "MCP_REGISTRY_LOG_LEVEL", Value: controllertypes.MongoDBLogLevel},
								{Name: "MCP_REGISTRY_SEED_IMPORT", Value: controllertypes.MongoDBSeedImport},
							},
							Ports: []corev1.ContainerPort{{ContainerPort: controllertypes.RegistryPort}},
						}},
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(owner, registryDeployment, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on registry deployment: %w", err)
		}
		if err := client.Create(ctx, registryDeployment); err != nil {
			return fmt.Errorf("failed to create registry deployment: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Create registry service
	registrySvcName := registryName + "-registry"
	regSvc := &corev1.Service{}
	if err := client.Get(ctx, types.NamespacedName{Name: registrySvcName, Namespace: namespace}, regSvc); apierrors.IsNotFound(err) {
		regSvc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registrySvcName,
				Namespace: namespace,
				Labels:    registryLabels,
			},
			Spec: corev1.ServiceSpec{
				Selector: registryLabels,
				Ports: []corev1.ServicePort{{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				}},
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		if err := controllerutil.SetControllerReference(owner, regSvc, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on registry service: %w", err)
		}
		if err := client.Create(ctx, regSvc); err != nil {
			return fmt.Errorf("failed to create registry service: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
