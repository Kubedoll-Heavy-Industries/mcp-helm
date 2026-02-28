package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// Default values for collapse options.
const (
	defaultMaxDepth      = 2
	defaultMaxArrayItems = 3
)

// CollapseOptions controls how YAML is collapsed for progressive disclosure.
//
// Depth semantics for structure { a: { b: { c: value } } }:
//
//	MaxDepth=0: Full YAML (unlimited)
//	MaxDepth=1: a: object (1 key)
//	MaxDepth=2: a:\n  b: object (1 key)
//	MaxDepth=3: Full expansion to c: value
type CollapseOptions struct {
	// MaxDepth controls how many nesting levels to expand before summarizing.
	// 0 means unlimited (return full YAML).
	MaxDepth int

	// MaxArrayItems limits how many array items to show before truncating.
	// 0 means unlimited. Default is 3.
	MaxArrayItems int

	// ShowDefaults includes actual values. When false, shows types only.
	ShowDefaults bool

	// ShowComments preserves YAML comments in output.
	ShowComments bool
}

// DefaultCollapseOptions returns the default options for collapsing YAML.
func DefaultCollapseOptions() CollapseOptions {
	return CollapseOptions{
		MaxDepth:      defaultMaxDepth,
		MaxArrayItems: defaultMaxArrayItems,
		ShowDefaults:  true,
		ShowComments:  false,
	}
}

// orderedEntry is a key-value pair that preserves insertion order.
type orderedEntry struct {
	key   string
	value interface{}
}

// orderedMap preserves the insertion order of keys, matching the source YAML.
type orderedMap struct {
	entries []orderedEntry
}

// CollapseYAML transforms YAML content with depth limiting for progressive disclosure.
// When depth is limited, nested structures are summarized (e.g., "object (5 keys)").
// Returns the original YAML unchanged if MaxDepth is 0 (unlimited).
//
// The second return value indicates whether any collapsing occurred.
// Returns an error only if the input is not valid YAML.
func CollapseYAML(data []byte, opts CollapseOptions) (string, bool, error) {
	// Unlimited depth - return as-is (possibly with comment stripping)
	if opts.MaxDepth == 0 {
		if !opts.ShowComments {
			return stripComments(data)
		}
		return string(data), false, nil
	}

	// Parse YAML AST to preserve key order and extract comments
	file, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		return "", false, fmt.Errorf("parsing YAML: %w", err)
	}

	// Extract comments keyed by full dotted path
	var comments map[string]string
	if opts.ShowComments {
		comments = extractComments(file)
	}

	// Build ordered tree from AST (use first document only; Helm values.yaml is single-document)
	var root interface{}
	if len(file.Docs) > 0 {
		root = astToOrdered(file.Docs[0].Body)
	}
	if root == nil {
		return "", true, nil
	}

	// Build collapsed output
	var sb strings.Builder
	sb.Grow(len(data) / 2)

	renderNode(&sb, root, "", "", 0, opts, comments)

	return strings.TrimSuffix(sb.String(), "\n"), true, nil
}

// unquoteKey strips surrounding double or single quotes from a YAML key string.
// The AST's Key.String() returns the quoted form for keys like "a.b.c", which
// would render as literal quote characters in the collapsed output.
func unquoteKey(key string) string {
	if len(key) >= 2 {
		if (key[0] == '"' && key[len(key)-1] == '"') ||
			(key[0] == '\'' && key[len(key)-1] == '\'') {
			return key[1 : len(key)-1]
		}
	}
	return key
}

// astToOrdered converts an AST node into an ordered tree structure.
// Maps become *orderedMap, sequences become []interface{}, scalars become Go values.
func astToOrdered(node ast.Node) interface{} {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.MappingNode:
		om := &orderedMap{entries: make([]orderedEntry, 0, len(n.Values))}
		for _, v := range n.Values {
			if v.Key != nil {
				om.entries = append(om.entries, orderedEntry{
					key:   unquoteKey(v.Key.String()),
					value: astToOrdered(v.Value),
				})
			}
		}
		return om

	case *ast.MappingValueNode:
		// A single mapping value at the document root
		om := &orderedMap{entries: []orderedEntry{{
			key:   unquoteKey(n.Key.String()),
			value: astToOrdered(n.Value),
		}}}
		return om

	case *ast.SequenceNode:
		arr := make([]interface{}, 0, len(n.Values))
		for _, v := range n.Values {
			arr = append(arr, astToOrdered(v))
		}
		return arr

	case *ast.TagNode:
		return astToOrdered(n.Value)

	case *ast.AnchorNode:
		return astToOrdered(n.Value)

	case *ast.AliasNode:
		return n.String()

	case *ast.NullNode:
		return nil

	case *ast.BoolNode:
		return n.Value

	case *ast.IntegerNode:
		return n.Value

	case *ast.FloatNode:
		return n.Value

	case *ast.StringNode:
		return n.Value

	case *ast.InfinityNode:
		return n.String()

	case *ast.NanNode:
		return n.String()

	default:
		return node.String()
	}
}

