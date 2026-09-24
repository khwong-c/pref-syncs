//go:generate go tool openapi-client-generator generate --spec ../openapi.yaml --out ./client
package tests

import (
	"github.com/khwong-c/httptestclient"

	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tests/client"
)

func createAPIClient(svr *server.Server, opts ...client.ClientOption) *client.Client {
	opts = append(
		opts,
		client.WithHTTPClient(httptestclient.New(svr.Handler)),
	)
	c := client.NewClient(
		"http://localhost:7086",
		opts...,
	)
	return c
}
