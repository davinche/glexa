package glexa

import (
	"encoding/json"
	"io"
)

// Body is the post body from an AWS Alexa request
type Body struct {
	Version string `json:"version"`
	Session struct {
		New         bool   `json:"new"`
		SessionID   string `json:"sessionId"`
		Application struct {
			ApplicationID string `json:"applicationId"`
		} `json:"application"`
		Attributes map[string]struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json: "attributes"`
		User struct {
			UserID      string `json:"userId"`
			AccessToken string `json:"accessToken"`
		} `json:"user,omitempty"`
	} `json:"session"`
	Request alexaRequest `json:"request"`
}

type alexaRequest struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
	Reason    string `json:"reason,omitempty"`
	Intent    struct {
		Name  string `json:"name"`
		Slots map[string]struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"slots"`
	} `json:"intent,omitempty"`
}

type alexaSpeech struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	SSML string `json:"ssml,omitempty"`
}

type alexaCard struct {
	Type    string          `json:"type,omitempty"`
	Title   string          `json:"title,omitempty"`
	Content string          `json:"content,omitempty"`
	Text    string          `json:"text,omitempty"`
	Image   *alexaCardImage `json:"image,omitempty"`
}

type alexaCardImage struct {
	SmallImageURL string `json:"smallImageUrl,omitempty"`
	LargeImageURL string `json:"largeImageUrl,omitempty"`
}

type alexaReprompt struct {
	OutputSpeech *alexaSpeech `json:"outputSpeech,omitempty"`
}

type alexaResponse struct {
	OutputSpeech     *alexaSpeech   `json:"outputSpeech,omitempty"`
	Card             *alexaCard     `json:"card,omitempty"`
	Reprompt         *alexaReprompt `json:"reprompt,omitempty"`
	ShouldEndSession bool           `json:"shouldEndSession"`
}

// Response is the response object for an Alexa Request
type Response struct {
	Version           string `json:"version"`
	SessionAttributes map[string]struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"sessionAttributes,omitempty"`
	Response alexaResponse `json:"response,omitempty"`
}

// ParseBody returns a new Body struct
func ParseBody(r io.Reader) (*Body, error) {
	reqBody := Body{}
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&reqBody)
	if err != nil {
		return nil, err
	}
	return &reqBody, nil
}

func (r *alexaRequest) IsLaunch() bool {
	return r.Type == "LaunchRequest"
}

func (r *alexaRequest) IsIntent() bool {
	return r.Type == "IntentRequest"
}

func (r *alexaRequest) IsSessionEnded() bool {
	return r.Type == "SessionEndedRequest"
}

// NewResponse creates a new response for an Alexa Request
func NewResponse() *Response {
	return &Response{
		Version: "1.0",
	}
}

// Tell responds with a given speech text
func (r *Response) Tell(text string) {
	r.Response.OutputSpeech = &alexaSpeech{
		Type: "PlainText",
		Text: text,
	}
}

// Ask responds with a given speech text to prompt retry
func (r *Response) Ask(text string) {
	r.Response.Reprompt = &alexaReprompt{
		OutputSpeech: &alexaSpeech{
			Type: "PlainText",
			Text: text,
		},
	}
	r.Response.ShouldEndSession = false
}
