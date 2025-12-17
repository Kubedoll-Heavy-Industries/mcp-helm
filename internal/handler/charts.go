package handler

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/mcputil"
)

// Input/output types for chart tools

type listChartsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
}

type listChartsOutput struct {
	Charts []string `json:"charts,omitempty" jsonschema:"Chart names available in the repository"`
}

type getValuesInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string `json:"chart_version,omitempty" jsonschema:"Chart version (latest if omitted)"`
}

type getValuesOutput struct {
	Values string `json:"values" jsonschema:"Contents of values.yaml"`
}

type getValuesSchemaInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string `json:"chart_version,omitempty" jsonschema:"Chart version (latest if omitted)"`
}

type getValuesSchemaOutput struct {
	Schema  string `json:"schema" jsonschema:"Contents of values.schema.json"`
	Present bool   `json:"present" jsonschema:"Whether the schema file exists"`
}

type getContentsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string `json:"chart_version,omitempty" jsonschema:"Chart version (latest if omitted)"`
	Recursive     bool   `json:"recursive,omitempty" jsonschema:"Include sub-chart contents"`
}

type getContentsOutput struct {
	Contents string `json:"contents" jsonschema:"Formatted chart contents"`
}

type getDependenciesInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string `json:"chart_version,omitempty" jsonschema:"Chart version (latest if omitted)"`
}

type getDependenciesOutput struct {
	Dependencies []string `json:"dependencies" jsonschema:"Chart dependencies as JSON strings"`
}

// Handler implementations

func (h *Handler) listCharts() mcp.ToolHandlerFor[listChartsInput, listChartsOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listChartsInput) (*mcp.CallToolResult, listChartsOutput, error) {
		emptyOutput := listChartsOutput{Charts: []string{}}

		if err := validateRequired(map[string]string{"repository_url": in.RepositoryURL}); err != nil {
			return err, emptyOutput, nil
		}

		charts, err := h.svc.ListCharts(ctx, strings.TrimSpace(in.RepositoryURL))
		if err != nil {
			return mcputil.HandleError(err), emptyOutput, nil
		}

		return nil, listChartsOutput{Charts: charts}, nil
	}
}

func (h *Handler) getValues() mcp.ToolHandlerFor[getValuesInput, getValuesOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getValuesInput) (*mcp.CallToolResult, getValuesOutput, error) {
		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, getValuesOutput{}, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		version, err := h.resolveVersion(ctx, repo, chart, in.ChartVersion)
		if err != nil {
			return mcputil.HandleError(err), getValuesOutput{}, nil
		}

		values, err := h.svc.GetValues(ctx, repo, chart, version)
		if err != nil {
			return mcputil.HandleError(err), getValuesOutput{}, nil
		}

		return nil, getValuesOutput{Values: string(values)}, nil
	}
}

func (h *Handler) getValuesSchema() mcp.ToolHandlerFor[getValuesSchemaInput, getValuesSchemaOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getValuesSchemaInput) (*mcp.CallToolResult, getValuesSchemaOutput, error) {
		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, getValuesSchemaOutput{}, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		version, err := h.resolveVersion(ctx, repo, chart, in.ChartVersion)
		if err != nil {
			return mcputil.HandleError(err), getValuesSchemaOutput{}, nil
		}

		schema, present, err := h.svc.GetValuesSchema(ctx, repo, chart, version)
		if err != nil {
			return mcputil.HandleError(err), getValuesSchemaOutput{}, nil
		}

		return nil, getValuesSchemaOutput{Schema: string(schema), Present: present}, nil
	}
}

func (h *Handler) getContents() mcp.ToolHandlerFor[getContentsInput, getContentsOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getContentsInput) (*mcp.CallToolResult, getContentsOutput, error) {
		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, getContentsOutput{}, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		version, err := h.resolveVersion(ctx, repo, chart, in.ChartVersion)
		if err != nil {
			return mcputil.HandleError(err), getContentsOutput{}, nil
		}

		contents, err := h.svc.GetContents(ctx, repo, chart, version, in.Recursive)
		if err != nil {
			return mcputil.HandleError(err), getContentsOutput{}, nil
		}

		return nil, getContentsOutput{Contents: contents}, nil
	}
}

func (h *Handler) getDependencies() mcp.ToolHandlerFor[getDependenciesInput, getDependenciesOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getDependenciesInput) (*mcp.CallToolResult, getDependenciesOutput, error) {
		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, getDependenciesOutput{}, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		version, err := h.resolveVersion(ctx, repo, chart, in.ChartVersion)
		if err != nil {
			return mcputil.HandleError(err), getDependenciesOutput{}, nil
		}

		deps, err := h.svc.GetDependencies(ctx, repo, chart, version)
		if err != nil {
			return mcputil.HandleError(err), getDependenciesOutput{}, nil
		}

		// Convert to JSON strings for backward compatibility
		depStrings := make([]string, 0, len(deps))
		for _, d := range deps {
			b, _ := json.Marshal(d)
			depStrings = append(depStrings, string(b))
		}

		return nil, getDependenciesOutput{Dependencies: depStrings}, nil
	}
}
