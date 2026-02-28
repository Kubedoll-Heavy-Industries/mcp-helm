package handler

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollapseYAML_UnlimitedDepth(t *testing.T) {
	input := `name: nginx
version: 1.0.0
`
	opts := CollapseOptions{MaxDepth: 0, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.False(t, collapsed, "Unlimited depth should not be collapsed")
	assert.Equal(t, input, result)
}

func TestCollapseYAML_Depth1_ShowDefaults(t *testing.T) {
	input := `replicaCount: 3
image:
  repository: nginx
  tag: latest
service:
  type: ClusterIP
  port: 80
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "replicaCount: 3")
	assert.Contains(t, result, "image: object (2 keys)")
	assert.Contains(t, result, "service: object (2 keys)")
}

func TestCollapseYAML_Depth1_NoDefaults(t *testing.T) {
	input := `enabled: true
count: 5
name: test
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: false, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "enabled: boolean")
	assert.Contains(t, result, "count: number")
	assert.Contains(t, result, "name: string")
}

func TestCollapseYAML_Depth2(t *testing.T) {
	input := `persistence:
  enabled: true
  size: 10Gi
  advanced:
    storageClass: fast
    accessModes:
      - ReadWriteOnce
`
	opts := CollapseOptions{MaxDepth: 2, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "persistence:")
	assert.Contains(t, result, "enabled: true")
	assert.Contains(t, result, "size: 10Gi")
	assert.Contains(t, result, "advanced: object (2 keys)")
}

func TestCollapseYAML_EmptyObject(t *testing.T) {
	input := `annotations: {}
labels: {}
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "annotations: object (empty)")
	assert.Contains(t, result, "labels: object (empty)")
}

func TestCollapseYAML_ArraySummarized(t *testing.T) {
	input := `ports:
  - 80
  - 443
  - 8080
volumes: []
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "ports: array (3 items)")
	assert.Contains(t, result, "volumes: array (empty)")
}

func TestCollapseYAML_ArrayTruncation(t *testing.T) {
	input := `items:
  - one
  - two
  - three
  - four
  - five
  - six
`
	// Depth 2 expands the array, default MaxArrayItems=3 truncates
	opts := DefaultCollapseOptions()
	opts.MaxDepth = 2

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "- one")
	assert.Contains(t, result, "- two")
	assert.Contains(t, result, "- three")
	assert.Contains(t, result, "... and 3 more items")
	assert.NotContains(t, result, "- four")
}

func TestCollapseYAML_ArrayTruncationUnlimited(t *testing.T) {
	input := `items:
  - one
  - two
  - three
  - four
  - five
`
	opts := CollapseOptions{
		MaxDepth:      2,
		MaxArrayItems: 0, // unlimited
		ShowDefaults:  true,
		ShowComments:  true,
	}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "- one")
	assert.Contains(t, result, "- five")
	assert.NotContains(t, result, "... and")
}

func TestCollapseYAML_ArrayTruncationCustomLimit(t *testing.T) {
	input := `items:
  - one
  - two
  - three
  - four
  - five
`
	opts := CollapseOptions{
		MaxDepth:      2,
		MaxArrayItems: 2, // Only show 2 items
		ShowDefaults:  true,
		ShowComments:  true,
	}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "- one")
	assert.Contains(t, result, "- two")
	assert.Contains(t, result, "... and 3 more items")
	assert.NotContains(t, result, "- three")
}

func TestCollapseYAML_NullValue(t *testing.T) {
	input := `optional: null
required: value
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "optional: null")
	assert.Contains(t, result, "required: value")
}

func TestCollapseYAML_BooleanValues(t *testing.T) {
	input := `enabled: true
disabled: false
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "enabled: true")
	assert.Contains(t, result, "disabled: false")
}

func TestCollapseYAML_QuotedStrings(t *testing.T) {
	input := `empty: ""
colon: "host:port"
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, `empty: ""`)
	assert.Contains(t, result, `colon: "host:port"`)
}

func TestCollapseYAML_StripComments(t *testing.T) {
	input := `# This is a comment
name: test
# Another comment
value: 123
`
	opts := CollapseOptions{MaxDepth: 0, ShowDefaults: true, ShowComments: false}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.False(t, collapsed)
	assert.NotContains(t, result, "# This is a comment")
	assert.Contains(t, result, "name: test")
}

func TestCollapseYAML_PreservesKeyOrder(t *testing.T) {
	input := `zebra: 1
apple: 2
mango: 3
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, _, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)

	// Verify original order is preserved (not sorted alphabetically)
	zebraIdx := strings.Index(result, "zebra:")
	appleIdx := strings.Index(result, "apple:")
	mangoIdx := strings.Index(result, "mango:")

	assert.True(t, zebraIdx < appleIdx, "zebra should come before apple (original order)")
	assert.True(t, appleIdx < mangoIdx, "apple should come before mango (original order)")
}

