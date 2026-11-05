package rp

import (
	"context"
	"net/url"
	"strconv"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// WorldBookService provides project and catalog world-book APIs.
type WorldBookService struct{ client *transport.Client }

// WorldBooksService is a pluralized compatibility alias for WorldBookService.
type WorldBooksService = WorldBookService

type WorldBookQuery struct {
	Limit  *uint64
	Cursor string
}

func (w *WorldBookService) List(ctx context.Context, query WorldBookQuery, opts ...RequestOption) (JSONMap, error) {
	return w.list(ctx, "/world-books", query, opts...)
}

func (w *WorldBookService) Get(ctx context.Context, id string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, w.client, "GET", "/world-books/"+url.PathEscape(id), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *WorldBookService) Create(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, w.client, "POST", "/world-books", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *WorldBookService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	return requestJSON(ctx, w.client, "DELETE", "/world-books/"+url.PathEscape(id), nil, headersFromOptions(opts), nil)
}

func (w *WorldBookService) CatalogList(ctx context.Context, query WorldBookQuery, opts ...RequestOption) (JSONMap, error) {
	return w.list(ctx, "/catalog/world-books", query, opts...)
}

func (w *WorldBookService) CatalogGet(ctx context.Context, id string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, w.client, "GET", "/catalog/world-books/"+url.PathEscape(id), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *WorldBookService) list(ctx context.Context, route string, query WorldBookQuery, opts ...RequestOption) (JSONMap, error) {
	values := url.Values{}
	if query.Limit != nil {
		values.Set("limit", strconv.FormatUint(*query.Limit, 10))
	}
	if query.Cursor != "" {
		values.Set("cursor", query.Cursor)
	}
	if encoded := values.Encode(); encoded != "" {
		route += "?" + encoded
	}
	var out JSONMap
	if err := requestJSON(ctx, w.client, "GET", route, nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}