// renderNode renders any YAML node with depth tracking.
func renderNode(sb *strings.Builder, node interface{}, path string, indent string, depth int, opts CollapseOptions, comments map[string]string) {
	switch v := node.(type) {
	case *orderedMap:
		renderMap(sb, v, path, indent, depth, opts, comments)
	case []interface{}:
		renderArray(sb, v, path, indent, depth, opts, comments)
	default:
		renderScalar(sb, v, opts.ShowDefaults)
	}
}

// renderMap handles ordered map nodes with depth limiting.
func renderMap(sb *strings.Builder, m *orderedMap, path string, indent string, depth int, opts CollapseOptions, comments map[string]string) {
	for _, entry := range m.entries {
		childPath := entry.key
		if path != "" {
			childPath = path + "." + entry.key
		}

		// Add comment if available, enabled, and the entry won't be immediately collapsed
		if opts.ShowComments && !willCollapse(entry.value, depth+1, opts) {
			if comment, ok := comments[childPath]; ok {
				sb.WriteString(indent)
				sb.WriteString("# ")
				sb.WriteString(comment)
				sb.WriteString("\n")
			}
		}

		sb.WriteString(indent)
		sb.WriteString(entry.key)
		sb.WriteString(": ")

		renderValue(sb, entry.value, childPath, indent+"  ", depth+1, opts, comments)
	}
}

// renderArray handles array nodes with depth limiting and item truncation.
func renderArray(sb *strings.Builder, arr []interface{}, path string, indent string, depth int, opts CollapseOptions, comments map[string]string) {
	maxItems := opts.MaxArrayItems
	if maxItems == 0 {
		maxItems = len(arr) // unlimited
	}

	for i, item := range arr {
		if i >= maxItems {
			remaining := len(arr) - maxItems
			sb.WriteString(indent)
			fmt.Fprintf(sb, "... and %d more items\n", remaining)
			break
		}

		sb.WriteString(indent)
		sb.WriteString("- ")

		renderArrayItem(sb, item, path, indent, depth+1, opts, comments)
	}
}

// renderArrayItem handles a single array item, with special formatting for objects.
func renderArrayItem(sb *strings.Builder, item interface{}, path string, indent string, depth int, opts CollapseOptions, comments map[string]string) {
	switch v := item.(type) {
	case *orderedMap:
		if len(v.entries) == 0 {
			sb.WriteString("object (empty)\n")
			return
		}
		if depth >= opts.MaxDepth {
			sb.WriteString(summarizeOrderedMap(v))
			sb.WriteString("\n")
			return
		}
		renderInlineMap(sb, v, path, indent+"  ", depth, opts, comments)

	case []interface{}:
		if len(v) == 0 {
			sb.WriteString("array (empty)\n")
			return
		}
		if depth >= opts.MaxDepth {
			sb.WriteString(summarizeArray(v))
			sb.WriteString("\n")
			return
		}
		sb.WriteString("\n")
		renderArray(sb, v, path, indent+"  ", depth, opts, comments)

	default:
		renderScalar(sb, v, opts.ShowDefaults)
		sb.WriteString("\n")
	}
}

// renderInlineMap renders a map with the first key on the current line (for array items).
func renderInlineMap(sb *strings.Builder, m *orderedMap, path string, indent string, depth int, opts CollapseOptions, comments map[string]string) {
	first := m.entries[0]
	childPath := first.key
	if path != "" {
		childPath = path + "." + first.key
	}
	sb.WriteString(first.key)
	sb.WriteString(": ")
	renderValue(sb, first.value, childPath, indent, depth+1, opts, comments)

	for _, entry := range m.entries[1:] {
		childPath = entry.key
		if path != "" {
			childPath = path + "." + entry.key
		}
		sb.WriteString(indent)
		sb.WriteString(entry.key)
		sb.WriteString(": ")
		renderValue(sb, entry.value, childPath, indent, depth+1, opts, comments)
	}
}

// renderValue renders a single value with depth checking.
func renderValue(sb *strings.Builder, value interface{}, path string, indent string, depth int, opts CollapseOptions, comments map[string]string) {
	switch v := value.(type) {
	case *orderedMap:
		if len(v.entries) == 0 {
			sb.WriteString("object (empty)\n")
			return
		}
		if depth >= opts.MaxDepth {
			sb.WriteString(summarizeOrderedMap(v))
			sb.WriteString("\n")
			return
		}
		sb.WriteString("\n")
		renderMap(sb, v, path, indent, depth, opts, comments)

	case []interface{}:
		if len(v) == 0 {
			sb.WriteString("array (empty)\n")
			return
		}
		if depth >= opts.MaxDepth {
			sb.WriteString(summarizeArray(v))
			sb.WriteString("\n")
			return
		}
		sb.WriteString("\n")
		renderArray(sb, v, path, indent, depth, opts, comments)

	default:
		renderScalar(sb, v, opts.ShowDefaults)
		sb.WriteString("\n")
	}
}

// renderScalar writes a scalar value to the builder.
func renderScalar(sb *strings.Builder, v interface{}, showDefaults bool) {
	if showDefaults {
		sb.WriteString(formatScalar(v))
	} else {
		sb.WriteString(inferType(v))
	}
}