func TestCollapseYAML_CommentCollision(t *testing.T) {
	input := `redis:
  # Enable redis caching
  enabled: true
postgresql:
  # Enable postgresql database
  enabled: false
`
	opts := CollapseOptions{MaxDepth: 3, ShowDefaults: true, ShowComments: true}

	result, _, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.Contains(t, result, "# Enable redis caching")
	assert.Contains(t, result, "# Enable postgresql database")

	// Verify each comment appears in the correct section
	redisIdx := strings.Index(result, "redis:")
	redisCommentIdx := strings.Index(result, "# Enable redis caching")
	pgIdx := strings.Index(result, "postgresql:")
	pgCommentIdx := strings.Index(result, "# Enable postgresql database")

	assert.True(t, redisCommentIdx > redisIdx, "redis comment should appear after redis key")
	assert.True(t, redisCommentIdx < pgIdx, "redis comment should appear before postgresql key")
	assert.True(t, pgCommentIdx > pgIdx, "postgresql comment should appear after postgresql key")
}

func TestCollapseYAML_InvalidYAML(t *testing.T) {
	input := `invalid: yaml: content: [[[`

	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	_, _, err := CollapseYAML([]byte(input), opts)

	assert.Error(t, err)
}

func TestCollapseYAML_NestedArrayOfObjects(t *testing.T) {
	input := `containers:
  - name: app
    image: nginx
  - name: sidecar
    image: envoy
`
	// MaxDepth=3 shows: root (0) -> containers array (1) -> array items (2) -> item keys (3)
	opts := CollapseOptions{MaxDepth: 3, MaxArrayItems: 3, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "containers:")
	assert.Contains(t, result, "image: nginx")
	assert.Contains(t, result, "name: app")
}

func TestCollapseYAML_NestedArrayOfObjects_Summarized(t *testing.T) {
	input := `containers:
  - name: app
    image: nginx
  - name: sidecar
    image: envoy
`
	// MaxDepth=2 summarizes array items
	opts := CollapseOptions{MaxDepth: 2, MaxArrayItems: 3, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "containers:")
	assert.Contains(t, result, "object (2 keys)")
}

func TestCollapseYAML_ArrayOfObjects_ProducesValidYAML(t *testing.T) {
	input := `items:
  - name: app
    config:
      nested: value
  - name: db
    config:
      nested: value2
`
	opts := CollapseOptions{MaxDepth: 4, MaxArrayItems: 10, ShowDefaults: true, ShowComments: true}

	result, _, err := CollapseYAML([]byte(input), opts)
	require.NoError(t, err)

	// Verify result is valid YAML by parsing it back
	var parsed interface{}
	parseErr := yaml.Unmarshal([]byte(result), &parsed)
	assert.NoError(t, parseErr, "Collapsed output should be valid YAML: %s", result)
}

func TestCollapseYAML_DeeplyNested(t *testing.T) {
	input := `level1:
  level2:
    level3:
      level4:
        level5:
          value: deep
`
	opts := CollapseOptions{MaxDepth: 3, MaxArrayItems: 3, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "level1:")
	assert.Contains(t, result, "level2:")
	assert.Contains(t, result, "level3: object (1 key)")
	assert.NotContains(t, result, "level4:")
}

func TestDefaultCollapseOptions(t *testing.T) {
	opts := DefaultCollapseOptions()

	assert.Equal(t, 2, opts.MaxDepth)
	assert.Equal(t, 3, opts.MaxArrayItems)
	assert.True(t, opts.ShowDefaults)
	assert.False(t, opts.ShowComments)
}

func TestInferType(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil", nil, "null"},
		{"true", true, "boolean"},
		{"false", false, "boolean"},
		{"int", 42, "number"},
		{"int64", int64(42), "number"},
		{"float64", 3.14, "number"},
		{"uint64", uint64(100), "number"},
		{"string", "hello", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, inferType(tt.value))
		})
	}
}

