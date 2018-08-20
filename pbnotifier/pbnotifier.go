package pbnotifier

import (
	"net/http"

	pushbullet "github.com/xconstruct/go-pushbullet"
)

// Client struct
type Client struct {
	client pushbullet.Client
}

// NewClient creates an instance of a PushBulletClient
func NewClient(apikey, endpoint string) Client {
	endpointPB := pushbullet.Endpoint{URL: endpoint}
	return Client{
		client: pushbullet.Client{
			Key:      apikey,
			Client:   http.DefaultClient,
			Endpoint: endpointPB},
	}
}

// Notify sends a notification to a pushbullet device
func (pb Client) Notify(title, body string) error {
	devs, err := pb.client.Devices()
	if err != nil {
		return err
	}
	return pb.client.PushNote(devs[0].Iden, title, body)
}
