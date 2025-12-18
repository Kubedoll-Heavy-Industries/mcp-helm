package helm

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
)

// chartYAML represents the structure of Chart.yaml for dependency parsing.
type chartYAML struct {
	Dependencies []Dependency `yaml:"dependencies"`
}

// extractDependencies parses a chart's dependencies from Chart.yaml.
// It recursively extracts dependencies from sub-charts as well.
func extractDependencies(chart *chartv2.Chart) ([]Dependency, error) {
	var rawYAML []byte
	for _, file := range chart.Raw {
		if file.Name == "Chart.yaml" {
			rawYAML = file.Data
			break
		}
	}

	if len(rawYAML) == 0 {
		return nil, fmt.Errorf("chart.yaml not found in chart %s", chart.Name())
	}

	var cy chartYAML
	if err := yaml.Unmarshal(rawYAML, &cy); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	if len(cy.Dependencies) == 0 {
		return nil, nil
	}

	// Validate dependencies have required fields
	// Per Helm spec: name and version are required, repository is optional
	// (empty repository means the dependency is bundled in charts/ directory)
	for i, dep := range cy.Dependencies {
		if dep.Name == "" {
			return nil, fmt.Errorf("dependency %d missing required field: name", i)
		}
		if dep.Version == "" {
			return nil, fmt.Errorf("dependency %d (%s) missing required field: version", i, dep.Name)
		}
		// Note: repository can be empty for bundled dependencies
	}

	// Recursively get sub-chart dependencies
	subCharts := chart.Dependencies()
	result := make([]Dependency, 0, len(cy.Dependencies)*2)

	for _, dep := range cy.Dependencies {
		result = append(result, dep)

		// Find matching sub-chart and get its dependencies
		for _, sub := range subCharts {
			if sub.Name() == dep.Name && sub.Metadata.Version == dep.Version {
				subDeps, err := extractDependencies(sub)
				if err != nil {
					return nil, fmt.Errorf("failed to get dependencies for sub-chart %s: %w", sub.Name(), err)
				}
				result = append(result, subDeps...)
				break
			}
		}
	}

	return result, nil
}

// formatDependenciesJSON returns dependencies as JSON strings for backward compatibility.
func formatDependenciesJSON(deps []Dependency) ([]string, error) {
	result := make([]string, 0, len(deps))
	for _, dep := range deps {
		b, err := json.Marshal(dep)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dependency: %w", err)
		}
		result = append(result, string(b))
	}
	return result, nil
}

// formatContents returns a formatted string of all chart files.
func formatContents(chart *chartv2.Chart, recursive bool) string {
	var sb strings.Builder

	// Raw files (Chart.yaml, values.yaml, etc.)
	for _, file := range chart.Raw {
		sb.WriteString(fmt.Sprintf("# file: %s/%s\n", chart.Name(), file.Name))
		sb.Write(file.Data)
		sb.WriteString("\n\n")
	}

	// Template files
	for _, file := range chart.Templates {
		sb.WriteString(fmt.Sprintf("# file: %s/%s\n", chart.Name(), file.Name))
		sb.Write(file.Data)
		sb.WriteString("\n\n")
	}

	// Extra files
	for _, file := range chart.Files {
		sb.WriteString(fmt.Sprintf("# file: %s/%s\n", chart.Name(), file.Name))
		sb.Write(file.Data)
		sb.WriteString("\n\n")
	}

	// Recursively include sub-charts
	if recursive {
		for _, sub := range chart.Dependencies() {
			sb.WriteString(fmt.Sprintf("# Subchart: %s\n", sub.Name()))
			sb.WriteString(formatContents(sub, recursive))
		}
	}

	return sb.String()
}

// listChartFiles returns a list of all file paths in the chart.
func listChartFiles(chart *chartv2.Chart) []string {
	var files []string

	for _, f := range chart.Raw {
		files = append(files, f.Name)
	}
	for _, f := range chart.Files {
		files = append(files, f.Name)
	}
	for _, f := range chart.Templates {
		files = append(files, f.Name)
	}

	return files
}

// findChartFile searches for a file in the chart and returns its contents.
func findChartFile(chart *chartv2.Chart, path string) ([]byte, bool) {
	// Check raw files
	for _, f := range chart.Raw {
		if f.Name == path {
			return f.Data, true
		}
	}

	// Check extra files
	for _, f := range chart.Files {
		if f.Name == path {
			return f.Data, true
		}
	}

	// Check templates with flexible path matching
	for _, f := range chart.Templates {
		if f.Name == path {
			return f.Data, true
		}
		// Try with templates/ prefix
		if f.Name == "templates/"+path {
			return f.Data, true
		}
		// Try stripping templates/ prefix
		if strings.TrimPrefix(f.Name, "templates/") == strings.TrimPrefix(path, "templates/") {
			return f.Data, true
		}
	}

	return nil, false
}

// convertToChart converts a Helm SDK chart to our domain Chart type.
func convertToChart(hc *chartv2.Chart) *Chart {
	c := &Chart{
		Name:    hc.Name(),
		Version: hc.Metadata.Version,
	}

	for _, f := range hc.Raw {
		c.Raw = append(c.Raw, File{Name: f.Name, Data: f.Data})
	}
	for _, f := range hc.Templates {
		c.Templates = append(c.Templates, File{Name: f.Name, Data: f.Data})
	}
	for _, f := range hc.Files {
		c.Files = append(c.Files, File{Name: f.Name, Data: f.Data})
	}

	return c
}
