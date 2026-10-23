package rp

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// CardsService provides the /v1/cards role-card APIs.
type CardsService struct {
	client        *transport.Client
	controlClient *transport.Client
}

// ImportBatch imports 1..100 project cards using the Control API and the same
// project bearer token. body is {cards: [...]} (rows is a supported alias).
// Successful earlier cards are retained if a later card fails.
// Deprecated: retained for older deployments; current release onboarding uses Fork.
func (c *CardsService) ImportBatch(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	if c.controlClient == nil {
		return nil, &Error{Kind: ErrInvalid, Message: "Cards.ImportBatch requires ClientConfig.ControlAPIBaseURL (full /admin/v1 base)"}
	}
	var out JSONMap
	if err := requestJSON(ctx, c.controlClient, "POST", "/cards/import/batch", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Fork copies 1..100 catalog_ids into the project using the explicitly
// configured Control API. Inspect items and errors even when HTTP is 200.
func (c *CardsService) Fork(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	if c.controlClient == nil {
		return nil, &Error{Kind: ErrInvalid, Message: "Cards.Fork requires ClientConfig.ControlAPIBaseURL (full /admin/v1 base)"}
	}
	var out JSONMap
	if err := requestJSON(ctx, c.controlClient, "POST", "/projects/"+url.PathEscape(projectID)+"/cards/fork", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UntranslatedCardsQuery filters incomplete translations of latest cards.
type UntranslatedCardsQuery struct {
	Locale         string
	Limit          *uint64
	Offset         *uint64
	IncludePartial *bool
}

// Untranslated lists cards missing translations; summary describes this page.
func (c *CardsService) Untranslated(ctx context.Context, query UntranslatedCardsQuery, opts ...RequestOption) (JSONMap, error) {
	values := url.Values{"locale": {query.Locale}}
	if query.Limit != nil {
		values.Set("limit", uintString(*query.Limit))
	}
	if query.Offset != nil {
		values.Set("offset", uintString(*query.Offset))
	}
	if query.IncludePartial != nil {
		values.Set("include_partial", strconv.FormatBool(*query.IncludePartial))
	}
	var out JSONMap
	if err := requestJSON(ctx, c.client, "GET", "/cards/untranslated?"+values.Encode(), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ForkTranslations copies stored catalog translations using {items:[{card_id,
// card_version,locales}]}. Partial failures appear in the response errors array.
func (c *CardsService) ForkTranslations(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "POST", "/cards/translations/fork", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CardQuery contains optional localization and filter parameters.
type CardQuery struct {
	Lang    string
	IDs     []string
	CardIDs []string
	Limit   *uint64
	Offset  *uint64
}

// CardVersionQuery pages immutable versions, newest first.
type CardVersionQuery struct {
	Limit         *uint64
	BeforeVersion *uint64
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
// Deprecated: POST /v1/cards is scheduled to sunset on 2027-01-01.
// Use Import for delegated imports or Fork for catalog cards.
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

// Import passes through one delegated import payload. The current release
// handler accepts {card: {...}, overwrite?: false}; some docs instead show row.
// This engine endpoint requires a short-lived project delegation token.
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
	return c.ListVersionsWithQuery(ctx, id, CardVersionQuery{}, opts...)
}

// ListVersionsWithQuery pages versions using limit and before_version.
func (c *CardsService) ListVersionsWithQuery(ctx context.Context, id string, query CardVersionQuery, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	values := url.Values{}
	if query.Limit != nil {
		values.Set("limit", uintString(*query.Limit))
	}
	if query.BeforeVersion != nil {
		values.Set("before_version", uintString(*query.BeforeVersion))
	}
	suffix := ""
	if len(values) > 0 {
		suffix = "?" + values.Encode()
	}
	if err := requestJSON(ctx, c.client, "GET", "/cards/"+url.PathEscape(id)+"/versions"+suffix, nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetVersion returns one card version.
func (c *CardsService) GetVersion(ctx context.Context, id string, version uint64, opts ...RequestOption) (JSONMap, error) {
	return c.GetVersionWithQuery(ctx, id, version, CardQuery{}, opts...)
}

// GetVersionWithQuery returns a version localized using query.Lang.
func (c *CardsService) GetVersionWithQuery(ctx context.Context, id string, version uint64, query CardQuery, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "GET", "/cards/"+url.PathEscape(id)+"/versions/"+uintString(version)+cardQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
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

// RestoreVersion is retained for source compatibility.
// Deprecated: use RestoreVersionWithBody with expected_latest_version.
func (c *CardsService) RestoreVersion(ctx context.Context, id string, version uint64, opts ...RequestOption) (JSONMap, error) {
	return nil, &Error{Kind: ErrInvalid, Message: "restore requires expected_latest_version; use RestoreVersionWithBody"}
}

// RestoreVersionWithBody restores a version using an explicit latest-version CAS.
// body requires expected_latest_version and optionally change_note.
func (c *CardsService) RestoreVersionWithBody(ctx context.Context, id string, version uint64, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "POST", "/cards/"+url.PathEscape(id)+"/versions/"+uintString(version)+"/restore", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListTranslations returns stored translations grouped by language and field.
func (c *CardsService) ListTranslations(ctx context.Context, id string, version uint64, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "GET", "/cards/"+url.PathEscape(id)+"/versions/"+uintString(version)+"/translations", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveTranslations saves owner-provided translations. body requires user_id,
// expected_latest_version, lang and translations (field -> text); no model call.
func (c *CardsService) SaveTranslations(ctx context.Context, id string, version uint64, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "POST", "/cards/"+url.PathEscape(id)+"/versions/"+uintString(version)+"/translations", body, headersFromOptions(opts), &out); err != nil {
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
	if query.Limit != nil {
		values.Set("limit", uintString(*query.Limit))
	}
	if query.Offset != nil {
		values.Set("offset", uintString(*query.Offset))
	}
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
