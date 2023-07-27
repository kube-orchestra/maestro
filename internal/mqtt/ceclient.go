package mqtt

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kube-orchestra/maestro/internal/db"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	workv1 "open-cluster-management.io/api/work/v1"
	"open-cluster-management.io/work/pkg/clients/mqclient"
	"open-cluster-management.io/work/pkg/clients/workclient/payload"
	"strconv"
	"time"
)

type Encoder struct{}

var _ mqclient.Encoder[*db.Resource] = &Encoder{}

func (e *Encoder) EncodeSpec(eventType, source string, obj *db.Resource) (*cloudevents.Event, error) {
	evt := cloudevents.NewEvent()
	evt.SetID(uuid.New().String())
	evt.SetSource(source)
	evt.SetType(eventType)
	evt.SetTime(time.Now())
	evt.SetDataContentType("application/json")
	evt.SetExtension("resourceID", obj.Id)
	evt.SetExtension("resourceVersion", obj.ResourceGenerationID)

	if !obj.Object.GetDeletionTimestamp().IsZero() {
		evt.SetExtension("deletionTimestamp", obj.Object.GetDeletionTimestamp().Time)
		return &evt, nil
	}

	resourcePayload := &payload.Resource{
		Manifest: &obj.Object,
	}

	resourcePayloadJSON, err := json.Marshal(resourcePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource payload: %v", err)
	}

	if err := evt.SetData(cloudevents.ApplicationJSON, resourcePayloadJSON); err != nil {
		return nil, fmt.Errorf("failed to encode resource to cloud event: %v", err)
	}

	return &evt, nil
}

func (e *Encoder) EncodeStatus(eventType, source string, obj *db.Resource) (*cloudevents.Event, error) {
	evt := cloudevents.NewEvent()
	evt.SetID(uuid.New().String())
	evt.SetSource(source)
	evt.SetType(eventType)
	evt.SetTime(time.Now())
	evt.SetDataContentType("application/json")

	evt.SetExtension("resourceID", obj.Id)
	evt.SetExtension("resourceVersion", obj.ResourceGenerationID)

	rawStatus, err := json.Marshal(obj.Status.ContentStatus)
	if err != nil {
		return nil, err
	}
	resourceStatusPayload := &payload.ResourceStatus{
		ReconcileStatus: payload.ReconcileStatus{
			Conditions: obj.Status.ReconcileStatus.Conditions,
		},
		ResourceMeta: workv1.ManifestResourceMeta{
			Group:     obj.Object.GroupVersionKind().Group,
			Version:   obj.Object.GroupVersionKind().Version,
			Kind:      obj.Object.GroupVersionKind().Kind,
			Name:      obj.Object.GetName(),
			Namespace: obj.Object.GetNamespace(),
		},
		ContentStatus: &runtime.RawExtension{Raw: rawStatus},
	}
	resourceStatusPayloadJSON, err := json.Marshal(resourceStatusPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource status payload: %v", err)
	}

	if err := evt.SetData(cloudevents.ApplicationJSON, resourceStatusPayloadJSON); err != nil {
		return nil, fmt.Errorf("failed to encode manifests status to cloud event: %v", err)
	}

	return &evt, nil
}

type Decoder struct{}

var _ mqclient.Decoder[*db.Resource] = &Decoder{}

func (e *Decoder) DecodeSpec(evt *cloudevents.Event) (*db.Resource, error) {
	resourceID, err := evt.Context.GetExtension("resourceID")
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceID extension: %v", err)
	}

	resourceIDStr, ok := resourceID.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert resourceID - %v to string", resourceID)
	}

	resourceVersion, err := evt.Context.GetExtension("resourceVersion")
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceVersion extension: %v", err)
	}

	resourceVersionStr, ok := resourceVersion.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert resourceVersion - %v to string", resourceVersion)
	}

	resourceVersionInt, err := strconv.ParseInt(resourceVersionStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resourceVersion - %v to int64", resourceVersion)
	}

	data := evt.Data()
	resourcePayload := payload.Resource{}
	if err := json.Unmarshal(data, &resourcePayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data as resource: %v", err)
	}

	resource := &db.Resource{
		Id:                   resourceIDStr,
		ResourceGenerationID: resourceVersionInt,
		Object:               *resourcePayload.Manifest,
	}

	deletionTimestamp, err := evt.Context.GetExtension("deletionTimestamp")
	if err == nil {
		deletionTimestampStr, ok := deletionTimestamp.(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert deletionTimestamp - %v to time.Time", deletionTimestamp)
		}
		deletionTimestampVal, err := time.Parse(time.RFC3339, deletionTimestampStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse deletionTimestamp - %v to time.Time", deletionTimestamp)
		}
		resource.Object.SetDeletionTimestamp(&metav1.Time{Time: deletionTimestampVal})
	} else {
		return nil, err
	}

	return resource, nil
}

func (e *Decoder) DecodeStatus(evt *cloudevents.Event) (*db.Resource, error) {
	resourceID, err := evt.Context.GetExtension("resourceID")
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceID extension: %v", err)
	}

	resourceIDStr, ok := resourceID.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert resourceID - %v to string", resourceID)
	}

	resourceVersion, err := evt.Context.GetExtension("resourceVersion")
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceVersion extension: %v", err)
	}

	resourceVersionStr, ok := resourceVersion.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert resourceVersion - %v to string", resourceVersion)
	}

	resourceVersionInt, err := strconv.ParseInt(resourceVersionStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resourceVersion - %v to int64", resourceVersion)
	}

	data := evt.Data()
	resourceStatusPayload := payload.ResourceStatus{}
	if err := json.Unmarshal(data, &resourceStatusPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data as resource status: %v", err)
	}

	unsObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resourceStatusPayload.ContentStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to convert manifest to unstructured object: %v", err)
	}

	resource := &db.Resource{
		Id:                   resourceIDStr,
		ResourceGenerationID: resourceVersionInt,
		Status: db.StatusMessage{
			MessageMeta: db.MessageMeta{
				ResourceGenerationID: resourceVersionInt,
			},
			ReconcileStatus: db.ReconcileStatus{
				Conditions: resourceStatusPayload.ReconcileStatus.Conditions,
			},
			ContentStatus: unsObj,
		},
	}

	return resource, nil
}

type ResourceLister struct{}

// not implemented.
func (r *ResourceLister) List() []*db.Resource {
	return nil
}

type ResourceStatusHashGetter struct{}

var _ mqclient.StatusHashGetter[*db.Resource] = &ResourceStatusHashGetter{}

func (r *ResourceStatusHashGetter) Get(obj *db.Resource) (string, error) {
	statusBytes, err := json.Marshal(obj.Status)
	if err != nil {
		return "", fmt.Errorf("failed to marshal work status, %v", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(statusBytes)), nil
}
