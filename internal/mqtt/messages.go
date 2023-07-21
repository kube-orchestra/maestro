package mqtt

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceMessage struct {
	MessageMeta `json:",inline"`

	// Kubernetes Manifest to apply on the target.
	Content *unstructured.Unstructured `json:"content"`
}

type StatusMessage struct {
	MessageMeta `json:",inline"`
	// agent status information.
	ReconcileStatus ReconcileStatus `json:"reconcileStatus"`
	// content status as observed on the target.
	ContentStatus *runtime.RawExtension `json:"contentStatus"`
}

const (
	// Reconciled condition tracks the state of the reconcile operation.
	// "True" indicates that the object has been successfully applied.
	// "False" indicates a non-transitive error prevents reconciliation from succeeding.
	StatusMessageReconciled = "Reconciled"
	// Deleted condition tracks the if deletion succeeded.
	// "True" indicates that the object has been successfully removed from the target.
	// "False" indicates that deletion is blocked or did not yet succeed.
	StatusMessageDeleted = "Deleted"
)

type ReconcileStatus struct {
	// MAY when object exists/
	// Object generation as observed on the target.
	// .metadata.generation
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// MAY when object exists.
	// RFC3339 Timestamp.
	// .metadata.creationTimestamp as observed on the target.
	CreationTimestamp *metav1.Time `json:"creationTimestamp,omitempty"`
	// Kubernetes style status conditions,
	// describing the state of the object on the target.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type MessageMeta struct {
	Id         string `json:"-"`
	ConsumerId string `json:"-"`

	// Unix Timestamp (UTC) at which time
	// the message was sent to the broker.
	SentTimestamp int64 `json:"sentTimestamp"`

	// Server-side opaque corelation ID.
	// MUST be passed back in status responses unchanged.
	ResourceGenerationID string `json:"resourceGenerationID"`
}
