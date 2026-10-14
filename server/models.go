package server

import (
	"context"
	"log/slog"
	"net/http"

	"afastchat.com/agents"
	"afastchat.com/db"
	"afastchat.com/views"
	"afastchat.com/views/ui"

	"github.com/starfederation/datastar-go/datastar"
)

func (x *Server) modelPicker(ctx context.Context, user db.User) (views.ModelPicker, error) {
	favorites, err := x.db.ListFavoriteModels(ctx, user.ID)
	if err != nil {
		return views.ModelPicker{}, err
	}

	providers := make([]views.ModelProvider, 0, len(agents.Providers))
	for _, provider := range agents.Providers {
		providers = append(providers, views.ModelProvider{ID: string(provider.ID), Label: provider.Label})
	}

	options := make([]views.ModelOption, 0, len(agents.Models))
	for _, model := range agents.Models {
		options = append(options, views.ModelOption{
			ID:          model.ID,
			Name:        model.Name,
			Provider:    string(model.Provider),
			Description: model.Description,
			CostTier:    model.CostTier,
		})
	}

	return views.ModelPicker{
		Providers: providers,
		Options:   options,
		Selected:  agents.ResolveModel(user.SelectedModel).ID,
		Favorites: favorites,
	}, nil
}

func (x *Server) handleSelectModel(w http.ResponseWriter, r *http.Request) {
	user := mustGetUser(r)
	var sig struct {
		Model string `json:"model"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		slog.Error("read selected model signals", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "The model could not be saved. Refresh the page, then try again."})
		return
	}

	model, ok := agents.LookupModel(sig.Model)
	if !ok {
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastWarning, Message: "That model is no longer available."})
		return
	}

	if err := x.db.UpdateUserSelectedModel(r.Context(), user.ID, model.ID); err != nil {
		slog.Error("save selected model", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "The model could not be saved. Try again."})
	}
}

func (x *Server) handleFavoriteModel(w http.ResponseWriter, r *http.Request) {
	user := mustGetUser(r)
	model, ok := agents.LookupModel(r.URL.Query().Get("model"))
	if !ok {
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastWarning, Message: "That model is no longer available."})
		return
	}

	favorite := r.URL.Query().Get("favorite") == "true"
	if err := x.db.SetFavoriteModel(r.Context(), user.ID, model.ID, favorite); err != nil {
		slog.Error("save favorite model", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Your favorites could not be saved. Try again."})
	}
}
