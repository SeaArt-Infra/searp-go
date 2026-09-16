package rp

import (
	"context"
	"net/url"
	"strings"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// CardsService provides the /v1/cards role-card APIs.
type CardsService struct {
	client *transport.Client
}

// CardQuery contains optional localization and filter parameters.
type CardQuery struct {
	Lang    string
	IDs     []string
	CardIDs []string
}

// List returns the project role cards.
func (c *CardsService) List(ctx context.Context, query CardQuery, opts ...RequestOption) ([]JSONMap, error) {
	var out []JSONMap
	if err := requestJSON(ctx, c.client, "GET", "/cards"+cardQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates or reuses a card.
func (c *CardsService) Create(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "POST", "/cards", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one card.
func (c *CardsService) Get(ctx context.Context, id string, query CardQuery, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "GET", "/cards/"+id+cardQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Update patches a card.
func (c *CardsService) Update(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "PATCH", "/cards/"+id, body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete deletes a card.
func (c *CardsService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	return requestJSON(ctx, c.client, "DELETE", "/cards/"+id, nil, headersFromOptions(opts), nil)
}

// Import imports a batch of cards.
func (c *CardsService) Import(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "POST", "/cards/import", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetListing updates a card listing state.
func (c *CardsService) SetListing(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "PATCH", "/cards/"+id+"/listing", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListVersions lists versions for a card.
func (c *CardsService) ListVersions(ctx context.Context, id string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "GET", "/cards/"+id+"/versions", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetVersion returns one card version.
func (c *CardsService) GetVersion(ctx context.Context, id string, version uint64, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "GET", "/cards/"+id+"/versions/"+uintString(version), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteVersion deletes one card version.
func (c *CardsService) DeleteVersion(ctx context.Context, id string, version uint64, opts ...RequestOption) error {
	return requestJSON(ctx, c.client, "DELETE", "/cards/"+id+"/versions/"+uintString(version), nil, headersFromOptions(opts), nil)
}

// UpdateTranslation updates a translated field for a card version.
func (c *CardsService) UpdateTranslation(ctx context.Context, id string, version uint64, lang, field string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	path := "/cards/" + id + "/versions/" + uintString(version) + "/translations/" + url.PathEscape(lang) + "/" + url.PathEscape(field)
	if err := requestJSON(ctx, c.client, "PATCH", path, body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RestoreVersion restores a card version.
func (c *CardsService) RestoreVersion(ctx context.Context, id string, version uint64, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "POST", "/cards/"+id+"/versions/"+uintString(version)+"/restore", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListByUser returns cards created by a user.
func (c *CardsService) ListByUser(ctx context.Context, userID string, query CardQuery, opts ...RequestOption) ([]JSONMap, error) {
	var out []JSONMap
	if err := requestJSON(ctx, c.client, "GET", "/users/"+userID+"/cards"+cardQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func cardQueryString(query CardQuery) string {
	values := url.Values{}
	if query.Lang != "" {
		values.Set("lang", query.Lang)
	}
	if len(query.IDs) > 0 {
		values.Set("ids", strings.Join(query.IDs, ","))
	}
	if len(query.CardIDs) > 0 {
		values.Set("card_ids", strings.Join(query.CardIDs, ","))
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
