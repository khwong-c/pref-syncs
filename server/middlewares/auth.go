package middlewares

import (
	"context"
	"net/http"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/samber/oops"
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

type Authenticator struct{}

func NewAuthenticator(injector do.Injector) *Authenticator {
	return &Authenticator{}
}

func (a *Authenticator) UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		info := &UserInfoCtx{
			Issuer:           "stub",
			UserIDFromIssuer: "stub-user",
			UserID:           fixedUserID,
		}
		if token == "fugu" {
			info.UserID = fixedAdminID
		}

		ctx := context.WithValue(r.Context(), userCtxKey{}, info)
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