func TestFormatScalar(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil", nil, "null"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"int", 42, "42"},
		{"int64", int64(100), "100"},
		{"float64_integer", float64(5), "5"},
		{"float64_decimal", 3.14, "3.14"},
		{"string", "hello", "hello"},
		{"empty_string", "", `""`},
		{"string_with_colon", "has:colon", `"has:colon"`},
		{"yaml_keyword", "true", `"true"`},
		{"numeric_string", "123", `"123"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatScalar(tt.value))
		})
	}
}

func TestNeedsQuoting(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", false},
		{"", true},
		{"with:colon", true},
		{"with#hash", true},
		{"true", true},
		{"false", true},
		{"null", true},
		{"yes", true},
		{"no", true},
		{"on", true},
		{"off", true},
		{"~", true},
		{"123", true},       // Numeric string
		{" leading", true},  // Leading whitespace
		{"trailing ", true}, // Trailing whitespace
		{"has[bracket", true},
		{"has{brace", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, needsQuoting(tt.input))
		})
	}
}

func TestSummarizeOrderedMap(t *testing.T) {
	tests := []struct {
		name     string
		input    *orderedMap
		expected string
	}{
		{"empty", &orderedMap{}, "object (empty)"},
		{"one key", &orderedMap{entries: []orderedEntry{{key: "a", value: 1}}}, "object (1 key)"},
		{"multiple keys", &orderedMap{entries: []orderedEntry{{key: "a", value: 1}, {key: "b", value: 2}, {key: "c", value: 3}}}, "object (3 keys)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, summarizeOrderedMap(tt.input))
		})
	}
}

func TestSummarizeArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected string
	}{
		{"empty", []interface{}{}, "array (empty)"},
		{"one item", []interface{}{1}, "array (1 item)"},
		{"multiple items", []interface{}{1, 2, 3}, "array (3 items)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, summarizeArray(tt.input))
		})
	}
}

func TestUnquoteKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain key", "name", "name"},
		{"double quoted", `"a.b.c"`, "a.b.c"},
		{"single quoted", "'a.b.c'", "a.b.c"},
		{"empty string", "", ""},
		{"single char", "a", "a"},
		{"mismatched quotes", `"hello'`, `"hello'`},
		{"only double quotes", `""`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, unquoteKey(tt.input))
		})
	}
}

func TestCollapseYAML_QuotedKeysWithDots(t *testing.T) {
	input := `"a.b.c": value1
normal: value2
'x.y': value3
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	// Keys should NOT have literal quote characters
	assert.Contains(t, result, "a.b.c: value1")
	assert.NotContains(t, result, `"a.b.c"`)
	assert.Contains(t, result, "normal: value2")
	assert.Contains(t, result, "x.y: value3")
	assert.NotContains(t, result, `'x.y'`)
}

func TestCollapseYAML_QuotedKeysComments(t *testing.T) {
	input := `# Comment on dotted key
"a.b": nested_value
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, _, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.Contains(t, result, "# Comment on dotted key")
	assert.Contains(t, result, "a.b: nested_value")
}

func TestCollapseYAML_Infinity(t *testing.T) {
	input := `pos: .inf
neg: -.inf
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "pos: .inf")
	assert.Contains(t, result, "neg: -.inf")
	// Must NOT contain Go's float representation
	assert.NotContains(t, result, "+Inf")
	assert.NotContains(t, result, "-Inf")
}

func TestCollapseYAML_NaN(t *testing.T) {
	input := `val: .nan
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	assert.Contains(t, result, "val: .nan")
}

func TestCollapseYAML_AnchorAlias(t *testing.T) {
	input := `defaults: &base
  timeout: 30
  retries: 3
service:
  <<: *base
  name: myapp
`
	opts := CollapseOptions{MaxDepth: 3, ShowDefaults: true, ShowComments: true}

	result, collapsed, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.True(t, collapsed)
	// The alias should be rendered visibly, not resolved to the anchor name string
	assert.Contains(t, result, "*base")
}

func TestCollapseYAML_FilterSchemaAnnotations(t *testing.T) {
	input := `# @schema
# type: object
# @schema
# -- Enable the service
service:
  enabled: true
`
	opts := CollapseOptions{MaxDepth: 3, ShowDefaults: true, ShowComments: true}

	result, _, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.NotContains(t, result, "@schema")
	assert.NotContains(t, result, "type: object")
	assert.Contains(t, result, "# Enable the service")
}

func TestCollapseYAML_MultiLineCommentFirstLineOnly(t *testing.T) {
	input := `# -- Main description line
# Additional detail line 1
# Additional detail line 2
name: test
`
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, _, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.Contains(t, result, "# Main description line")
	assert.NotContains(t, result, "Additional detail")
}

