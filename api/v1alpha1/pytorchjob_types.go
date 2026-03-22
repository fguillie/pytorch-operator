package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PyTorchJobSpec defines the desired state of PyTorchJob.
type PyTorchJobSpec struct {
	// PytorchVersion is the NVIDIA PyTorch container version tag (e.g. "24.01-py3", "2.3.0").
	// The final image is constructed as <image>:<pytorchVersion>.
	// +kubebuilder:validation:Required
	PytorchVersion string `json:"pytorchVersion"`

	// GPUCount is the number of NVIDIA GPUs (nvidia.com/gpu) to allocate per pod.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	GPUCount int32 `json:"gpuCount"`

	// Replicas is the number of pods to run (default: 1).
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Image overrides the base container image (default: nvcr.io/nvidia/pytorch).
	// The pytorchVersion is appended as the tag: <image>:<pytorchVersion>.
	// +optional
	Image string `json:"image,omitempty"`

	// Command overrides the container entrypoint.
	// +optional
	Command []string `json:"command,omitempty"`

	// Args are arguments passed to the container entrypoint.
	// +optional
	Args []string `json:"args,omitempty"`
}

// PyTorchJobStatus defines the observed state of PyTorchJob.
type PyTorchJobStatus struct {
	// Phase is the current lifecycle phase: Pending, Running, Ready, or Failed.
	// +optional
	Phase string `json:"phase,omitempty"`

	// ReadyReplicas is the number of pods currently ready.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Conditions holds the latest available observations of the job's state.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=ptjob,scope=Namespaced
//+kubebuilder:printcolumn:name="PyTorch Version",type=string,JSONPath=`.spec.pytorchVersion`
//+kubebuilder:printcolumn:name="GPUs",type=integer,JSONPath=`.spec.gpuCount`
//+kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
//+kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// PyTorchJob manages NVIDIA PyTorch containers with a specified version and GPU allocation.
type PyTorchJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PyTorchJobSpec   `json:"spec,omitempty"`
	Status PyTorchJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PyTorchJobList contains a list of PyTorchJob.
type PyTorchJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PyTorchJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PyTorchJob{}, &PyTorchJobList{})
}
