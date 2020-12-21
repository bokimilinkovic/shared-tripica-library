package tripica

import (
	"fmt"
	gohttp "net/http"
	"tripica-client/http"
	"tripica-client/http/errors"
)

const (
	notifyPathSendNotification = "/notif"
	notifyPathSendTermination  = "/terminate"
	// Listing of possible event names.
	NotifSentEventName         = "EXTERNAL_NOTIF_SENT"
	TerminateContractEventName = "EXTERNAL_TERMINATE_CONTRACT"
)

// notifyAPI manages endpoints for notifying triPica.
type notifyAPI struct {
	httpClient *http.Client
	address    string
}

// Notify notifies triPica about certain event.
func (n *notifyAPI) Notify(req *NotifyRequest) error {
	var url string

	switch req.EventName {
	case NotifSentEventName:
		url = fmt.Sprintf(n.address + notifyPathSendNotification)
	case TerminateContractEventName:
		url = fmt.Sprintf(n.address + notifyPathSendTermination)
	default:
		return NewTriPicaError(fmt.Errorf("unknown event name in NotifyRequest: %s", req.EventName))
	}

	resp, err := n.httpClient.Post(url, req, http.JSONContent())
	if err != nil {
		return NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() != gohttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}

		return NewTriPicaError(fmt.Errorf("notify request failed with %w", err))
	}

	return nil
}

// EventStatus represents one of the possible statuses an event can have.
type EventStatus string

// NotifyRequest represents a notify request towards triPica.
type NotifyRequest struct {
	EventName       string            `json:"eventName"`
	CaseExternalID  string            `json:"caseExternalId"`
	EventExternalID string            `json:"eventExternalId"`
	EventDate       int               `json:"eventDate"`
	Status          string            `json:"status"`
	Comment         string            `json:"comment"`
	Attachments     []Attachment      `json:"attachments,omitempty"`
	Characteristics map[string]string `json:"characteristics,omitempty"`
}

// Attachment represents attachment in NotifyRequest.
type Attachment struct {
	FileName        string                    `json:"fileName"`
	MediaType       string                    `json:"mediaType"`
	Characteristics AttachmentCharacteristics `json:"characteristics"`
	Content         AttachmentContent         `json:"content"`
}

// AttachmentCharacteristics represents characteristics in Attachment.
type AttachmentCharacteristics struct {
	PDFTemplate string `json:"pdfTemplate"`
}

// AttachmentContent represents content in Attachment.
type AttachmentContent struct {
	Body string `json:"body"`
}
