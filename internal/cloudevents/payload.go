package cloudevents

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	workv1 "open-cluster-management.io/api/work/v1"
)

type Resource struct {
	// Kubernetes Manifest to apply on the target.
	Manifest *unstructured.Unstructured `json:"manifest"`
}

type ResourceStatus struct {
	// work agent reconcile status information.
	ReconcileStatus ReconcileStatus `json:"reconcileStatus"`
	// resource meta information.
	ResourceMeta workv1.ManifestResourceMeta `json:"resourceMeta"`
	// content status as observed on the target.
	ContentStatus *runtime.RawExtension `json:"contentStatus"`
}

type ReconcileStatus struct {
	// Kubernetes style status conditions,
	// describing the state of the object on the target.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
