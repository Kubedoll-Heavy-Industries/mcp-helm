package helm

import (
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
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
