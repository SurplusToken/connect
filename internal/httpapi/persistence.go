package httpapi

import (
	"errors"
	"net/http"
	"surplustoken-connect/internal/configuration"
	"surplustoken-connect/internal/gateway"
	"time"
)

type CredentialStore interface {
	Load() (string, error)
	Save(string) error
	Delete() error
}

func (a *API) remember(id string, local *localSession, response any) {
	if a.Store == nil {
		return
	}
	a.storeMu.Lock()
	a.savedID = id
	a.storeMu.Unlock()
	err := local.remote.Remember(func(token string) error {
		a.storeMu.Lock()
		defer a.storeMu.Unlock()
		if a.savedID != id {
			return nil
		} // Older sessions cannot overwrite a newer login.
		return a.Store.Save(token)
	})
	if err != nil {
		response.(map[string]any)["warning"] = err.Error() + "；本次登录仍可使用"
	}
}
func (a *API) forgetSaved(id string) error {
	a.storeMu.Lock()
	defer a.storeMu.Unlock()
	if a.Store == nil || a.savedID != id {
		return nil
	}
	if err := a.Store.Delete(); err != nil {
		return err
	}
	a.savedID = ""
	return nil
}
func (a *API) resume(w http.ResponseWriter, r *http.Request) {
	if a.Store == nil {
		respond(w, 200, map[string]bool{"saved": false})
		return
	}
	a.restoreMu.Lock()
	defer a.restoreMu.Unlock()
	a.storeMu.Lock()
	id := a.savedID
	a.storeMu.Unlock()
	if local := a.session(id); local != nil {
		v, err := a.finishLogin(r.Context(), id, local)
		if err != nil {
			fail(w, err)
			return
		}
		respond(w, 200, v)
		return
	}
	token, err := a.Store.Load()
	if err != nil {
		fail(w, err)
		return
	}
	if token == "" {
		respond(w, 200, map[string]bool{"saved": false})
		return
	}
	remote, err := a.Gateway.Resume(r.Context(), token)
	token = ""
	if err != nil {
		var failure *gateway.RemoteError
		if errors.As(err, &failure) && (failure.Status == 401 || failure.Status == 403) {
			if deleteErr := a.Store.Delete(); deleteErr != nil {
				fail(w, deleteErr)
				return
			}
			respond(w, 200, map[string]any{"saved": false, "warning": "登录已过期，请重新登录"})
			return
		}
		fail(w, errors.New("暂时无法恢复登录，请检查网络后重试"))
		return
	}
	id = configuration.ID()
	local := &localSession{remote: remote, created: time.Now(), ready: true}
	a.mu.Lock()
	a.sessions[id] = local
	a.mu.Unlock()
	// Save rotated token even if fetching the profile or keys subsequently fails.
	metadata := map[string]any{}
	a.remember(id, local, metadata)
	v, err := a.finishLogin(r.Context(), id, local)
	if err != nil {
		fail(w, err)
		return
	}
	if warning, ok := metadata["warning"]; ok {
		v.(map[string]any)["warning"] = warning
	}
	respond(w, 200, v)
}