func TestCollapseYAML_SkipCommentsOnCollapsedEntries(t *testing.T) {
	input := `# Service configuration
service:
  type: ClusterIP
  port: 80
`
	// depth=1 means service will be collapsed to "object (2 keys)"
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, _, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	assert.Contains(t, result, "service: object (2 keys)")
	assert.NotContains(t, result, "# Service configuration")
}

func TestCollapseYAML_CommentsOnExpandedEntries(t *testing.T) {
	input := `# Replica count
replicaCount: 3
# Service configuration
service:
  type: ClusterIP
  port: 80
`
	// depth=2 means service is expanded, replicaCount is a scalar
	opts := CollapseOptions{MaxDepth: 2, ShowDefaults: true, ShowComments: true}

	result, _, err := CollapseYAML([]byte(input), opts)

	require.NoError(t, err)
	// Scalar entry — comment should be shown
	assert.Contains(t, result, "# Replica count")
	// Expanded map entry — comment should be shown
	assert.Contains(t, result, "# Service configuration")
}

// TestCollapseYAML_RealisticChartSizeBudget generates a Traefik-scale values.yaml
// (~60KB raw with comments, @schema annotations, deeply nested structures) and
// verifies that the default collapse settings produce output well under the 40KB
// token budget. This is the key validation that the new defaults achieve the goal.
func TestCollapseYAML_RealisticChartSizeBudget(t *testing.T) {
	// Build a realistic Traefik-scale chart: ~60 top-level keys, each with
	// @schema annotations, multi-line comments, and 3-5 nested keys (some 3+ deep).
	var sb strings.Builder
	sections := []struct {
		name     string
		comment  string
		children int
		depth    int // extra nesting levels below children
	}{
		{"image", "Container image configuration", 6, 1},
		{"deployment", "Deployment configuration", 10, 1},
		{"rollingUpdate", "Rolling update strategy", 4, 1},
		{"ingressRoute", "IngressRoute CRD configuration", 8, 2},
		{"providers", "Provider configuration", 6, 2},
		{"ports", "Port configuration", 12, 1},
		{"service", "Kubernetes service configuration", 8, 1},
		{"autoscaling", "HPA configuration", 6, 0},
		{"persistence", "Persistent volume configuration", 5, 1},
		{"certResolvers", "ACME certificate resolvers", 5, 2},
		{"env", "Environment variables", 4, 0},
		{"envFrom", "Environment from sources", 4, 0},
		{"additionalArguments", "Additional CLI arguments", 3, 0},
		{"additionalVolumeMounts", "Additional volume mounts", 3, 1},
		{"resources", "Resource requests and limits", 4, 1},
		{"nodeSelector", "Node selector constraints", 4, 0},
		{"tolerations", "Pod tolerations", 3, 1},
		{"affinity", "Pod affinity rules", 5, 2},
		{"topologySpreadConstraints", "Topology spread constraints", 4, 1},
		{"priorityClassName", "Priority class name", 0, 0},
		{"securityContext", "Pod security context", 6, 0},
		{"podSecurityContext", "Container security context", 5, 0},
		{"rbac", "RBAC configuration", 5, 1},
		{"serviceAccount", "ServiceAccount configuration", 4, 0},
		{"metrics", "Prometheus metrics configuration", 6, 2},
		{"tracing", "Tracing configuration", 5, 1},
		{"globalArguments", "Global CLI arguments", 3, 0},
		{"logs", "Logging configuration", 5, 1},
		{"accessLogs", "Access log configuration", 6, 1},
		{"pilot", "Traefik Pilot configuration", 4, 0},
		{"experimental", "Experimental features", 6, 2},
		{"ingressClass", "IngressClass configuration", 4, 0},
		{"hub", "Traefik Hub integration", 5, 1},
		{"gateway", "Gateway API configuration", 5, 2},
		{"tlsOptions", "TLS options configuration", 4, 1},
		{"tlsStore", "TLS store configuration", 3, 1},
		{"middlewares", "Middleware configuration", 6, 2},
		{"commonLabels", "Labels added to all resources", 0, 0},
		{"commonAnnotations", "Annotations added to all resources", 0, 0},
		{"podLabels", "Labels added to pods", 0, 0},
		{"podAnnotations", "Annotations added to pods", 0, 0},
		{"namespaceOverride", "Override release namespace", 0, 0},
		{"instanceLabelOverride", "Override instance label", 0, 0},
		{"updateStrategy", "Deployment update strategy", 3, 1},
		{"hostNetwork", "Use host networking", 0, 0},
		{"dnsPolicy", "DNS policy for pods", 0, 0},
		{"dnsConfig", "DNS configuration for pods", 3, 1},
	}

	for _, sec := range sections {
		// @schema annotation block (realistic: 4-8 lines of schema props)
		sb.WriteString("# @schema\n")
		sb.WriteString("# type: object\n")
		sb.WriteString("# required: false\n")
		sb.WriteString("# additionalProperties: true\n")
		sb.WriteString("# description: " + sec.comment + "\n")
		sb.WriteString("# @schema\n")
		// Multi-line Helm comment (3-5 lines typical in Traefik)
		fmt.Fprintf(&sb, "# -- %s\n", sec.comment)
		sb.WriteString("# Additional detail about this section that provides\n")
		sb.WriteString("# more context but is not essential for understanding.\n")
		sb.WriteString("# See https://doc.traefik.io/traefik/ for full reference.\n")
		sb.WriteString("# @default -- See values.yaml for the complete default.\n")

		if sec.children == 0 {
			fmt.Fprintf(&sb, "%s: \"\"\n", sec.name)
		} else {
			fmt.Fprintf(&sb, "%s:\n", sec.name)
			for i := 0; i < sec.children; i++ {
				childName := fmt.Sprintf("child%d", i)
				// Each child also gets @schema + comment
				sb.WriteString("  # @schema\n")
				sb.WriteString("  # type: string\n")
				sb.WriteString("  # @schema\n")
				fmt.Fprintf(&sb, "  # -- %s child %d setting\n", sec.name, i)
				sb.WriteString("  # Detailed explanation of what this child controls.\n")
				if sec.depth == 0 {
					fmt.Fprintf(&sb, "  %s: default_value_%d\n", childName, i)
				} else {
					fmt.Fprintf(&sb, "  %s:\n", childName)
					for j := 0; j < 4; j++ {
						sb.WriteString("    # @schema\n")
						sb.WriteString("    # type: string\n")
						sb.WriteString("    # @schema\n")
						fmt.Fprintf(&sb, "    # -- Nested setting %d\n", j)
						if sec.depth > 1 {
							fmt.Fprintf(&sb, "    nested%d:\n", j)
							for k := 0; k < 3; k++ {
								fmt.Fprintf(&sb, "      deep%d: value_%d_%d_%d\n", k, i, j, k)
							}
						} else {
							fmt.Fprintf(&sb, "    nested%d: value_%d_%d\n", j, i, j)
						}
					}
				}
			}
		}
	}

	input := sb.String()
	rawBytes := len(input)
	t.Logf("Raw input size: %d bytes (%d lines)", rawBytes, strings.Count(input, "\n"))
	require.Greater(t, rawBytes, 40000, "Input should be large enough to test budget (>40KB)")

	// Test 1: Default options (depth=2, comments=off)
	defaults := DefaultCollapseOptions()
	result, collapsed, err := CollapseYAML([]byte(input), defaults)
	require.NoError(t, err)
	assert.True(t, collapsed)

	defaultBytes := len(result)
	defaultLines := strings.Count(result, "\n") + 1
	t.Logf("Default (depth=%d, comments=%v): %d bytes, %d lines",
		defaults.MaxDepth, defaults.ShowComments, defaultBytes, defaultLines)
	assert.Less(t, defaultBytes, MaxResponseBytes,
		"Default output (%d bytes) should be well under MaxResponseBytes (%d)", defaultBytes, MaxResponseBytes)

	// Should not contain any @schema content
	assert.NotContains(t, result, "@schema")

	// Test 2: depth=2 with comments ON — should still be under budget but larger
	withComments := CollapseOptions{MaxDepth: 2, MaxArrayItems: 3, ShowDefaults: true, ShowComments: true}
	resultComments, _, err := CollapseYAML([]byte(input), withComments)
	require.NoError(t, err)

	commentBytes := len(resultComments)
	commentLines := strings.Count(resultComments, "\n") + 1
	t.Logf("With comments (depth=2): %d bytes, %d lines", commentBytes, commentLines)

	assert.Greater(t, commentBytes, defaultBytes,
		"Output with comments should be larger than without")
	// Comments should contain only first lines, no @schema
	assert.NotContains(t, resultComments, "@schema")
	assert.NotContains(t, resultComments, "Additional detail")
	assert.NotContains(t, resultComments, "See https://doc.traefik.io")

	// Comments on collapsed entries should be skipped
	// At depth=2, sections with depth>0 have children that contain sub-maps which collapse.
	// The comment for those collapsed sub-entries should not appear.

	// Test 3: Verify explicit show_comments=true still works as override
	assert.Contains(t, resultComments, "# Container image configuration")

	// Test 4: Log the reduction ratio for manual review
	t.Logf("Reduction: raw %d -> default %d (%.0f%% reduction)",
		rawBytes, defaultBytes, 100*(1-float64(defaultBytes)/float64(rawBytes)))
	t.Logf("Comments add: %d bytes (%.0f%% overhead)",
		commentBytes-defaultBytes, 100*float64(commentBytes-defaultBytes)/float64(defaultBytes))
}

