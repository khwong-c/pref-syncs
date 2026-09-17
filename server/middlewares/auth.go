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
	"github.com/khwong-c/pref-syncs/tooling"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

// TODO: Implement This.
var (
	fixedUserID  = uuid.MustParse("01a0734e-795a-72bc-874b-9b90c41ddc8b")
	fixedAdminID = uuid.MustParse("01a08408-6237-7008-9508-263f5cf27759")
)

type userCtxKey struct{}

type UserInfoCtx struct {
	Issuer           string
	UserIDFromIssuer string
	UserID           uuid.UUID
}

type Authenticator struct {
	appLogic   *features.AppLogic
	providers  *jwks.MultiIssuerProvider
	validator  *validator.Validator
	middleware func(http.Handler) http.Handler
}

func NewAuthenticator(injector do.Injector, cfg *config.Config) *Authenticator {
	validIssuers := []string{}
	validAudiences := []string{}
	if cfg.IDP.Enable {
		validIssuers = append(validIssuers, fmt.Sprintf("http://127.0.0.1:%d%s", cfg.Port, cfg.IDP.Path))
		validAudiences = append(validAudiences, fmt.Sprintf("http://127.0.0.1:%d%s", cfg.Port, cfg.IDP.Path))
	}

	providers := tooling.Must(jwks.NewMultiIssuerProvider())
	tokenValidator := tooling.Must(validator.New(
		validator.WithKeyFunc(providers.KeyFunc),
		validator.WithAlgorithm(validator.RS256),
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

func (a *Authenticator) UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !jwtmiddleware.HasClaims(ctx) {
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authenticator) IsUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if GetUserInfo(ctx) == nil {
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
		info := GetUserInfo(ctx)
		isAdmin := info.UserID == fixedAdminID // TODO: Implement me
		if !isAdmin {
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

func GetUserInfo(ctx context.Context) *UserInfoCtx {
	userInfoCtx, ok := ctx.Value(userCtxKey{}).(*UserInfoCtx)
	if !ok {
		return nil
	}
	return userInfoCtx
}
