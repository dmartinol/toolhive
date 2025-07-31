package mcp_registry

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	controllertypes "github.com/stacklok/toolhive/cmd/thv-operator/controllers/types"
)

// CreateMongoDBResources creates MongoDB deployment, service, and PVC
func CreateMongoDBResources(ctx context.Context, client client.Client, scheme *runtime.Scheme, owner metav1.Object, name, namespace string) error {
	mongoLabels := map[string]string{controllertypes.AppLabel: name, controllertypes.ComponentLabel: controllertypes.MongoDBComponent}

	// Create MongoDB PVC
	mongoPVCName := name + "-mongodb-pvc"
	mongoPVC := &corev1.PersistentVolumeClaim{}
	if err := client.Get(ctx, types.NamespacedName{Name: mongoPVCName, Namespace: namespace}, mongoPVC); apierrors.IsNotFound(err) {
		mongoPVC = &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mongoPVCName,
				Namespace: namespace,
				Labels:    mongoLabels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("500Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("500Mi"),
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(owner, mongoPVC, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on mongodb pvc: %w", err)
		}
		if err := client.Create(ctx, mongoPVC); err != nil {
			return fmt.Errorf("failed to create mongodb pvc: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Create MongoDB Deployment
	mongoDepName := name + "-mongodb"
	mongoDep := &appsv1.Deployment{}
	if err := client.Get(ctx, types.NamespacedName{Name: mongoDepName, Namespace: namespace}, mongoDep); apierrors.IsNotFound(err) {
		mongoDep = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mongoDepName,
				Namespace: namespace,
				Labels:    mongoLabels,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{MatchLabels: mongoLabels},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: mongoLabels},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "mongodb",
							Image: "mongo:5.0",
							Ports: []corev1.ContainerPort{{ContainerPort: 27017}},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "mongodb-data",
								MountPath: "/data/db",
							}},
						}},
						Volumes: []corev1.Volume{{
							Name: "mongodb-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: mongoPVCName,
								},
							},
						}},
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(owner, mongoDep, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on mongodb deployment: %w", err)
		}
		if err := client.Create(ctx, mongoDep); err != nil {
			return fmt.Errorf("failed to create mongodb deployment: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Create MongoDB Service
	mongoSvcName := GetMongoServiceName(name)
	mongoSvc := &corev1.Service{}
	if err := client.Get(ctx, types.NamespacedName{Name: mongoSvcName, Namespace: namespace}, mongoSvc); apierrors.IsNotFound(err) {
		mongoSvc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mongoSvcName,
				Namespace: namespace,
				Labels:    mongoLabels,
			},
			Spec: corev1.ServiceSpec{
				Selector: mongoLabels,
				Ports: []corev1.ServicePort{{
					Name:       "mongodb",
					Port:       27017,
					TargetPort: intstr.FromInt(27017),
				}},
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		if err := controllerutil.SetControllerReference(owner, mongoSvc, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on mongodb service: %w", err)
		}
		if err := client.Create(ctx, mongoSvc); err != nil {
			return fmt.Errorf("failed to create mongodb service: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

// GetMongoServiceName returns the MongoDB service name for a given registry name
func GetMongoServiceName(registryName string) string {
	return registryName + "-mongodb"
}

// Helper function to create int32 pointer
func int32Ptr(i int32) *int32 {
	return &i
}
