package rp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// adminService provides access to the SeaRP control-plane API served at
// /admin/v1 by the project gateway.
type adminService struct {
	client *transport.Client
}

// Request performs a JSON request against the /admin/v1 base and returns the
// decoded JSON object. It is useful for endpoints not yet exposed by a typed
// Admin method.
func (a *adminService) Request(ctx context.Context, method, path string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, method, path, body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Health returns the gateway health object from GET /admin/v1/health.
func (a *adminService) Health(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/health", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Whoami returns the current actor role and project ID, if any.
func (a *adminService) Whoami(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/whoami", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AgentContract returns the published client/agent contract.
func (a *adminService) AgentContract(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/agent/contract", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjects returns the projects visible to the current token.
func (a *adminService) ListProjects(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/projects", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProject creates a project with an explicit id and token.
func (a *adminService) CreateProject(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/projects", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProject returns one project by ID.
func (a *adminService) GetProject(ctx context.Context, id string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/projects/"+projectPathSegment(id), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteProject deletes one project by ID.
func (a *adminService) DeleteProject(ctx context.Context, id string, opts ...RequestOption) error {
	return adminRequestJSON(ctx, a.client, http.MethodDelete, "/projects/"+projectPathSegment(id), nil, headersFromOptions(opts), nil)
}

// RotateProjectToken rotates the project token and returns the updated project.
func (a *adminService) RotateProjectToken(ctx context.Context, id string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/projects/"+projectPathSegment(id)+"/token", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectLive returns the currently published live sample for a project.
func (a *adminService) GetProjectLive(ctx context.Context, id string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/projects/"+projectPathSegment(id)+"/live", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProjectLive updates the published live sample for a project.
func (a *adminService) UpdateProjectLive(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPut, "/projects/"+projectPathSegment(id)+"/live", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Raw performs a control-plane request and returns the status code and raw
// body. Use it for binary endpoints such as catalog card covers.
func (a *adminService) Raw(ctx context.Context, method, path string, body JSONMap, opts ...RequestOption) (int, []byte, error) {
	return a.client.Request(ctx, method, path, body, headersFromOptions(opts))
}

// ProjectRequest performs a JSON request against
// /admin/v1/projects/{projectID}/{subpath}. It is the generic escape hatch for
// project-scoped control-plane routes not covered by a typed method.
func (a *adminService) ProjectRequest(ctx context.Context, projectID, method, subpath string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	path := "/projects/" + projectPathSegment(projectID)
	if subpath != "" {
		path += "/" + subpath
	}
	if err := adminRequestJSON(ctx, a.client, method, path, body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGlobalPack returns the global prompt pack.
func (a *adminService) GetGlobalPack(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/pack", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateGlobalPack replaces the global prompt pack.
func (a *adminService) UpdateGlobalPack(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPut, "/pack", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectPack returns a project pack or its inherited global pack.
func (a *adminService) GetProjectPack(ctx context.Context, projectID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "pack"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PatchProjectPack updates a forked project pack.
func (a *adminService) PatchProjectPack(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPatch, projectSubpath(projectID, "pack"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteProjectPack removes a forked project pack and restores inheritance.
func (a *adminService) DeleteProjectPack(ctx context.Context, projectID string, opts ...RequestOption) error {
	return adminRequestJSON(ctx, a.client, http.MethodDelete, projectSubpath(projectID, "pack"), nil, headersFromOptions(opts), nil)
}

// ForkProjectPack forks the global pack into a project pack.
func (a *adminService) ForkProjectPack(ctx context.Context, projectID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "pack/fork"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListCatalog lists the shared role-card catalog.
func (a *adminService) ListCatalog(ctx context.Context, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/catalog"+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ImportCatalog imports one shared catalog card.
func (a *adminService) ImportCatalog(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/catalog/import", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCatalogCard returns one shared catalog card.
func (a *adminService) GetCatalogCard(ctx context.Context, cardID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/catalog/"+projectPathSegment(cardID), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateCatalogCard patches one shared catalog card.
func (a *adminService) UpdateCatalogCard(ctx context.Context, cardID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPatch, "/catalog/"+projectPathSegment(cardID), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteCatalogCard deletes one shared catalog card.
func (a *adminService) DeleteCatalogCard(ctx context.Context, cardID string, opts ...RequestOption) error {
	return adminRequestJSON(ctx, a.client, http.MethodDelete, "/catalog/"+projectPathSegment(cardID), nil, headersFromOptions(opts), nil)
}

// GetCatalogCardCover returns the binary cover image for a catalog card.
func (a *adminService) GetCatalogCardCover(ctx context.Context, cardID string, opts ...RequestOption) ([]byte, error) {
	return adminRawRequest(ctx, a.client, http.MethodGet, "/catalog/"+projectPathSegment(cardID)+"/cover", nil, headersFromOptions(opts))
}

// ListProjectCards lists project role-card metadata.
func (a *adminService) ListProjectCards(ctx context.Context, projectID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "cards")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectCard returns one project role card.
func (a *adminService) GetProjectCard(ctx context.Context, projectID, cardID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectCardPath(projectID, cardID), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProjectCard patches one project role card.
func (a *adminService) UpdateProjectCard(ctx context.Context, projectID, cardID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPatch, projectCardPath(projectID, cardID), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteProjectCard deletes one project role card.
func (a *adminService) DeleteProjectCard(ctx context.Context, projectID, cardID string, opts ...RequestOption) error {
	return adminRequestJSON(ctx, a.client, http.MethodDelete, projectCardPath(projectID, cardID), nil, headersFromOptions(opts), nil)
}

// SetProjectCardListing updates a project card listing flag.
func (a *adminService) SetProjectCardListing(ctx context.Context, projectID, cardID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPatch, projectCardPath(projectID, cardID)+"/listing", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ImportProjectCard imports one full card into a project.
func (a *adminService) ImportProjectCard(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "cards/import"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ImportProjectCardsBatch imports up to 100 full cards into a project.
func (a *adminService) ImportProjectCardsBatch(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "cards/import/batch"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ForkProjectCard forks a shared catalog card into a project.
func (a *adminService) ForkProjectCard(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "cards/fork"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectCardVersions lists versions for one project card.
func (a *adminService) ListProjectCardVersions(ctx context.Context, projectID, cardID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectCardPath(projectID, cardID)+"/versions", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectCardVersion returns one project card version.
func (a *adminService) GetProjectCardVersion(ctx context.Context, projectID, cardID string, version uint64, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectCardPath(projectID, cardID)+"/versions/"+uintString(version), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteProjectCardVersion deletes one project card version.
func (a *adminService) DeleteProjectCardVersion(ctx context.Context, projectID, cardID string, version uint64, opts ...RequestOption) error {
	return adminRequestJSON(ctx, a.client, http.MethodDelete, projectCardPath(projectID, cardID)+"/versions/"+uintString(version), nil, headersFromOptions(opts), nil)
}

// RestoreProjectCardVersion restores one project card version.
func (a *adminService) RestoreProjectCardVersion(ctx context.Context, projectID, cardID string, version uint64, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectCardPath(projectID, cardID)+"/versions/"+uintString(version)+"/restore", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectExperiments lists project experiments.
func (a *adminService) ListProjectExperiments(ctx context.Context, projectID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "experiments"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProjectExperiment creates a project experiment.
func (a *adminService) CreateProjectExperiment(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "experiments"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectExperiment returns one project experiment.
func (a *adminService) GetProjectExperiment(ctx context.Context, projectID, experimentID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "experiments/"+projectPathSegment(experimentID)), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProjectExperiment patches one project experiment.
func (a *adminService) UpdateProjectExperiment(ctx context.Context, projectID, experimentID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPatch, projectSubpath(projectID, "experiments/"+projectPathSegment(experimentID)), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// StartProjectExperiment starts one project experiment.
func (a *adminService) StartProjectExperiment(ctx context.Context, projectID, experimentID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "experiments/"+projectPathSegment(experimentID)+"/start"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PauseProjectExperiment pauses one project experiment.
func (a *adminService) PauseProjectExperiment(ctx context.Context, projectID, experimentID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "experiments/"+projectPathSegment(experimentID)+"/pause"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// StopProjectExperiment stops one project experiment.
func (a *adminService) StopProjectExperiment(ctx context.Context, projectID, experimentID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "experiments/"+projectPathSegment(experimentID)+"/stop"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectLLM returns the project LLM connection state.
func (a *adminService) GetProjectLLM(ctx context.Context, projectID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "llm"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProjectLLM updates the project LLM provider or key.
func (a *adminService) UpdateProjectLLM(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPut, projectSubpath(projectID, "llm"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteProjectLLM resets the project LLM connection.
func (a *adminService) DeleteProjectLLM(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodDelete, projectSubpath(projectID, "llm"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectVersions lists project versions.
func (a *adminService) ListProjectVersions(ctx context.Context, projectID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "versions")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProjectVersion creates a project version.
func (a *adminService) CreateProjectVersion(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "versions"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectVersion returns one project version.
func (a *adminService) GetProjectVersion(ctx context.Context, projectID, versionID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "versions/"+projectPathSegment(versionID)), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DiffProjectVersion diffs one project version against another.
func (a *adminService) DiffProjectVersion(ctx context.Context, projectID, versionID, against string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	query := map[string]string{"against": against}
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "versions/"+projectPathSegment(versionID)+"/diff")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PublishProjectVersion publishes a project version.
func (a *adminService) PublishProjectVersion(ctx context.Context, projectID, versionID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "versions/"+projectPathSegment(versionID)+"/publish"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectRelease returns the current project release pointer.
func (a *adminService) GetProjectRelease(ctx context.Context, projectID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "release"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectReleases lists project release records.
func (a *adminService) ListProjectReleases(ctx context.Context, projectID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "releases")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RollbackProjectRelease rolls the project release pointer back to a record.
func (a *adminService) RollbackProjectRelease(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "release/rollback"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectSystemPrompts lists project system prompts.
func (a *adminService) ListProjectSystemPrompts(ctx context.Context, projectID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "system-prompts"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProjectSystemPrompt creates a project system prompt.
func (a *adminService) CreateProjectSystemPrompt(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "system-prompts"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetProjectSystemPromptDefault selects the default project system prompt.
func (a *adminService) SetProjectSystemPromptDefault(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPut, projectSubpath(projectID, "system-prompts"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectSystemPrompt returns one project system prompt.
func (a *adminService) GetProjectSystemPrompt(ctx context.Context, projectID, promptID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "system-prompts/"+projectPathSegment(promptID)), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProjectSystemPrompt forks or updates one project system prompt.
func (a *adminService) UpdateProjectSystemPrompt(ctx context.Context, projectID, promptID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPut, projectSubpath(projectID, "system-prompts/"+projectPathSegment(promptID)), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListGlobalSystemPrompts lists global system prompts.
func (a *adminService) ListGlobalSystemPrompts(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/global-system-prompts", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateGlobalSystemPrompt creates a global system prompt.
func (a *adminService) CreateGlobalSystemPrompt(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/global-system-prompts", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetGlobalSystemPromptDefault selects the default global system prompt.
func (a *adminService) SetGlobalSystemPromptDefault(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPut, "/global-system-prompts", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGlobalSystemPrompt returns one global system prompt.
func (a *adminService) GetGlobalSystemPrompt(ctx context.Context, promptID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/global-system-prompts/"+projectPathSegment(promptID), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectUserSessions lists project user-session relationships.
func (a *adminService) ListProjectUserSessions(ctx context.Context, projectID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "user-sessions")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectUserSession returns one project user-session relationship.
func (a *adminService) GetProjectUserSession(ctx context.Context, projectID, sessionID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "user-sessions/"+projectPathSegment(sessionID))+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectIdentityMigration returns the project identity migration status.
func (a *adminService) GetProjectIdentityMigration(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/project-identity-migration", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// StartProjectIdentityMigration starts the project identity switch.
func (a *adminService) StartProjectIdentityMigration(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/project-identity-migration", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PrepareProjectIdentityMigration prepares the legacy identity migration schema.
func (a *adminService) PrepareProjectIdentityMigration(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/project-identity-migration/prepare", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PurgeProjectIdentityMigration purges or previews purging legacy identity keys.
func (a *adminService) PurgeProjectIdentityMigration(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/project-identity-migration/purge", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AdoptProjectIdentityMigration adopts or previews a legacy project identity.
func (a *adminService) AdoptProjectIdentityMigration(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/project-identity-migration/adopt", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RevertProjectIdentityMigration reverts or previews a legacy project identity.
func (a *adminService) RevertProjectIdentityMigration(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/project-identity-migration/revert", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PreviewProjectIdentityMigration previews the legacy project identity migration.
func (a *adminService) PreviewProjectIdentityMigration(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/project-identity-migration/preview", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectRollouts lists project rollouts.
func (a *adminService) ListProjectRollouts(ctx context.Context, projectID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "rollouts")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProjectRollout creates a project rollout.
func (a *adminService) CreateProjectRollout(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "rollouts"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCurrentProjectRollouts returns currently active or scheduled rollouts.
func (a *adminService) GetCurrentProjectRollouts(ctx context.Context, projectID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "rollouts/current"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectRollout returns one project rollout.
func (a *adminService) GetProjectRollout(ctx context.Context, projectID, rolloutID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "rollouts/"+projectPathSegment(rolloutID)), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProjectRollout patches one published project rollout.
func (a *adminService) UpdateProjectRollout(ctx context.Context, projectID, rolloutID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPatch, projectSubpath(projectID, "rollouts/"+projectPathSegment(rolloutID)), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteProjectRollout deletes one project rollout.
func (a *adminService) DeleteProjectRollout(ctx context.Context, projectID, rolloutID string, opts ...RequestOption) error {
	return adminRequestJSON(ctx, a.client, http.MethodDelete, projectSubpath(projectID, "rollouts/"+projectPathSegment(rolloutID)), nil, headersFromOptions(opts), nil)
}

// StopProjectRollout stops one project rollout.
func (a *adminService) StopProjectRollout(ctx context.Context, projectID, rolloutID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "rollouts/"+projectPathSegment(rolloutID)+"/stop"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectRolloutAudits lists audit records for one project rollout.
func (a *adminService) ListProjectRolloutAudits(ctx context.Context, projectID, rolloutID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "rollouts/"+projectPathSegment(rolloutID)+"/audit")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectPresets lists console sampling presets for a project.
func (a *adminService) ListProjectPresets(ctx context.Context, projectID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "presets"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProjectPreset creates a console sampling preset for a project.
func (a *adminService) CreateProjectPreset(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "presets"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProjectPreset updates a console sampling preset for a project.
func (a *adminService) UpdateProjectPreset(ctx context.Context, projectID, presetID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPut, projectSubpath(projectID, "presets/"+projectPathSegment(presetID)), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PublishProjectPreset publishes a console preset to the project live sample.
func (a *adminService) PublishProjectPreset(ctx context.Context, projectID, presetID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "presets/"+projectPathSegment(presetID)+"/publish"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectSessions lists console sessions for a project.
func (a *adminService) ListProjectSessions(ctx context.Context, projectID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "sessions")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProjectSession archives or unarchives a console session.
func (a *adminService) UpdateProjectSession(ctx context.Context, projectID, sessionID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPatch, projectSubpath(projectID, "sessions/"+projectPathSegment(sessionID)), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectSuites lists evaluation suites for a project.
func (a *adminService) ListProjectSuites(ctx context.Context, projectID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "suites")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProjectSuite creates an evaluation suite for a project.
func (a *adminService) CreateProjectSuite(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "suites"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectSuite returns one evaluation suite for a project.
func (a *adminService) GetProjectSuite(ctx context.Context, projectID, suiteID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "suites/"+projectPathSegment(suiteID)), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectEvaluations lists evaluation runs for a project.
func (a *adminService) ListProjectEvaluations(ctx context.Context, projectID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "evaluations")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectEvaluation returns one evaluation run for a project.
func (a *adminService) GetProjectEvaluation(ctx context.Context, projectID, evaluationID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "evaluations/"+projectPathSegment(evaluationID)), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CompareProjectEvaluation compares one evaluation run with another.
func (a *adminService) CompareProjectEvaluation(ctx context.Context, projectID, evaluationID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "evaluations/"+projectPathSegment(evaluationID)+"/compare"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CancelProjectEvaluation cancels one evaluation run.
func (a *adminService) CancelProjectEvaluation(ctx context.Context, projectID, evaluationID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "evaluations/"+projectPathSegment(evaluationID)+"/cancel"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ResumeProjectEvaluation resumes one evaluation run.
func (a *adminService) ResumeProjectEvaluation(ctx context.Context, projectID, evaluationID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "evaluations/"+projectPathSegment(evaluationID)+"/resume"), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjectFeedback lists evaluation feedback for a project.
func (a *adminService) ListProjectFeedback(ctx context.Context, projectID string, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, projectSubpath(projectID, "feedback")+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProjectFeedback creates evaluation feedback for a project.
func (a *adminService) CreateProjectFeedback(ctx context.Context, projectID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, projectSubpath(projectID, "feedback"), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ProjectEngine proxies a request to the project engine API at
// /admin/v1/projects/{projectID}/engine/{enginePath}.
func (a *adminService) ProjectEngine(ctx context.Context, projectID, method, enginePath string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	return a.ProjectRequest(ctx, projectID, method, "engine/"+enginePath, body, opts...)
}

// ListAdminCards lists all project card metadata visible to an administrator.
func (a *adminService) ListAdminCards(ctx context.Context, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/cards"+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TranslationsQueue returns the current translation queue, optionally filtered
// by project_id.
func (a *adminService) TranslationsQueue(ctx context.Context, query map[string]string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/translations/queue"+adminQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TranslationsCallback sends a translation callback body to the engine.
func (a *adminService) TranslationsCallback(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/translations/callback", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func projectPathSegment(id string) string {
	return url.PathEscape(id)
}
func projectSubpath(projectID, subpath string) string {
	path := "/projects/" + projectPathSegment(projectID)
	if subpath != "" {
		path += "/" + subpath
	}
	return path
}

func projectCardPath(projectID, cardID string) string {
	return projectSubpath(projectID, "cards/"+projectPathSegment(cardID))
}

func adminQueryString(query map[string]string) string {
	if len(query) == 0 {
		return ""
	}
	values := url.Values{}
	for key, value := range query {
		if value != "" {
			values.Set(key, value)
		}
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

func adminRawRequest(ctx context.Context, c *transport.Client, method, path string, body any, headers http.Header) ([]byte, error) {
	status, payload, err := c.Request(ctx, method, path, body, headers)
	if err != nil {
		return nil, publicError(err)
	}
	if status >= 400 {
		return nil, parseAdminError(status, payload)
	}
	return payload, nil
}

func adminRequestJSON(ctx context.Context, c *transport.Client, method, path string, body any, headers http.Header, out any) error {
	status, payload, err := c.Request(ctx, method, path, body, headers)
	if err != nil {
		return publicError(err)
	}
	if status >= 400 {
		return parseAdminError(status, payload)
	}
	if out == nil || len(payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return &Error{Kind: ErrGeneral, Message: "failed to decode response: " + err.Error()}
	}
	return nil
}

func parseAdminError(status int, payload []byte) error {
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(payload, &envelope); err == nil && (envelope.Error.Code != "" || envelope.Error.Message != "") {
		return newHTTPErrorWithCode(status, envelope.Error.Code, adminErrorMessage(status, envelope.Error.Message))
	}
	return parseRawResponse(status, payload)
}

func adminErrorMessage(status int, message string) string {
	if message == "" {
		return http.StatusText(status)
	}
	return message
}