// summarizeOrderedMap returns a type summary for an ordered map.
func summarizeOrderedMap(m *orderedMap) string {
	switch len(m.entries) {
	case 0:
		return "object (empty)"
	case 1:
		return "object (1 key)"
	default:
		return fmt.Sprintf("object (%d keys)", len(m.entries))
	}
}

// summarizeArray returns a summary for an array.
func summarizeArray(arr []interface{}) string {
	switch len(arr) {
	case 0:
		return "array (empty)"
	case 1:
		return "array (1 item)"
	default:
		return fmt.Sprintf("array (%d items)", len(arr))
	}
}

// inferType returns the type name for a scalar value.
func inferType(v interface{}) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return "number"
	case string:
		return "string"
	default:
		return "unknown"
	}
}

// formatScalar formats a scalar value for YAML output.
func formatScalar(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case bool:
		return strconv.FormatBool(val)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'g', -1, 64)
	case string:
		if val == "" {
			return `""`
		}
		if needsQuoting(val) {
			return strconv.Quote(val)
		}
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// needsQuoting returns true if a string needs YAML quoting.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}

	if strings.TrimSpace(s) != s {
		return true
	}

	for _, c := range s {
		switch c {
		case ':', '#', '\n', '"', '\'', '[', ']', '{', '}', '&', '*', '!', '|', '>', '%', '@', '`':
			return true
		}
	}

	lower := strings.ToLower(s)
	switch lower {
	case "true", "false", "null", "yes", "no", "on", "off", "~":
		return true
	}

	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}

	return false
}

// extractComments extracts comments from a parsed YAML file.
// Returns a map of full dotted paths to their associated comments.
func extractComments(file *ast.File) map[string]string {
	comments := make(map[string]string, 8)

	if len(file.Docs) > 0 {
		extractCommentsFromNode(file.Docs[0].Body, "", comments)
	}

	return comments
}

// extractCommentsFromNode recursively extracts comments from AST nodes.
// Comments are keyed by full dotted path to avoid collisions.
func extractCommentsFromNode(node ast.Node, path string, comments map[string]string) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.MappingNode:
		for _, value := range n.Values {
			extractCommentsFromNode(value, path, comments)
		}
	case *ast.MappingValueNode:
		keyNode := n.Key
		if keyNode == nil {
			return
		}

		key := unquoteKey(keyNode.String())
		newPath := key
		if path != "" {
			newPath = path + "." + key
		}

		// Extract comment associated with the key, the mapping value node, or
		// the value node itself. goccy/go-yaml attaches preceding-line comments to
		// the MappingValueNode, inline comments on the key node, and value-trailing
		// comments (e.g., `key: value  # comment`) on the value AST node.
		var commentNode *ast.CommentGroupNode
		if c := keyNode.GetComment(); c != nil {
			commentNode = c
		} else if c := n.GetComment(); c != nil {
			commentNode = c
		} else if n.Value != nil {
			if c := n.Value.GetComment(); c != nil {
				commentNode = c
			}
		}
		if commentNode != nil {
			text := extractFirstCommentLine(commentNode.String())
			if text != "" {
				comments[newPath] = text
			}
		}

		extractCommentsFromNode(n.Value, newPath, comments)

	case *ast.SequenceNode:
		for i, value := range n.Values {
			extractCommentsFromNode(value, fmt.Sprintf("%s[%d]", path, i), comments)
		}

	case *ast.AnchorNode:
		extractCommentsFromNode(n.Value, path, comments)

	case *ast.TagNode:
		extractCommentsFromNode(n.Value, path, comments)
	}
}

// extractFirstCommentLine processes a raw comment string (potentially multi-line)
// and returns only the first meaningful line. Lines starting with @schema
// (Helm schema annotations) are skipped as they render as garbled output for LLMs.
func extractFirstCommentLine(raw string) string {
	inSchema := false
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "@schema") {
			inSchema = !inSchema
			continue
		}
		if inSchema {
			continue
		}
		// Helm convention: "-- description" prefix
		line = strings.TrimPrefix(line, "-- ")
		return line
	}
	return ""
}

// willCollapse returns true if the value would be immediately collapsed
// (summarized) at the given depth with the given options.
func willCollapse(value interface{}, depth int, opts CollapseOptions) bool {
	if opts.MaxDepth == 0 {
		return false
	}
	switch v := value.(type) {
	case *orderedMap:
		return len(v.entries) > 0 && depth >= opts.MaxDepth
	case []interface{}:
		return len(v) > 0 && depth >= opts.MaxDepth
	default:
		return false
	}
}

// stripComments removes comments from YAML while preserving structure.
func stripComments(data []byte) (string, bool, error) {
	var root interface{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return "", false, fmt.Errorf("parsing YAML: %w", err)
	}

	out, err := yaml.Marshal(root)
	if err != nil {
		return "", false, fmt.Errorf("marshaling YAML: %w", err)
	}

	return strings.TrimSuffix(string(out), "\n"), false, nil
}
