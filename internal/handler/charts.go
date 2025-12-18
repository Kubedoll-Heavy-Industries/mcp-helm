package handler

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/mcputil"
)

// Default pagination limits
const (
	defaultChartListLimit = 50
	defaultValuesMaxLines = 200
)

// Input/output types for chart tools

type listRepositoryChartsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum charts to return (default 50, 0 = unlimited)"`
	Offset        int    `json:"offset,omitempty" jsonschema:"Number of charts to skip"`
	Search        string `json:"search,omitempty" jsonschema:"Filter charts by name (case-insensitive substring match)"`
}

type listRepositoryChartsOutput struct {
	Charts []string `json:"charts" jsonschema:"Chart names"`
	Total  int      `json:"total" jsonschema:"Total charts before filtering/pagination"`
	Limit  int      `json:"limit" jsonschema:"Applied limit"`
	Offset int      `json:"offset" jsonschema:"Applied offset"`
}

type getChartValuesInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string `json:"chart_version,omitempty" jsonschema:"Chart version (latest if omitted)"`
	Path          string `json:"path,omitempty" jsonschema:"YAML path to extract (e.g., '.prometheus.server' or '.ingress.enabled'). Returns full file if omitted."`
	MaxLines      int    `json:"max_lines,omitempty" jsonschema:"Maximum lines to return (default 200, 0 = unlimited)"`
}

type getChartValuesOutput struct {
	Values    string `json:"values" jsonschema:"Values content (YAML)"`
	Truncated bool   `json:"truncated" jsonschema:"Whether output was truncated"`
	Path      string `json:"path,omitempty" jsonschema:"Path that was extracted, if any"`
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

type listChartContentsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string `json:"chart_version,omitempty" jsonschema:"Chart version (latest if omitted)"`
	Pattern       string `json:"pattern,omitempty" jsonschema:"Glob pattern to filter files (e.g., 'templates/*.yaml')"`
}

type fileInfo struct {
	Path string `json:"path" jsonschema:"File path within chart"`
	Size int    `json:"size" jsonschema:"File size in bytes"`
}

type listChartContentsOutput struct {
	Files []fileInfo `json:"files" jsonschema:"Files in the chart"`
	Total int        `json:"total" jsonschema:"Total number of files"`
}

type getChartContentInput struct {
	RepositoryURL string   `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string   `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string   `json:"chart_version,omitempty" jsonschema:"Chart version (latest if omitted)"`
	Files         []string `json:"files" jsonschema:"File paths to retrieve (e.g., ['templates/deployment.yaml', 'values.yaml'])"`
}

type fileContent struct {
	Path    string `json:"path" jsonschema:"File path"`
	Content string `json:"content" jsonschema:"File content"`
}

type getChartContentOutput struct {
	Files []fileContent `json:"files" jsonschema:"Requested file contents"`
}

type getChartDependenciesInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string `json:"chart_version,omitempty" jsonschema:"Chart version (latest if omitted)"`
}

type getChartDependenciesOutput struct {
	Dependencies []dependencyInfo `json:"dependencies" jsonschema:"Chart dependencies"`
}

type dependencyInfo struct {
	Name       string `json:"name" jsonschema:"Dependency name"`
	Version    string `json:"version" jsonschema:"Version constraint"`
	Repository string `json:"repository,omitempty" jsonschema:"Repository URL (empty for bundled)"`
	Condition  string `json:"condition,omitempty" jsonschema:"Condition for enabling"`
	Alias      string `json:"alias,omitempty" jsonschema:"Alias name"`
}

// Handler implementations

