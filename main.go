package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
)

// ChangeType defines the type of change observed in the JSON structure
type ChangeType string

const (
	Unchanged   ChangeType = "unchanged"
	Added       ChangeType = "added"
	Deleted     ChangeType = "deleted"
	Modified    ChangeType = "modified"
	TypeChanged ChangeType = "type_changed"
)

// DiffNode represents a node in the comparison result tree
type DiffNode struct {
	Type     ChangeType           `json:"type"`
	OldValue interface{}          `json:"old_value,omitempty"`
	NewValue interface{}          `json:"new_value,omitempty"`
	Children map[string]*DiffNode `json:"children,omitempty"`
	Items    []*DiffNode          `json:"items,omitempty"` // For arrays
}

// Prune recursively removes Unchanged nodes from the tree.
// Returns true if the node itself is Unchanged and has no Modified/Added/Deleted descendants.
func (node *DiffNode) Prune() bool {
	if node.Children != nil {
		newChildren := make(map[string]*DiffNode)
		for k, child := range node.Children {
			if !child.Prune() {
				newChildren[k] = child
			}
		}
		node.Children = newChildren
		if len(node.Children) == 0 && node.Type == Unchanged {
			return true
		}
		// If node was Modified but all children became empty/unchanged (not really possible with our logic, but safe)
		if len(node.Children) == 0 && node.Type == Modified {
			node.Type = Unchanged
			return true
		}
	}

	if node.Items != nil {
		hasChange := false
		for _, item := range node.Items {
			if item.Type != Unchanged {
				hasChange = true
				item.Prune()
			}
		}
		if !hasChange && node.Type == Unchanged {
			node.Items = nil
			return true
		}
	}

	return node.Type == Unchanged
}

// Diff compares two JSON-compatible objects recursively
func Diff(oldVal, newVal interface{}) *DiffNode {
	node := &DiffNode{}

	if reflect.TypeOf(oldVal) != reflect.TypeOf(newVal) {
		if oldVal == nil {
			node.Type = Added
			node.NewValue = newVal
		} else if newVal == nil {
			node.Type = Deleted
			node.OldValue = oldVal
		} else {
			node.Type = TypeChanged
			node.OldValue = oldVal
			node.NewValue = newVal
		}
		return node
	}

	switch v1 := oldVal.(type) {
	case map[string]interface{}:
		v2 := newVal.(map[string]interface{})
		node.Children = make(map[string]*DiffNode)
		allKeys := make(map[string]bool)
		for k := range v1 {
			allKeys[k] = true
		}
		for k := range v2 {
			allKeys[k] = true
		}

		hasChanges := false
		for k := range allKeys {
			childDiff := Diff(v1[k], v2[k])
			node.Children[k] = childDiff
			if childDiff.Type != Unchanged {
				hasChanges = true
			}
		}

		if hasChanges {
			node.Type = Modified
		} else {
			node.Type = Unchanged
		}

	case []interface{}:
		v2 := newVal.([]interface{})
		maxLen := len(v1)
		if len(v2) > maxLen {
			maxLen = len(v2)
		}
		node.Items = make([]*DiffNode, maxLen)

		hasChanges := false
		for i := 0; i < maxLen; i++ {
			var item1, item2 interface{}
			if i < len(v1) {
				item1 = v1[i]
			}
			if i < len(v2) {
				item2 = v2[i]
			}
			itemDiff := Diff(item1, item2)
			node.Items[i] = itemDiff
			if itemDiff.Type != Unchanged {
				hasChanges = true
			}
		}

		if hasChanges || len(v1) != len(v2) {
			node.Type = Modified
		} else {
			node.Type = Unchanged
		}

	default:
		node.NewValue = newVal
		if reflect.DeepEqual(oldVal, newVal) {
			node.Type = Unchanged
		} else {
			node.Type = Modified
			node.OldValue = oldVal
		}
	}

	return node
}

func main() {
	var file1, file2 string
	var diffOnly bool
	var htmlPath string
	var outPath string

	flag.StringVar(&file1, "file1", "", "Path to the first JSON file")
	flag.StringVar(&file2, "file2", "", "Path to the second JSON file")
	flag.BoolVar(&diffOnly, "diff-only", false, "Only show differences")
	flag.StringVar(&htmlPath, "html", "", "Output HTML report to this path")
	flag.StringVar(&outPath, "out", "", "Output colored text report to this file path")
	flag.Parse()

	if file1 == "" || file2 == "" {
		fmt.Println("Usage: json-diff -file1 path/to/file1.json -file2 path/to/file2.json [-diff-only] [-html report.html] [-out report.txt]")
		// Providing a default demonstration if no files provided
		fmt.Println("\nRunning demonstration with sample data:")
		runDemo(diffOnly, htmlPath, outPath)
		return
	}

	data1, err := os.ReadFile(file1)
	if err != nil {
		fmt.Printf("Error reading file 1: %v\n", err)
		return
	}
	data2, err := os.ReadFile(file2)
	if err != nil {
		fmt.Printf("Error reading file 2: %v\n", err)
		return
	}

	var obj1, obj2 interface{}
	if err := json.Unmarshal(data1, &obj1); err != nil {
		fmt.Printf("Error parsing JSON 1: %v\n", err)
		return
	}
	if err := json.Unmarshal(data2, &obj2); err != nil {
		fmt.Printf("Error parsing JSON 2: %v\n", err)
		return
	}

	diffResult := Diff(obj1, obj2)

	if htmlPath != "" {
		if err := GenerateHTMLReport(diffResult, obj1, obj2, htmlPath); err != nil {
			fmt.Printf("Error generating HTML report: %v\n", err)
		} else {
			fmt.Printf("HTML report generated at: %s\n", htmlPath)
		}
	} else if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			return
		}
		defer f.Close()
		GenerateTextReport(f, diffResult, "", diffOnly)
		fmt.Printf("Text report saved to: %s\n", outPath)
	} else {
		// Output colored text report to console
		GenerateTextReport(os.Stdout, diffResult, "", diffOnly)
	}
}

func runDemo(diffOnly bool, htmlPath string, outPath string) {
	json1 := `{"a": 1, "b": 2, "children": ["1", "2"]}`
	json2 := `{"b": 2, "a": 3, "children": [], "child": [{"name": "xx", "age": 12}]}`

	var obj1, obj2 interface{}
	json.Unmarshal([]byte(json1), &obj1)
	json.Unmarshal([]byte(json2), &obj2)

	diffResult := Diff(obj1, obj2)

	if htmlPath != "" {
		GenerateHTMLReport(diffResult, obj1, obj2, htmlPath)
		fmt.Printf("HTML report generated at: %s\n", htmlPath)
	} else if outPath != "" {
		f, _ := os.Create(outPath)
		defer f.Close()
		GenerateTextReport(f, diffResult, "", diffOnly)
		fmt.Printf("Text report saved to: %s\n", outPath)
	} else {
		GenerateTextReport(os.Stdout, diffResult, "", diffOnly)
	}
}
