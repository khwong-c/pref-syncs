//go:generate go tool openapi-client-generator generate --spec ../openapi.yaml --out ./client
package tests

import (
	"net/http"

	"github.com/khwong-c/httptestclient"

	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tests/client"
)

func createClients(svr *server.Server) (*client.Client, *http.Client) {
	httpClient := httptestclient.New(svr.Handler)
	c := client.NewClient(
		"http://localhost:7086",
		client.WithHTTPClient(httpClient),
	)
	return c, httpClient
}
