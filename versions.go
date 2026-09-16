package rp

import (
	"context"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// VersionsService provides the version preview endpoint.
type VersionsService struct {
	client *transport.Client
}

// Preview renders a released version's assembled messages without creating a
// session.
func (v *VersionsService) Preview(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, v.client, "POST", "/versions/"+id+"/preview", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}
