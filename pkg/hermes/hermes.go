package hermes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	log "github.com/sirupsen/logrus"
)

type (
	// Service Codefresh Service
	Service interface {
		TriggerEvent(eventURI string, event *NormalizedEvent) error
	}

	// APIEndpoint Hermes API endpoint
	APIEndpoint struct {
		endpoint *sling.Sling
	}

	// NormalizedEvent normalized event: {event-uri, original-payload, secret, variables-map}
	NormalizedEvent struct {
		Original  string            `json:"original,omitempty"`
		Secret    string            `json:"secret,omitempty"`
		Variables map[string]string `json:"variables,omitempty"`
	}

	// ServiceError hermes error
	ServiceError struct {
		Status int    `json:"status,omitempty"`
		Title  string `json:"message,omitempty"`
		Body   string `json:"error,omitempty"`
	}
)

// NewNormalizedEvent init NormalizedEvent struct
func NewNormalizedEvent() *NormalizedEvent {
	var event NormalizedEvent
	event.Variables = make(map[string]string)
	return &event
}

// NewHermesEndpoint create new Hermes API endpoint from url and API token
func NewHermesEndpoint(url, token string) Service {
	log.WithField("hermes url", url).Debug("Binding to Hermes service")
	endpoint := sling.New().Base(url).Set("Authorization", token)
	return &APIEndpoint{endpoint}
}

// TriggerEvent send normalized event to Hermes trigger-manager server
func (api *APIEndpoint) TriggerEvent(eventURI string, event *NormalizedEvent) error {
	log.WithField("event-uri", eventURI).Debug("Triggering event")
	// runs response
	type PipelineRun struct {
		ID    string `json:"id"`
		Error error  `json:"error, omitempty"`
	}
	runs := new([]PipelineRun)

	var failure ServiceError
	log.WithField("event", *event).Debug("Sending normalized event payload")
	resp, err := api.endpoint.New().Post(fmt.Sprint("trigger/", eventURI)).BodyJSON(event).Receive(&runs, &failure)
	if err != nil {
		log.WithError(err).Error("Failed to invoke Hermes POST /trigger/ API")
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.WithField("http status", resp.Status).Error("Herems POST /trigger/ API failed")
		return fmt.Errorf("%s: error triggering event '%s'", resp.Status, eventURI)
	}
	log.WithField("event-uri", eventURI).Debug("Event triggered")
	log.WithField("runs", runs).Debug("Running pipelines")
	return nil
}

func (e NormalizedEvent) String() string {
	d, err := json.Marshal(&e)
	if err != nil {
		log.WithError(err).Error("Failed to convert NormalizedEvent to JSON")
	}
	return string(d)
}