// TestCollapseYAML_CollapsedEntryCommentsSkipped verifies that comments
// are skipped specifically on entries that will be collapsed, while
// sibling scalar entries retain their comments.
func TestCollapseYAML_CollapsedEntryCommentsSkipped(t *testing.T) {
	input := `# -- Scalar setting
enabled: true
# -- Object that will be collapsed
nested:
  key1: val1
  key2: val2
# -- Another scalar
name: test
`
	// depth=1: nested will collapse to "object (2 keys)"
	opts := CollapseOptions{MaxDepth: 1, ShowDefaults: true, ShowComments: true}

	result, _, err := CollapseYAML([]byte(input), opts)
	require.NoError(t, err)

	// Scalar entries keep their comments
	assert.Contains(t, result, "# Scalar setting")
	assert.Contains(t, result, "# Another scalar")
	// Collapsed entry's comment is skipped
	assert.NotContains(t, result, "# Object that will be collapsed")
	assert.Contains(t, result, "nested: object (2 keys)")
}

// TestExtractFirstCommentLine validates the comment processing logic directly.
func TestExtractFirstCommentLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "# hello", "hello"},
		{"helm convention", "# -- Enable the feature", "Enable the feature"},
		{"multi-line keeps first", "# -- First line\n# Second line\n# Third line", "First line"},
		{"schema block filtered", "# @schema\n# type: object\n# @schema\n# -- Description", "Description"},
		{"schema only", "# @schema\n# type: string\n# @schema", ""},
		{"empty", "", ""},
		{"blank lines skipped", "#\n#\n# actual content", "actual content"},
		{"schema then content", "# @schema\n# required: true\n# @schema\n# After schema", "After schema"},
		{"nested schema props", "# @schema\n# type: object\n# properties:\n#   foo:\n#     type: string\n# @schema\n# -- The real desc", "The real desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractFirstCommentLine(tt.input))
		})
	}
}

