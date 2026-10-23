package rp

import (
	"context"
	"net/url"
	"strconv"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// GroupCardsService manages multi-character story cards and their rosters.
type GroupCardsService struct{ client *transport.Client }

// GroupCardQuery contains list filters and localization. Explicit false values
// for Listed and IncludeDeleted are preserved.
type GroupCardQuery struct {
	Q              string
	Lang           string
	Listed         *bool
	IncludeDeleted *bool
	Limit          *uint64
	Offset         *uint64
}

func groupCardPath(id string) string { return "/group-cards/" + url.PathEscape(id) }

func groupCardQueryString(query GroupCardQuery) string {
	values := url.Values{}
	if query.Q != "" {
		values.Set("q", query.Q)
	}
	if query.Lang != "" {
		values.Set("lang", query.Lang)
	}
	if query.Listed != nil {
		values.Set("listed", strconv.FormatBool(*query.Listed))
	}
	if query.IncludeDeleted != nil {
		values.Set("include_deleted", strconv.FormatBool(*query.IncludeDeleted))
	}
	if query.Limit != nil {
		values.Set("limit", uintString(*query.Limit))
	}
	if query.Offset != nil {
		values.Set("offset", uintString(*query.Offset))
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

// List lists story cards with optional filters.
func (g *GroupCardsService) List(ctx context.Context, query GroupCardQuery, opts ...RequestOption) ([]JSONMap, error) {
	var out []JSONMap
	if err := requestJSON(ctx, g.client, "GET", "/group-cards"+groupCardQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates a story card with referenced member cards.
func (g *GroupCardsService) Create(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "POST", "/group-cards", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get reads a story card, optionally localized.
func (g *GroupCardsService) Get(ctx context.Context, id string, query GroupCardQuery, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "GET", groupCardPath(id)+groupCardQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Update changes content and creates a new group version.
func (g *GroupCardsService) Update(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "PATCH", groupCardPath(id), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete soft-deletes a story card.
func (g *GroupCardsService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	return requestJSON(ctx, g.client, "DELETE", groupCardPath(id), nil, headersFromOptions(opts), nil)
}

// SetListing sets the story card listing state.
func (g *GroupCardsService) SetListing(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "PATCH", groupCardPath(id)+"/listing", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListVersions returns the server's version array.
func (g *GroupCardsService) ListVersions(ctx context.Context, id string, opts ...RequestOption) ([]JSONMap, error) {
	var out []JSONMap
	if err := requestJSON(ctx, g.client, "GET", groupCardPath(id)+"/versions", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetVersion reads one immutable group version.
func (g *GroupCardsService) GetVersion(ctx context.Context, id string, version uint64, query GroupCardQuery, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "GET", groupCardPath(id)+"/versions/"+uintString(version)+groupCardQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListMembers reads the roster.
func (g *GroupCardsService) ListMembers(ctx context.Context, id string, opts ...RequestOption) ([]JSONMap, error) {
	var out []JSONMap
	if err := requestJSON(ctx, g.client, "GET", groupCardPath(id)+"/members", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AddMember adds a referenced member.
func (g *GroupCardsService) AddMember(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "POST", groupCardPath(id)+"/members", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateMember updates a member with expected_latest_version.
func (g *GroupCardsService) UpdateMember(ctx context.Context, id, participantID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "PATCH", groupCardPath(id)+"/members/"+url.PathEscape(participantID), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RemoveMember soft-deletes a member; body includes user_id and expected_latest_version.
func (g *GroupCardsService) RemoveMember(ctx context.Context, id, participantID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "DELETE", groupCardPath(id)+"/members/"+url.PathEscape(participantID), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ExtractCharacters infers candidates without changing the roster; requires prompt access.
func (g *GroupCardsService) ExtractCharacters(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "POST", groupCardPath(id)+"/characters", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Preview renders a prompt without calling a model.
func (g *GroupCardsService) Preview(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "POST", groupCardPath(id)+"/preview", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Import imports a source card; dry_run returns the proposed roster only.
func (g *GroupCardsService) Import(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "POST", "/group-cards/import", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListTranslations reads owner translations for one version.
func (g *GroupCardsService) ListTranslations(ctx context.Context, id string, version uint64, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "GET", groupCardPath(id)+"/versions/"+uintString(version)+"/translations", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveTranslations saves owner-provided group translations.
func (g *GroupCardsService) SaveTranslations(ctx context.Context, id string, version uint64, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "POST", groupCardPath(id)+"/versions/"+uintString(version)+"/translations", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}
