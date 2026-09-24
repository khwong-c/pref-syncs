package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"uuid"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v3"
	"github.com/auth0/go-jwt-middleware/v3/jwks"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/samber/do/v2"
	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/features"
	"github.com/khwong-c/pref-syncs/models"
	"github.com/khwong-c/pref-syncs/tooling"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

// TODO: Implement This.
var (
	fixedAdminID = uuid.MustParse("01a08408-6237-7008-9508-263f5cf27759")
)

type userCtxKey struct{}

type UserInfoCtx struct {
	ID      uuid.UUID
	Issuer  string
	Subject string
	IsAdmin bool
}

type Authenticator struct {
	appLogic   *features.AppLogic
	providers  *jwks.MultiIssuerProvider
	validator  *validator.Validator
	middleware func(http.Handler) http.Handler
}

func NewAuthenticator(injector do.Injector, cfg *config.Config, localClient *http.Client) *Authenticator {
	validIssuers := []string{
		"https://FAKE.URL/STUB",
	}
	validAudiences := []string{
		"https://FAKE.URL/STUB",
	}
	providerOpts := make([]jwks.MultiIssuerProviderOption, 0, 1)
	localIssuer := fmt.Sprintf("http://127.0.0.1:%d%s", cfg.Port, cfg.IDP.Path)
	if cfg.IDP.Enable {
		validIssuers = append(validIssuers, localIssuer)
		validAudiences = append(validAudiences, localIssuer)
		if localClient != nil {
			providerOpts = append(
				providerOpts,
				jwks.WithMultiIssuerHTTPClient(localClient),
			)
		}
	}

	providers := tooling.Must(jwks.NewMultiIssuerProvider(providerOpts...))
	tokenValidator := tooling.Must(validator.New(
		validator.WithKeyFunc(providers.KeyFunc),
		validator.WithAlgorithms([]validator.SignatureAlgorithm{
			validator.RS256,
			validator.RS384,
			validator.RS512,
			validator.ES256,
			validator.ES384,
			validator.ES512,
			validator.ES256K,
			validator.PS256,
			validator.PS384,
			validator.PS512,
		}),
		validator.WithIssuers(validIssuers),
		validator.WithAudiences(validAudiences),
	))
	mwOptions := []jwtmiddleware.Option{
		jwtmiddleware.WithValidator(tokenValidator),
		jwtmiddleware.WithCredentialsOptional(true),
	}
	if cfg.IDP.Enable {
		mwOptions = append(mwOptions, jwtmiddleware.WithExclusionUrls([]string{
			fmt.Sprintf("%s/*", cfg.IDP.Path),
		}))
	}

	return &Authenticator{
		appLogic:   di.InvokeOrProvide(injector, features.NewAppLogic),
		providers:  providers,
		validator:  tokenValidator,
		middleware: tooling.Must(jwtmiddleware.New(mwOptions...)).CheckJWT,
	}
}

func (a *Authenticator) Middleware() func(http.Handler) http.Handler {
	return a.middleware
}

func (a *Authenticator) UserContext(cfg *config.Config) func(http.Handler) http.Handler {
	userCtxMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if !jwtmiddleware.HasClaims(ctx) {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := jwtmiddleware.GetClaims[*validator.ValidatedClaims](ctx)
			if err != nil {
				// Shouldn't occur.
				SimpleHTTPError(
					ctx, w, oops.FromContext(ctx).Wrap(err),
					"Internal Server Error", http.StatusInternalServerError,
				)
			}
			user, err := a.appLogic.GetUserByIssuer(
				ctx,
				claims.RegisteredClaims.Issuer,
				claims.RegisteredClaims.Subject,
			)
			if err != nil {
				SimpleHTTPError(
					ctx, w, oops.FromContext(ctx).Wrap(err),
					"Internal Server Error", http.StatusInternalServerError,
				)
			}
			ctx = context.WithValue(ctx, userCtxKey{}, &UserInfoCtx{
				Issuer:  user.AuthProvider,
				Subject: user.AuthUserID,
				ID:      user.ID,
				IsAdmin: user.IsAdmin,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	if !cfg.IDP.Enable {
		return userCtxMiddleware
	}

	// Promoting Specific Local User to Admin when Local IDP is enabled.
	localIssuer := fmt.Sprintf("http://127.0.0.1:%d%s", cfg.Port, cfg.IDP.Path)
	promotionMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if !jwtmiddleware.HasClaims(ctx) {
				next.ServeHTTP(w, r)
				return
			}
			u := a.GetUserInfo(ctx)
			if u != nil && !u.IsAdmin &&
				(u.Issuer == localIssuer) &&
				(u.Subject == models.StubClientService) {
				if _, err := a.appLogic.PromoteUserToAdmin(ctx, u.ID); err != nil {
					HandleHTTPError(ctx, w, err)
					return
				}
				u.IsAdmin = true
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	return func(next http.Handler) http.Handler {
		return userCtxMiddleware(promotionMiddleware(next))
	}
}

func (a *Authenticator) IsUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if a.GetUserInfo(ctx) == nil {
			err := oops.FromContext(ctx).
				New("Unauthorised")
			SimpleHTTPError(
				ctx, w, err,
				"Unauthorised", http.StatusUnauthorized,
			)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authenticator) IsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		info := a.GetUserInfo(ctx)
		if info == nil || !info.IsAdmin {
			err := oops.FromContext(ctx).
				New("Unauthorized")
			SimpleHTTPError(
				ctx, w, err,
				"Unauthorized", http.StatusUnauthorized,
			)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authenticator) GetUserInfo(ctx context.Context) *UserInfoCtx {
	userInfoCtx, ok := ctx.Value(userCtxKey{}).(*UserInfoCtx)
	if !ok {
		return nil
	}
	return userInfoCtx
}

func (a *Authenticator) GetUserID(ctx context.Context) uuid.UUID {
	userInfoCtx, ok := ctx.Value(userCtxKey{}).(*UserInfoCtx)
	if !ok {
		return uuid.Nil()
	}
	return userInfoCtx.ID
}
