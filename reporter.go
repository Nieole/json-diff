package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>JSON Diff 可视化报告</title>
    <style>
        :root {
            --bg-color: #0d1117;
            --container-bg: #161b22;
            --text-color: #c9d1d9;
            --added-bg: rgba(46, 160, 67, 0.15);
            --added-border: #2ea043;
            --deleted-bg: rgba(248, 81, 73, 0.15);
            --deleted-border: #f85149;
            --modified-bg: rgba(187, 128, 9, 0.15);
            --modified-border: #bb8009;
            --unchanged-color: #8b949e;
            --accent-color: #58a6ff;
        }

        body {
            font-family: 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-color);
            margin: 0;
            padding: 20px;
            font-size: 14px;
        }

        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 1px solid #30363d;
        }

        h1 { font-size: 1.5rem; margin: 0; }

        .toolbar {
            display: flex;
            gap: 15px;
            align-items: center;
        }

        .switch {
            position: relative;
            display: inline-block;
            width: 40px;
            height: 20px;
        }

        .switch input { opacity: 0; width: 0; height: 0; }

        .slider {
            position: absolute;
            cursor: pointer;
            top: 0; left: 0; right: 0; bottom: 0;
            background-color: #ccc;
            transition: .4s;
            border-radius: 20px;
        }

        .slider:before {
            position: absolute;
            content: "";
            height: 16px; width: 16px;
            left: 2px; bottom: 2px;
            background-color: white;
            transition: .4s;
            border-radius: 50%;
        }

        input:checked + .slider { background-color: var(--accent-color); }
        input:checked + .slider:before { transform: translateX(20px); }

        .container {
            display: flex;
            justify-content: center;
            height: calc(100vh - 100px);
        }

        .diff-view {
            width: 80%;
            max-width: 1200px;
            background: var(--container-bg);
            border-radius: 8px;
            border: 1px solid #30363d;
            overflow: auto;
            padding: 20px;
        }

        pre {
            margin: 0;
            white-space: pre-wrap;
            word-wrap: break-word;
            line-height: 1.5;
        }

        .node {
            margin-left: 20px;
            border-left: 1px solid #30363d;
            padding-left: 5px;
        }

        .node-line {
            display: flex;
            padding: 2px 4px;
            border-radius: 4px;
            margin: 1px 0;
        }

        .key { color: #79c0ff; font-weight: bold; }
        .string { color: #a5d6ff; }
        .number { color: #ffab70; }
        .boolean { color: #79c0ff; }
        .null { color: #ff7b72; }

        .type-added { background-color: var(--added-bg); border: 1px solid var(--added-border); }
        .type-deleted { background-color: var(--deleted-bg); border: 1px solid var(--deleted-border); }
        .type-modified { background-color: var(--modified-bg); border: 1px solid var(--modified-border); }
        
        .val-old { text-decoration: line-through; opacity: 0.6; margin-right: 10px; }
        .val-new { font-weight: bold; }

        .hidden { display: none !important; }

        .toggle-btn {
            cursor: pointer;
            width: 12px;
            display: inline-block;
            user-select: none;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>JSON Diff Report</h1>
        <div class="toolbar">
            <label>仅显示差异</label>
            <label class="switch">
                <input type="checkbox" id="diffOnlyToggle">
                <span class="slider"></span>
            </label>
        </div>
    </div>

    <div class="container">
        <div class="diff-view" id="diffView">
            <h3>Diff Result</h3>
            <div id="diffContent"></div>
        </div>
    </div>

    <script>
        const diffData = {{.DiffDataJSON}};

        function renderNode(key, node, isLast) {
            const container = document.createElement('div');
            container.className = 'node';
            if (node.type === 'unchanged') container.classList.add('unchanged-node');

            const line = document.createElement('div');
            line.className = 'node-line type-' + node.type;
            
            let content = '';
            if (key) content += '<span class="key">"' + key + '"</span>: ';

            if (node.children || node.items) {
                const isArray = !!node.items;
                const openChar = isArray ? '[' : '{';
                const closeChar = isArray ? ']' : '}';
                
                content += '<span>' + openChar + '</span>';
                line.innerHTML = content;
                container.appendChild(line);

                const list = isArray ? node.items : node.children;
                const keys = isArray ? Object.keys(list) : Object.keys(list).sort();
                
                keys.forEach((k, idx) => {
                   const child = list[k];
                   container.appendChild(renderNode(isArray ? null : k, child, idx === keys.length - 1));
                });

                const endLine = document.createElement('div');
                endLine.className = 'node-line';
                endLine.innerHTML = '<span>' + closeChar + (isLast ? '' : ',') + '</span>';
                container.appendChild(endLine);
            } else {
                if (node.type === 'added') {
                    content += formatValue(node.new_value);
                } else if (node.type === 'deleted') {
                    content += '<span class="val-old">' + formatValue(node.old_value) + '</span>';
                } else if (node.type === 'modified' || node.type === 'type_changed') {
                    content += '<span class="val-old">' + formatValue(node.old_value) + '</span>';
                    content += '<span class="val-new">' + formatValue(node.new_value) + '</span>';
                } else {
                    content += formatValue(node.new_value || node.old_value);
                }
                content += (isLast ? '' : ',');
                line.innerHTML = content;
                container.appendChild(line);
            }

            return container;
        }

        function formatValue(v) {
            if (v === null) return '<span class="null">null</span>';
            if (typeof v === 'string') return '<span class="string">"' + v + '"</span>';
            if (typeof v === 'number') return '<span class="number">' + v + '</span>';
            if (typeof v === 'boolean') return '<span class="boolean">' + v + '</span>';
            if (typeof v === 'object') return JSON.stringify(v);
            return String(v);
        }

        function updateDiffView() {
            const diffOnly = document.getElementById('diffOnlyToggle').checked;
            const content = document.getElementById('diffContent');
            content.innerHTML = '';
            
            // Note: We use the full diffData, but JS handles the visual filtering
            const root = renderNode(null, diffData, true);
            content.appendChild(root);

            if (diffOnly) {
                const unchanged = document.querySelectorAll('.unchanged-node');
                unchanged.forEach(el => {
                    // Only hide if it has no modified children
                    if (!el.querySelector('.type-added, .type-deleted, .type-modified, .type-type_changed')) {
                        el.classList.add('hidden');
                    }
                });
            }
        }

        document.getElementById('diffOnlyToggle').addEventListener('change', updateDiffView);
        updateDiffView();
    </script>
</body>
</html>
`

type HTMLData struct {
	DiffDataJSON template.JS
	OldJSON      template.JS
	NewJSON      template.JS
}

func GenerateHTMLReport(diffResult *DiffNode, oldObj, newObj interface{}, outputPath string) error {
	diffJSON, _ := json.Marshal(diffResult)
	oldJSON, _ := json.Marshal(oldObj)
	newJSON, _ := json.Marshal(newObj)

	data := HTMLData{
		DiffDataJSON: template.JS(diffJSON),
		OldJSON:      template.JS(oldJSON),
		NewJSON:      template.JS(newJSON),
	}

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// GenerateTextReport prints a colored text representation of the diff to the provided writer.
func GenerateTextReport(w io.Writer, node *DiffNode, indent string, diffOnly bool) {
	renderTextNode(w, "", node, indent, true, diffOnly)
}

func renderTextNode(w io.Writer, key string, node *DiffNode, indent string, isLast bool, diffOnly bool) {
	if diffOnly && node.Type == Unchanged && !hasChangedDescendant(node) {
		return
	}

	prefix := indent
	if key != "" {
		prefix += colorCyan + "\"" + key + "\"" + colorReset + ": "
	}

	suffix := ","
	if isLast {
		suffix = ""
	}

	switch {
	case node.Children != nil:
		fmt.Fprintf(w, "%s{\n", prefix)
		keys := make([]string, 0, len(node.Children))
		for k := range node.Children {
			if !diffOnly || node.Children[k].Type != Unchanged || hasChangedDescendant(node.Children[k]) {
				keys = append(keys, k)
			}
		}
		// Sort keys for deterministic output
		sortStrings(keys)
		for i, k := range keys {
			renderTextNode(w, k, node.Children[k], indent+"  ", i == len(keys)-1, diffOnly)
		}
		fmt.Fprintf(w, "%s}%s\n", indent, suffix)

	case node.Items != nil:
		fmt.Fprintf(w, "%s[\n", prefix)
		changedIndices := make([]int, 0)
		for i, item := range node.Items {
			if !diffOnly || item.Type != Unchanged || hasChangedDescendant(item) {
				changedIndices = append(changedIndices, i)
			}
		}
		for i, idx := range changedIndices {
			renderTextNode(w, "", node.Items[idx], indent+"  ", i == len(changedIndices)-1, diffOnly)
		}
		fmt.Fprintf(w, "%s]%s\n", indent, suffix)

	default:
		switch node.Type {
		case Added:
			fmt.Fprintf(w, "%s%s+ %s%s%s\n", prefix, colorGreen, formatValueText(node.NewValue), colorReset, suffix)
		case Deleted:
			fmt.Fprintf(w, "%s%s- %s%s%s\n", prefix, colorRed, formatValueText(node.OldValue), colorReset, suffix)
		case Modified, TypeChanged:
			fmt.Fprintf(w, "%s%s%s%s -> %s%s%s%s\n", prefix, colorRed, formatValueText(node.OldValue), colorReset, colorGreen, formatValueText(node.NewValue), colorReset, suffix)
		default:
			fmt.Fprintf(w, "%s%s%s\n", prefix, formatValueText(node.NewValue), suffix)
		}
	}
}

func hasChangedDescendant(node *DiffNode) bool {
	if node.Type != Unchanged {
		return true
	}
	if node.Children != nil {
		for _, child := range node.Children {
			if hasChangedDescendant(child) {
				return true
			}
		}
	}
	if node.Items != nil {
		for _, item := range node.Items {
			if hasChangedDescendant(item) {
				return true
			}
		}
	}
	return false
}

func formatValueText(v interface{}) string {
	if v == nil {
		return "null"
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}
