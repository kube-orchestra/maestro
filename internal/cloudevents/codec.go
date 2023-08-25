package cloudevents

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kube-orchestra/maestro/internal/db"
	"k8s.io/apimachinery/pkg/runtime"
	ceclient "open-cluster-management.io/api/client/cloudevents"
	"open-cluster-management.io/api/client/cloudevents/types"
)

type Codec struct{}

var _ ceclient.Codec[*db.Resource] = &Codec{}

func (codec *Codec) EventDataType() types.CloudEventsDataType {
	return types.CloudEventsDataType{
		Group:    "io.open-cluster-management.works",
		Version:  "v1alpha1",
		Resource: "manifest",
	}
}

func (codec *Codec) Encode(source string, eventType types.CloudEventsType, obj *db.Resource) (*cloudevents.Event, error) {
	evt := cloudevents.NewEvent()
	evt.SetID(uuid.New().String())
	evt.SetSource(source)
	evt.SetType(eventType.String())
	evt.SetTime(time.Now())
	evt.SetDataContentType("application/json")
	evt.SetExtension("resourceID", obj.Id)
	evt.SetExtension("resourceVersion", obj.ResourceGenerationID)
	evt.SetExtension("clustername", obj.ConsumerId)

	if !obj.Object.GetDeletionTimestamp().IsZero() {
		evt.SetExtension("deletionTimestamp", obj.Object.GetDeletionTimestamp().Time)
		return &evt, nil
	}

	resourcePayload := &Resource{
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

func (codec *Codec) Decode(evt *cloudevents.Event) (*db.Resource, error) {
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
	resourceStatusPayload := ResourceStatus{}
	if err := json.Unmarshal(data, &resourceStatusPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data as resource status: %v", err)
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
		},
	}

	if resourceStatusPayload.ContentStatus != nil {
		unsObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resourceStatusPayload.ContentStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to convert manifest to unstructured object: %v", err)
		}
		resource.Status.ContentStatus = unsObj
	}

	return resource, nil
}
