package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/libraz/go-oidc-provider/op"
	"github.com/libraz/go-oidc-provider/op/feature"
	"github.com/libraz/go-oidc-provider/op/grant"
	"github.com/libraz/go-oidc-provider/op/store"
	"github.com/libraz/go-oidc-provider/op/storeadapter/inmem"
	"github.com/samber/oops"
	"golang.org/x/oauth2"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/server/middlewares"
	"github.com/khwong-c/pref-syncs/tooling"
)

const (
	StubScope         = "api:stub"
	StubClientPublic  = "stub-client-rp"
	StubClientService = "stub-client-service"
	StubClientSecret  = "stub-client-secret" //nolint:gosec
	StubRootSubject   = "root-user"
)

func createLocalIDP(cfg *config.Config, idpPath string, cbPath string) (http.Handler, error) {
	// Generate ephemeral keys for signing.
	const CookieKeyLength = 32
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	cookieKey := make([]byte, CookieKeyLength) // 32 bytes for AES-256-GCM
	tooling.Must(rand.Read(cookieKey))

	// Create storage and seed the storage with a default user.
	st := inmem.New()
	st.PutUserWithPassword(context.Background(), &store.User{
		Subject: StubRootSubject,
		Claims:  map[string]any{"name": "Root User"},
	},
		cfg.IDP.User,
		tooling.Must(op.HashPassword(cfg.IDP.Password)),
	)

	handler, err := op.New(
		op.WithIssuer(fmt.Sprintf("http://127.0.0.1:%d%s", cfg.Port, idpPath)),
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
		op.WithScope(
			op.PublicScope(StubScope, "Stub Scope for testing"),
		),
		op.WithFirstPartyClients(
			StubClientPublic,
			StubClientService,
		),
		op.WithStaticClients(
			op.ConfidentialClient{
				ID:         StubClientPublic,
				Secret:     StubClientSecret,
				AuthMethod: op.AuthClientSecretBasic,
				RedirectURIs: []string{
					fmt.Sprintf("http://127.0.0.1:%d%s", cfg.Port, cbPath),
					fmt.Sprintf("http://127.0.0.1:%d", cfg.Port),
				},
				GrantTypes: []string{
					"authorization_code",
					"refresh_token",
					"client_credentials",
				},
				Scopes: []string{"openid", "profile", "email"},
			},
			op.ConfidentialClient{
				ID:         StubClientService,
				Secret:     StubClientSecret,
				AuthMethod: op.AuthClientSecretBasic,
				RedirectURIs: []string{
					fmt.Sprintf("http://127.0.0.1:%d%s", cfg.Port, cbPath),
					fmt.Sprintf("http://127.0.0.1:%d", cfg.Port),
				},
				GrantTypes: []string{"client_credentials"},
				Scopes:     []string{StubScope},
			},
		),

		op.WithFeature(feature.Introspect),
		op.WithAllowLocalhostLoopback(),
		op.WithOpenIDScopeOptional(),
		op.WithCORSOrigins(
			fmt.Sprintf("http://127.0.0.1:%d", cfg.Port),
			fmt.Sprintf("http://localhost:%d", cfg.Port),
		),
	)
	if err != nil {
		return nil, oops.Wrap(err)
	}
	return handler, nil
}

func createLocalIDPCallbackHandler(cfg *config.Config, idpPath string, cbPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		ctx := r.Context()
		if code == "" {
			middlewares.SimpleHTTPError(
				ctx, w,
				oops.Errorf("Missing code parameter"),
				"Missing Code Parameter",
				http.StatusBadRequest,
			)
			return
		}

		oauthConfig := &oauth2.Config{
			ClientID:     StubClientPublic,
			ClientSecret: StubClientSecret,
			Scopes:       []string{"openid"},
			Endpoint: oauth2.Endpoint{
				AuthURL:   fmt.Sprintf("http://127.0.0.1:%d%s/oidc/auth", cfg.Port, idpPath),
				TokenURL:  fmt.Sprintf("http://127.0.0.1:%d%s/oidc/token", cfg.Port, idpPath),
				AuthStyle: oauth2.AuthStyleAutoDetect,
			},
			RedirectURL: fmt.Sprintf("http://127.0.0.1:%d%s", cfg.Port, cbPath),
		}
		token, err := oauthConfig.Exchange(context.Background(), code)
		if err != nil {
			middlewares.SimpleHTTPError(
				ctx, w,
				oops.Wrap(err),
				"Failed to exchange token",
				http.StatusInternalServerError,
			)
			return
		}
		render.JSON(w, r, token)
	}
}