func (h *Handler) listRepositoryCharts() mcp.ToolHandlerFor[listRepositoryChartsInput, listRepositoryChartsOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listRepositoryChartsInput) (*mcp.CallToolResult, listRepositoryChartsOutput, error) {
		emptyOutput := listRepositoryChartsOutput{Charts: []string{}}

		if err := validateRequired(map[string]string{"repository_url": in.RepositoryURL}); err != nil {
			return err, emptyOutput, nil
		}

		if in.Limit < 0 {
			return mcputil.TextError("limit must be >= 0"), emptyOutput, nil
		}
		if in.Offset < 0 {
			return mcputil.TextError("offset must be >= 0"), emptyOutput, nil
		}

		charts, err := h.svc.ListCharts(ctx, strings.TrimSpace(in.RepositoryURL))
		if err != nil {
			return mcputil.HandleError(err), emptyOutput, nil
		}

		// Apply search filter
		if in.Search != "" {
			search := strings.ToLower(in.Search)
			filtered := make([]string, 0)
			for _, c := range charts {
				if strings.Contains(strings.ToLower(c), search) {
					filtered = append(filtered, c)
				}
			}
			charts = filtered
		}

		total := len(charts)

		// Apply pagination with default limit
		limit := in.Limit
		if limit == 0 {
			limit = defaultChartListLimit
		}

		start := in.Offset
		if start > total {
			start = total
		}
		end := total
		if limit > 0 && start+limit < end {
			end = start + limit
		}

		return nil, listRepositoryChartsOutput{
			Charts: charts[start:end],
			Total:  total,
			Limit:  limit,
			Offset: in.Offset,
		}, nil
	}
}

func (h *Handler) getChartValues() mcp.ToolHandlerFor[getChartValuesInput, getChartValuesOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getChartValuesInput) (*mcp.CallToolResult, getChartValuesOutput, error) {
		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, getChartValuesOutput{}, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)
		path := strings.TrimSpace(in.Path)

		version, err := h.resolveVersion(ctx, repo, chart, in.ChartVersion)
		if err != nil {
			return mcputil.HandleError(err), getChartValuesOutput{}, nil
		}

		valuesBytes, err := h.svc.GetValues(ctx, repo, chart, version)
		if err != nil {
			return mcputil.HandleError(err), getChartValuesOutput{}, nil
		}

		var result string
		var extractedPath string

		// If path is specified, extract that portion using yq-style path
		if path != "" {
			extracted, err := extractYAMLPath(valuesBytes, path)
			if err != nil {
				return mcputil.TextError(fmt.Sprintf("invalid path %q: %v", path, err)), getChartValuesOutput{}, nil
			}
			result = extracted
			extractedPath = path
		} else {
			result = string(valuesBytes)
		}

		// Apply max_lines truncation
		maxLines := in.MaxLines
		if maxLines == 0 {
			maxLines = defaultValuesMaxLines
		}

		truncated := false
		if maxLines > 0 {
			result, truncated = truncateLines(result, maxLines)
		}

		return nil, getChartValuesOutput{
			Values:    result,
			Truncated: truncated,
			Path:      extractedPath,
		}, nil
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

func (h *Handler) listChartContents() mcp.ToolHandlerFor[listChartContentsInput, listChartContentsOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listChartContentsInput) (*mcp.CallToolResult, listChartContentsOutput, error) {
		emptyOutput := listChartContentsOutput{Files: []fileInfo{}}

		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, emptyOutput, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)
		pattern := strings.TrimSpace(in.Pattern)

		version, err := h.resolveVersion(ctx, repo, chart, in.ChartVersion)
		if err != nil {
			return mcputil.HandleError(err), emptyOutput, nil
		}

		files, err := h.svc.ListFiles(ctx, repo, chart, version)
		if err != nil {
			return mcputil.HandleError(err), emptyOutput, nil
		}

		// Apply pattern filter if specified
		result := make([]fileInfo, 0, len(files))
		for _, f := range files {
			if pattern != "" {
				matched, err := filepath.Match(pattern, f.Path)
				if err != nil {
					return mcputil.TextError(fmt.Sprintf("invalid pattern %q: %v", pattern, err)), emptyOutput, nil
				}
				if !matched {
					// Also try matching against just the filename
					matched, _ = filepath.Match(pattern, filepath.Base(f.Path))
					if !matched {
						continue
					}
				}
			}
			result = append(result, fileInfo{Path: f.Path, Size: f.Size})
		}

		return nil, listChartContentsOutput{
			Files: result,
			Total: len(result),
		}, nil
	}
}

