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
	cegeneric "open-cluster-management.io/api/cloudevents/generic"
	cetypes "open-cluster-management.io/api/cloudevents/generic/types"
	workpayload "open-cluster-management.io/api/cloudevents/work/payload"
)

type Codec struct{}

var _ cegeneric.Codec[*db.Resource] = &Codec{}

func (codec *Codec) EventDataType() cetypes.CloudEventsDataType {
	return workpayload.ManifestEventDataType
}

func (codec *Codec) Encode(source string, eventType cetypes.CloudEventsType, obj *db.Resource) (*cloudevents.Event, error) {
	evt := cloudevents.NewEvent()
	evt.SetID(uuid.New().String())
	evt.SetSource(source)
	evt.SetType(eventType.String())
	evt.SetTime(time.Now())
	evt.SetDataContentType("application/json")
	evt.SetExtension("resourceid", obj.Id)
	evt.SetExtension("resourceversion", obj.ResourceGenerationID)
	evt.SetExtension("clustername", obj.ConsumerId)

	if !obj.Object.GetDeletionTimestamp().IsZero() {
		evt.SetExtension("deletionTimestamp", obj.Object.GetDeletionTimestamp().Time)
		return &evt, nil
	}

	resourcePayload := &workpayload.Manifest{
		Manifest: obj.Object,
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
	resourceID, err := evt.Context.GetExtension("resourceid")
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceid extension: %v", err)
	}

	resourceIDStr, ok := resourceID.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert resourceid - %v to string", resourceID)
	}

	resourceVersion, err := evt.Context.GetExtension("resourceversion")
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceversion extension: %v", err)
	}

	resourceVersionStr, ok := resourceVersion.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert resourceversion - %v to string", resourceVersion)
	}

	resourceVersionInt, err := strconv.ParseInt(resourceVersionStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resourceversion - %v to int64", resourceVersion)
	}

	clusterName, err := evt.Context.GetExtension("clustername")
	if err != nil {
		return nil, fmt.Errorf("failed to get clustername extension: %v", err)
	}

	clusterNameStr, ok := clusterName.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert clustername - %v to string", clusterName)
	}

	data := evt.Data()
	resourceStatusPayload := &workpayload.ManifestStatus{}
	if err := json.Unmarshal(data, resourceStatusPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data as resource status: %v", err)
	}

	resource := &db.Resource{
		Id:                   resourceIDStr,
		ResourceGenerationID: resourceVersionInt,
		ConsumerId:           clusterNameStr,
		Status: db.StatusMessage{
			MessageMeta: db.MessageMeta{
				ResourceGenerationID: resourceVersionInt,
			},
			ReconcileStatus: db.ReconcileStatus{
				Conditions: resourceStatusPayload.Conditions,
			},
		},
	}

	if resourceStatusPayload.Status != nil {
		unsObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resourceStatusPayload.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to convert manifest to unstructured object: %v", err)
		}
		resource.Status.ContentStatus = unsObj
	}

	return resource, nil
}
