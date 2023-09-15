package cloudevents

import (
	"encoding/json"
	"fmt"
	"strconv"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudeventstypes "github.com/cloudevents/sdk-go/v2/types"
	"github.com/kube-orchestra/maestro/internal/db"
	cegeneric "open-cluster-management.io/api/cloudevents/generic"
	cetypes "open-cluster-management.io/api/cloudevents/generic/types"
	workpayload "open-cluster-management.io/api/cloudevents/work/payload"
	workv1 "open-cluster-management.io/api/work/v1"
)

type Codec struct{}

var _ cegeneric.Codec[*db.Resource] = &Codec{}

func (codec *Codec) EventDataType() cetypes.CloudEventsDataType {
	return workpayload.ManifestEventDataType
}

func (codec *Codec) Encode(source string, eventType cetypes.CloudEventsType, obj *db.Resource) (*cloudevents.Event, error) {
	evtBuilder := cetypes.NewEventBuilder(source, eventType).
		WithResourceID(obj.Id).
		WithResourceVersion(obj.ResourceGenerationID).
		WithClusterName(obj.ConsumerId)

	if !obj.Object.GetDeletionTimestamp().IsZero() {
		evtBuilder.WithDeletionTimestamp(obj.Object.GetDeletionTimestamp().Time)
	}

	evt := evtBuilder.NewEvent()

	resourcePayload := &workpayload.Manifest{
		Manifest: obj.Object,
		DeleteOption: &workv1.DeleteOption{
			PropagationPolicy: workv1.DeletePropagationPolicyTypeForeground,
		},
		ConfigOption: &workpayload.ManifestConfigOption{
			FeedbackRules: []workv1.FeedbackRule{
				{
					Type: workv1.JSONPathsType,
					JsonPaths: []workv1.JsonPath{
						{
							Name: "status",
							Path: ".status",
						},
					},
				},
			},
			UpdateStrategy: &workv1.UpdateStrategy{
				Type: workv1.UpdateStrategyTypeUpdate,
			},
		},
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
	eventType, err := cetypes.ParseCloudEventsType(evt.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to parse cloud event type %s, %v", evt.Type(), err)
	}

	if eventType.CloudEventsDataType != workpayload.ManifestEventDataType {
		return nil, fmt.Errorf("unsupported cloudevents data type %s", eventType.CloudEventsDataType)
	}

	evtExtensions := evt.Context.GetExtensions()

	resourceID, err := cloudeventstypes.ToString(evtExtensions[cetypes.ExtensionResourceID])
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceid extension: %v", err)
	}

	resourceVersion, err := cloudeventstypes.ToString(evtExtensions[cetypes.ExtensionResourceVersion])
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceversion extension: %v", err)
	}

	resourceVersionInt, err := strconv.ParseInt(resourceVersion, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resourceversion - %v to int64", resourceVersion)
	}

	clusterName, err := cloudeventstypes.ToString(evtExtensions[cetypes.ExtensionClusterName])
	if err != nil {
		return nil, fmt.Errorf("failed to get clustername extension: %v", err)
	}

	data := evt.Data()
	resourceStatusPayload := &workpayload.ManifestStatus{}
	if err := json.Unmarshal(data, resourceStatusPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data as resource status: %v", err)
	}

	resource := &db.Resource{
		Id:                   resourceID,
		ResourceGenerationID: resourceVersionInt,
		ConsumerId:           clusterName,
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
		for _, value := range resourceStatusPayload.Status.StatusFeedbacks.Values {
			if value.Name == "status" {
				contentStatus := make(map[string]interface{})
				if err := json.Unmarshal([]byte(*value.Value.JsonRaw), &contentStatus); err != nil {
					return nil, fmt.Errorf("failed to convert status feedback value to content status: %v", err)
				}
				resource.Status.ContentStatus = contentStatus
			}
		}
	}

	return resource, nil
}
