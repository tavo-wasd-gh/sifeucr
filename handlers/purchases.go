package handlers

import (
	"net/http"

	"sifeucr/config"
	"sifeucr/internal/db"

	"git.tavo.one/tavo/axiom/forms"
)

func (h *Handler) AddPurchase() {
	// TODO: AddPurchase
}

func (h *Handler) PurchaseMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID := getUserIDFromContext(ctx)
			accountID := getAccountIDFromContext(ctx)

			queries := db.New(h.DB())
			perm, err := queries.PermissionByUserIDAndAccountID(ctx, db.PermissionByUserIDAndAccountIDParams{
				UserID:    userID,
				AccountID: accountID,
			})
			if err != nil {
				h.Log().Error("failed to authenticate user: %v", err)
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			type purchaseClaims struct {
				Dist int64 `form:"purchase_dist"`
			}

			form, err := forms.FormToStruct[purchaseClaims](r)
			if err != nil {
				h.Log().Error("error casting form to struct: %v", err)
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			if form.Dist != 0 && !config.HasPermission(
				perm.PermissionInteger,
				config.WriteOther,
			) {
				h.Log().Error("cannot add request, insufficient permissions: %v", err)
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
