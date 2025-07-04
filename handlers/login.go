package handlers

import (
	"context"
	"net/http"
	"strings"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/mail"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
	"sifeucr/internal/db"
)

func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	type loginForm struct {
		Email    string `form:"email" fmt:"trim,lower" validate:"email" req:"1"`
		Password string `form:"password" req:"1"`
		Account  int    `form:"account"`
	}

	login, err := forms.FormToStruct[loginForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if strings.Contains(login.Email, "@") {
		if !strings.Contains(strings.ToLower(login.Email), "@ucr.ac.cr") {
			// Is an external provider
			views.RenderHTML(w, r, "login", map[string]any{"ExternalEmail": true})
			return
		}
	} else {
		login.Email += "@ucr.ac.cr"
	}

	dbuser := login.Email
	pos := strings.IndexRune(dbuser, '@')
	if pos != -1 {
		dbuser = dbuser[:pos]
	}

	queries := db.New(h.DB())

	userID, err := queries.UserIDByUserEmail(ctx, dbuser)
	if err != nil {
		h.Log().Error("error querying user_id by user_email: %v", err)
		views.RenderHTML(w, r, "login", map[string]any{"Error": true})
		return
	}

	allowedAccounts, err := queries.AllowedAccountsByUserID(ctx, userID)
	if err != nil {
		h.Log().Error("error querying allowed_accounts by user_id: %v", err)
		views.RenderHTML(w, r, "login", map[string]any{"Error": true})
		return
	}

	var chosenAccountID int64 = 0

	switch len(allowedAccounts) {
	case 0:
		// No allowed accounts, render error and return
		h.Log().Error("no allowed_accounts for user_id")
		views.RenderHTML(w, r, "login", map[string]any{"Error": true})
		return

	case 1:
		// One allowed account, set and continue
		chosenAccountID = allowedAccounts[0].AccountID

	default:
		type multiple struct {
			MultipleAccounts bool
			AllowedAccounts  []db.AllowedAccountsByUserIDRow
		}

		m := multiple{
			MultipleAccounts: true,
			AllowedAccounts:  allowedAccounts,
		}

		views.RenderHTML(w, r, "login", m)
		return
	}

	if h.Production() {
		s := smtp.Client("smtp.ucr.ac.cr", "587", login.Password)

		if err := s.Validate(login.Email); err != nil {
			h.Log().Error("error validating user %s: %v", login.Email, err)
			views.RenderHTML(w, r, "login", map[string]any{"Error": true})
			return
		}
	}

	perm, err := queries.GetPermission(ctx, db.GetPermissionParams{
		PermissionUser:    userID,
		PermissionAccount: chosenAccountID,
	})

	if err != nil {
		h.Log().Error("error querying permissions: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if !config.HasPermission(perm.PermissionInteger, config.Read) {
		h.Log().Error("incorrect permissions, got:%d want:%d", perm.PermissionInteger, config.Read)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	st, ct, err := h.Sessions().New(
		config.SessionMaxAge,
		config.Session{
			UserID:    userID,
			AccountID: chosenAccountID,
		},
	)
	if err != nil {
		h.Log().Error("failed to create session: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	ctx = context.WithValue(ctx, config.UserIDKey, userID)
	ctx = context.WithValue(ctx, config.AccountIDKey, chosenAccountID)
	ctx = context.WithValue(ctx, config.CSRFTokenKey, ct)

	dashboard, err := h.loadDashboard(ctx)
	if err != nil {
		h.Log().Error("failed to load dashboard: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.SessionTokenKey,
		Value:    st,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.Production(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   config.CookieMaxAge,
	})

	if err = views.RenderHTML(w, r, "dashboard", dashboard); err != nil {
		h.Log().Error("failed to render dashboard: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
