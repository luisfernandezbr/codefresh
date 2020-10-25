package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

const workflowETagStateKey = "workflow.etag"

// CodefreshIntegration is an integration for Codefresh
type CodefreshIntegration struct {
	config  sdk.Config
	manager sdk.Manager
	refType string
}

var _ sdk.Integration = (*CodefreshIntegration)(nil)

// Start is called when the integration is starting up
func (g *CodefreshIntegration) Start(logger sdk.Logger, config sdk.Config, manager sdk.Manager) error {
	g.config = config
	g.manager = manager
	g.refType = "codefresh"
	sdk.LogInfo(logger, "starting")
	return nil
}

// Stop is called when the integration is shutting down for cleanup
func (g *CodefreshIntegration) Stop(logger sdk.Logger) error {
	sdk.LogInfo(logger, "stopping")
	return nil
}

// Enroll is called when a new integration instance is added
func (g *CodefreshIntegration) Enroll(instance sdk.Instance) error {
	return nil
}

// Dismiss is called when an existing integration instance is removed
func (g *CodefreshIntegration) Dismiss(instance sdk.Instance) error {
	return nil
}

// WebHook is called when a webhook is received on behalf of the integration
func (g *CodefreshIntegration) WebHook(webhook sdk.WebHook) error {
	return nil
}

// Mutation is called when a mutation request is received on behalf of the integration
func (g *CodefreshIntegration) Mutation(mutation sdk.Mutation) (*sdk.MutationResponse, error) {
	return nil, nil
}

func (g *CodefreshIntegration) fetchBuilds(logger sdk.Logger, pipe sdk.Pipe, export sdk.Export, apiKey string) error {

	customerID := export.CustomerID()
	integrationInstanceID := export.IntegrationInstanceID()
	state := export.State()

	var etag string
	if !export.Historical() {
		_, err := state.Get(workflowETagStateKey, &etag)
		if err != nil {
			return err
		}
	}

	httpmanager := g.manager.HTTPManager().New("https://g.codefresh.io/api/workflow", map[string]string{"Authorization": apiKey})

	var page int
	var savedETag string
	var total int
	var retries int

	qs := url.Values{}
	started := time.Now()

	inclusions := export.Config().Inclusions
	exclusions := export.Config().Exclusions

	for {
		opts := make([]sdk.WithHTTPOption, 0)
		if etag != "" {
			opts = append(opts, sdk.WithHTTPHeader("If-None-Match", etag))
			etag = ""
		}
		qs.Set("page", strconv.Itoa(page))
		qs.Set("limit", "25")
		opts = append(opts, sdk.WithGetQueryParameters(qs))
		var result workflowResult
		sdk.LogInfo(logger, "starting build fetch", "page", page)
		ts := time.Now()
		resp, err := httpmanager.Get(&result, opts...)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") && retries < 5 {
				retries++
				time.Sleep(time.Second * time.Duration(5*retries))
				continue
			}
			return err
		}
		if savedETag == "" {
			savedETag = resp.Headers.Get("ETag")
		}
		sdk.LogInfo(logger, "fetched build result", "page", page, "status", resp.StatusCode, "pages", result.Workflows.Pages, "total", result.Workflows.Total, "duration", time.Since(ts))
		if resp.StatusCode == http.StatusNotModified {
			sdk.LogInfo(logger, "no builds found")
			break
		}
		for _, build := range result.Workflows.Builds {
			if inclusions != nil && !inclusions.Matches("codefresh", build.Pipeline) {
				sdk.LogInfo(logger, "skipping build based on inclusion rule not matched", "id", build.ID, "pipeline", build.Pipeline)
				continue
			}
			if exclusions != nil && exclusions.Matches("codefresh", build.Pipeline) {
				sdk.LogInfo(logger, "skipping build based on exclusion rule matched", "id", build.ID, "pipeline", build.Pipeline)
				continue
			}
			sdk.LogDebug(logger, "processing build", "id", build.ID, "status", build.Status)
			model, err := build.ToModel(g.refType, customerID, integrationInstanceID)
			if err != nil {
				sdk.LogError(logger, "error creating model from build", "err", err, "id", build.ID)
				return err
			}
			if err := pipe.Write(model); err != nil {
				return err
			}
			total++
		}
		page++
		if page >= result.Workflows.Pages {
			break
		}
		pipe.Flush() // some of these pages can take a while so we flush on each page
	}

	if err := state.Set(workflowETagStateKey, savedETag); err != nil {
		return fmt.Errorf("error saving state key: %w", err)
	}

	sdk.LogInfo(logger, "finished fetching", "total", total, "duration", time.Since(started))

	return nil
}

// Export is called to tell the integration to run an export
func (g *CodefreshIntegration) Export(export sdk.Export) error {
	logger := export.Logger()

	started := time.Now()
	sdk.LogInfo(logger, "export started")

	// Pipe must be called to begin an export and receive a pipe for sending data
	pipe := export.Pipe()

	// Config is any customer specific configuration for this customer
	config := export.Config()

	sdk.LogDebug(logger, "export starting")

	if config.APIKeyAuth == nil {
		return fmt.Errorf("error missing apikey")
	}

	if err := g.fetchBuilds(logger, pipe, export, config.APIKeyAuth.APIKey); err != nil {
		return err
	}

	sdk.LogInfo(logger, "export finished", "duration", time.Since(started))
	return nil
}

// AutoConfigure is called when a cloud integration has requested to be auto configured
func (g *CodefreshIntegration) AutoConfigure(autoconfig sdk.AutoConfigure) (*sdk.Config, error) {
	return nil, nil
}

// Validate is called before a new integration instance is added to determine
// if the config is valid and the integration can properly communicate with the
// source system. The result and the error will both be delivered to the App.
// Returning a nil error is considered a successful validation.
func (g *CodefreshIntegration) Validate(validate sdk.Validate) (result map[string]interface{}, err error) {
	if validate.Config().APIKeyAuth != nil {
		apiKey := validate.Config().APIKeyAuth.APIKey
		httpmanager := g.manager.HTTPManager().New("https://g.codefresh.io/api/workflow?limit=1", map[string]string{"Authorization": apiKey})
		var r interface{}
		resp, err := httpmanager.Get(&r)
		if err != nil {
			return nil, fmt.Errorf("error validating API Key: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error validating API Key (status code=%d)", resp.StatusCode)
		}
		return nil, nil
	}
	return nil, fmt.Errorf("required API Key not found")
}
