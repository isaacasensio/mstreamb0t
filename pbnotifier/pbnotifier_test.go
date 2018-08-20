package pbnotifier

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xconstruct/go-pushbullet"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func successfulPushStub(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp string
		switch r.RequestURI {
		case "/devices":
			resp = `{ "devices": [ {"Iden": "iOS"}, {"Iden": "PC"}, {"Iden": "browser"}] }`
		case "/pushes":
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			bodyString := string(body)
			assert.JSONEq(t, `{
				"device_iden": "iOS",
				"type": "note",
				"title": "title",
				"body": "body"
			}`, bodyString)
			resp = `{}`
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
		w.WriteHeader(200)
	}))
}

func getDeviceErrorStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error": 
		{
			"type": "type", 
			"message": "get devices failed!", 
			"cat": "cat"
		}
	}`, http.StatusInternalServerError)
	}))
}

func pushErrorStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp string
		switch r.RequestURI {
		case "/devices":
			resp = `{ "devices": [ {"Iden": "iOS"}, {"Iden": "PC"}, {"Iden": "browser"}] }`
		case "/pushes":
			http.Error(w, `{"error": 
					{
						"type": "type", 
						"message": "push notification failed!", 
						"cat": "cat"
					}
				}`, http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
		w.WriteHeader(200)
	}))
}

func TestNotify_Successful(t *testing.T) {
	s := successfulPushStub(t)
	defer s.Close()

	c := NewClient("apiKey", s.URL)
	assert.NoError(t, c.Notify("title", "body"))
}

func TestNotify_ApiCallErrors(t *testing.T) {
	var flagtests = []struct {
		name           string
		server         *httptest.Server
		expectedErrMsg string
	}{
		{
			name:           "returns error when get devices fail",
			server:         getDeviceErrorStub(),
			expectedErrMsg: "get devices failed!",
		},
		{
			name:           "returns error when push fails",
			server:         pushErrorStub(),
			expectedErrMsg: "push notification failed!",
		},
	}

	for _, tt := range flagtests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.server.Close()
			c := NewClient("apiKey", tt.server.URL)

			err := c.Notify("title", "body")
			errorResp := err.(*pushbullet.ErrResponse)
			assert.Error(t, err)
			logrus.Info(errorResp)
			assert.Equal(t, &pushbullet.ErrResponse{
				Type:    "type",
				Message: tt.expectedErrMsg,
				Cat:     "cat",
			}, errorResp)
		})
	}
}