func (h *Handler) getChartContent() mcp.ToolHandlerFor[getChartContentInput, getChartContentOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getChartContentInput) (*mcp.CallToolResult, getChartContentOutput, error) {
		emptyOutput := getChartContentOutput{Files: []fileContent{}}

		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, emptyOutput, nil
		}

		if len(in.Files) == 0 {
			return mcputil.TextError("files is required (specify which files to retrieve)"), emptyOutput, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		version, err := h.resolveVersion(ctx, repo, chart, in.ChartVersion)
		if err != nil {
			return mcputil.HandleError(err), emptyOutput, nil
		}

		result := make([]fileContent, 0, len(in.Files))
		for _, path := range in.Files {
			path = strings.TrimSpace(path)
			if path == "" {
				continue
			}

			content, err := h.svc.GetFile(ctx, repo, chart, version, path)
			if err != nil {
				// Include error in response rather than failing entirely
				result = append(result, fileContent{
					Path:    path,
					Content: fmt.Sprintf("error: %v", err),
				})
				continue
			}

			result = append(result, fileContent{
				Path:    path,
				Content: string(content),
			})
		}

		return nil, getChartContentOutput{Files: result}, nil
	}
}

func (h *Handler) getChartDependencies() mcp.ToolHandlerFor[getChartDependenciesInput, getChartDependenciesOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getChartDependenciesInput) (*mcp.CallToolResult, getChartDependenciesOutput, error) {
		emptyOutput := getChartDependenciesOutput{Dependencies: []dependencyInfo{}}

		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, emptyOutput, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		version, err := h.resolveVersion(ctx, repo, chart, in.ChartVersion)
		if err != nil {
			return mcputil.HandleError(err), emptyOutput, nil
		}

		deps, err := h.svc.GetDependencies(ctx, repo, chart, version)
		if err != nil {
			return mcputil.HandleError(err), emptyOutput, nil
		}

		// Convert to output format
		result := make([]dependencyInfo, 0, len(deps))
		for _, d := range deps {
			result = append(result, dependencyInfo{
				Name:       d.Name,
				Version:    d.Version,
				Repository: d.Repository,
				Condition:  d.Condition,
				Alias:      d.Alias,
			})
		}

		return nil, getChartDependenciesOutput{Dependencies: result}, nil
	}
}

// Helper functions

// extractYAMLPath extracts a value at the given yq-style path from YAML data.
// Supports paths like ".foo.bar" or ".foo.bar[0]"
func extractYAMLPath(data []byte, path string) (string, error) {
	// Parse the YAML
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	if root.Kind == 0 {
		return "", fmt.Errorf("empty YAML document")
	}

	// Navigate to the path
	node := &root
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		node = root.Content[0]
	}

	// Parse path components (simplified yq-style: .foo.bar)
	path = strings.TrimPrefix(path, ".")
	if path == "" {
		// Return the whole document
		out, err := yaml.Marshal(node)
		return string(out), err
	}

	parts := strings.Split(path, ".")
	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle array index like "items[0]"
		var arrayIndex int = -1
		if idx := strings.Index(part, "["); idx != -1 {
			end := strings.Index(part, "]")
			if end > idx {
				fmt.Sscanf(part[idx+1:end], "%d", &arrayIndex)
				part = part[:idx]
			}
		}

		if node.Kind != yaml.MappingNode {
			return "", fmt.Errorf("cannot navigate into non-mapping node at %q", part)
		}

		// Find the key in the mapping
		found := false
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == part {
				node = node.Content[i+1]
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("path not found: %q", part)
		}

		// Handle array indexing
		if arrayIndex >= 0 {
			if node.Kind != yaml.SequenceNode {
				return "", fmt.Errorf("cannot index non-sequence node at %q", part)
			}
			if arrayIndex >= len(node.Content) {
				return "", fmt.Errorf("index %d out of range for %q (len=%d)", arrayIndex, part, len(node.Content))
			}
			node = node.Content[arrayIndex]
		}
	}

	// Marshal the result back to YAML
	out, err := yaml.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// truncateLines truncates content to maxLines and returns whether truncation occurred.
func truncateLines(content string, maxLines int) (string, bool) {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content, false
	}

	truncated := strings.Join(lines[:maxLines], "\n")
	truncated += fmt.Sprintf("\n\n... truncated (%d more lines)", len(lines)-maxLines)
	return truncated, true
}

