package controller

import (
	"context"
    "fmt"
    "github.com/go-logr/logr"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    ctrl "sigs.k8s.io/controller-runtime"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    appsv1alpha1 "github.com/NikilLepcha/bg-deployment-operator/api/v1alpha1"
)

// BGDeploymentOperatorReconciler reconciles a BGDeploymentOperator object
type BGDeploymentOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log logr.Logger
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *BGDeploymentOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("bluegreendeployment", req.NamespacedName)

	// Fetch the BlueGreenDeployment instance
	bgd := &appsv1alpha1.BGDeploymentOperator{}
	err := r.Get(ctx, req.NamespacedName, bgd)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Determine active color
	activeColor := "blue"
	if bgd.Status.ActiveColor == "blue" {
		activeColor = "green"
	}

	// Create or update the deployment
	newDeploymentName := fmt.Sprintf("%s-%s", bgd.Name, activeColor)
	newDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: newDeploymentName, Namespace: bgd.Namespace}, newDeployment)
	if err != nil && errors.IsNotFound(err) {
		newDeployment = r.constructDeployment(newDeploymentName, bgd)
		err = r.Create(ctx, newDeployment)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		updated := r.updateDeployment(newDeployment, bgd)
		if updated {
			err = r.Update(ctx, newDeployment)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Create or update the service to point to the new deployment
	service := &corev1.Service{}
    err = r.Get(ctx, types.NamespacedName{Name: bgd.Name, Namespace: bgd.Namespace}, service)
    if err != nil && errors.IsNotFound(err) {
        service = r.constructService(bgd)
        err = r.Create(ctx, service)
        if err != nil {
            return ctrl.Result{}, err
        }
    } else if err != nil {
        return ctrl.Result{}, err
    } else {
        updated := r.updateService(service, activeColor)
        if updated {
            err = r.Update(ctx, service)
            if err != nil {
                return ctrl.Result{}, err
            }
        }
    }

	// Update the status to reflect the new active color
    bgd.Status.ActiveColor = activeColor
    err = r.Status().Update(ctx, bgd)
    if err != nil {
        return ctrl.Result{}, err
    }

	return ctrl.Result{}, nil
}

func (r *BGDeploymentOperatorReconciler) constructDeployment(name string, bgd *appsv1alpha1.BGDeploymentOperator) *appsv1.Deployment {
	labels := map[string]string {
		"app": bgd.Name,
		"color": name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: bgd.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &bgd.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name: "app",
						Image: bgd.Spec.Image,
						Ports: []corev1.ContainerPort{{
								ContainerPort: 80,
							}},
					}},
				},
			},
		},
	}
}

func (r *BGDeploymentOperatorReconciler) updateDeployment(deployment *appsv1.Deployment, bgd *appsv1alpha1.BGDeploymentOperator) bool {
	updated := false
	if *deployment.Spec.Replicas != bgd.Spec.Replicas {
		deployment.Spec.Replicas = &bgd.Spec.Replicas
		updated = true
	}
	if deployment.Spec.Template.Spec.Containers[0].Image != bgd.Spec.Image {
		deployment.Spec.Template.Spec.Containers[0].Image = bgd.Spec.Image
		updated = true
	}
	return updated
}

func (r *BGDeploymentOperatorReconciler) constructService(bgd *appsv1alpha1.BGDeploymentOperator) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: bgd.Name,
			Namespace: bgd.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": bgd.Name,
				"color": bgd.Status.ActiveColor,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Protocol: corev1.ProtocolTCP,
					Port: 80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}
}

func (r *BGDeploymentOperatorReconciler) updateService(service *corev1.Service, activeColor string) bool {
	updated := false
    if service.Spec.Selector["color"] != activeColor {
        service.Spec.Selector["color"] = activeColor
        updated = true
    }
    return updated
}

// SetupWithManager sets up the controller with the Manager.
func (r *BGDeploymentOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.BGDeploymentOperator{}).
		Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Complete(r)
}