// TestWillCollapse validates the helper function directly.
func TestWillCollapse(t *testing.T) {
	om := &orderedMap{entries: []orderedEntry{{key: "a", value: 1}}}
	emptyOm := &orderedMap{}
	arr := []interface{}{1, 2}
	emptyArr := []interface{}{}

	opts := CollapseOptions{MaxDepth: 2}

	// At depth < MaxDepth: should not collapse
	assert.False(t, willCollapse(om, 1, opts))
	assert.False(t, willCollapse(arr, 1, opts))

	// At depth >= MaxDepth: should collapse non-empty
	assert.True(t, willCollapse(om, 2, opts))
	assert.True(t, willCollapse(arr, 2, opts))
	assert.True(t, willCollapse(om, 3, opts))

	// Empty containers never collapse (they render inline)
	assert.False(t, willCollapse(emptyOm, 2, opts))
	assert.False(t, willCollapse(emptyArr, 2, opts))

	// Scalars never collapse
	assert.False(t, willCollapse(42, 2, opts))
	assert.False(t, willCollapse("str", 2, opts))
	assert.False(t, willCollapse(nil, 2, opts))

	// Unlimited depth: never collapse
	unlimited := CollapseOptions{MaxDepth: 0}
	assert.False(t, willCollapse(om, 100, unlimited))
}

func TestRenderScalar(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		showDefaults bool
		expected     string
	}{
		{"int_with_defaults", 42, true, "42"},
		{"int_without_defaults", 42, false, "number"},
		{"string_with_defaults", "hello", true, "hello"},
		{"string_without_defaults", "hello", false, "string"},
		{"bool_with_defaults", true, true, "true"},
		{"bool_without_defaults", true, false, "boolean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			renderScalar(&sb, tt.value, tt.showDefaults)
			assert.Equal(t, tt.expected, sb.String())
		})
	}
}
