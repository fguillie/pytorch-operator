package controller

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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pytorchv1alpha1 "github.com/pytorch-operator/pytorch-operator/api/v1alpha1"
)

const defaultBaseImage = "nvcr.io/nvidia/pytorch"

// PyTorchJobReconciler reconciles a PyTorchJob object.
type PyTorchJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pytorch.nvidia.com,resources=pytorchjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pytorch.nvidia.com,resources=pytorchjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pytorch.nvidia.com,resources=pytorchjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *PyTorchJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	job := &pytorchv1alpha1.PyTorchJob{}
	if err := r.Get(ctx, req.NamespacedName, job); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	desired := r.buildDeployment(job)
	if err := ctrl.SetControllerReference(job, desired, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	existing := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
	if errors.IsNotFound(err) {
		logger.Info("Creating Deployment", "name", desired.Name)
		if err := r.Create(ctx, desired); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, r.patchPhase(ctx, job, "Pending")
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile Deployment spec (image, replicas, GPU count may have changed).
	patch := client.MergeFrom(existing.DeepCopy())
	existing.Spec = desired.Spec
	if err := r.Patch(ctx, existing, patch); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.syncStatus(ctx, job, existing)
}

func (r *PyTorchJobReconciler) buildDeployment(job *pytorchv1alpha1.PyTorchJob) *appsv1.Deployment {
	replicas := int32(1)
	if job.Spec.Replicas != nil {
		replicas = *job.Spec.Replicas
	}

	baseImage := defaultBaseImage
	if job.Spec.Image != "" {
		baseImage = job.Spec.Image
	}
	image := fmt.Sprintf("%s:%s", baseImage, job.Spec.PytorchVersion)

	gpuQty := resource.MustParse(fmt.Sprintf("%d", job.Spec.GPUCount))

	labels := map[string]string{
		"app":                         job.Name,
		"pytorch.nvidia.com/job-name": job.Name,
	}

	container := corev1.Container{
		Name:    "pytorch",
		Image:   image,
		Command: job.Spec.Command,
		Args:    job.Spec.Args,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"nvidia.com/gpu": gpuQty,
			},
			Requests: corev1.ResourceList{
				"nvidia.com/gpu": gpuQty,
			},
		},
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.Name,
			Namespace: job.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{container}},
			},
		},
	}
}

func (r *PyTorchJobReconciler) syncStatus(ctx context.Context, job *pytorchv1alpha1.PyTorchJob, dep *appsv1.Deployment) error {
	phase := "Pending"
	switch {
	case dep.Status.ReadyReplicas > 0 && dep.Status.ReadyReplicas == dep.Status.Replicas:
		phase = "Ready"
	case dep.Status.ReadyReplicas > 0:
		phase = "Running"
	case dep.Status.UnavailableReplicas > 0:
		phase = "Pending"
	}

	patch := client.MergeFrom(job.DeepCopy())
	job.Status.ReadyReplicas = dep.Status.ReadyReplicas
	job.Status.Phase = phase
	return r.Status().Patch(ctx, job, patch)
}

func (r *PyTorchJobReconciler) patchPhase(ctx context.Context, job *pytorchv1alpha1.PyTorchJob, phase string) error {
	patch := client.MergeFrom(job.DeepCopy())
	job.Status.Phase = phase
	return r.Status().Patch(ctx, job, patch)
}

// SetupWithManager registers the controller with the manager and watches owned Deployments.
func (r *PyTorchJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pytorchv1alpha1.PyTorchJob{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
