package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/libraz/go-oidc-provider/op"
	"github.com/libraz/go-oidc-provider/op/feature"
	"github.com/libraz/go-oidc-provider/op/grant"
	"github.com/libraz/go-oidc-provider/op/store"
	"github.com/libraz/go-oidc-provider/op/storeadapter/inmem"
	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/tooling"
)

func createLocalIDP(cfg *config.Config, path string) (http.Handler, error) {
	// Generate ephemeral keys for signing.
	const CookieKeyLength = 32
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	cookieKey := make([]byte, CookieKeyLength) // 32 bytes for AES-256-GCM
	tooling.Must(rand.Read(cookieKey))

	// Create storage and seed the storage with a default user.
	st := inmem.New()
	st.PutUserWithPassword(context.Background(), &store.User{
		Subject: "root-user",
		Claims:  map[string]any{"name": "Root User"},
	},
		cfg.IDP.User,
		tooling.Must(op.HashPassword(cfg.IDP.Password)),
	)

	handler, err := op.New(
		op.WithIssuer(fmt.Sprintf("http://localhost:%d%s", cfg.Port, path)),
		op.WithStore(st),
		op.WithKeyset(op.Keyset{{KeyID: "k1", Signer: priv}}),
		op.WithCookieKeys(cookieKey),
		op.WithGrants(
			grant.AuthorizationCode,
			grant.RefreshToken,
			grant.ClientCredentials,
		),
		op.WithLoginFlow(op.LoginFlow{
			Primary: op.PrimaryPassword{Store: st.UserPasswords()},
		}),

		op.WithFeature(feature.Introspect),
		op.WithAllowLocalhostLoopback(),
		op.WithOpenIDScopeOptional(),
	)
	if err != nil {
		return nil, oops.Wrap(err)
	}
	return handler, nil
}
