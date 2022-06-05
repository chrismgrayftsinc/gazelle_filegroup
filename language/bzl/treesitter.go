package bzl

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

// Match all function calls in the file with name 'load'
// capture the first arguments of that function as 'module'
//go:embed loads.scm
var loadModuleQuery []byte

// getTreeSitterBzlFileLoads reimplement getBzlFileLoads but instead of using
// the go-starlark parser, use tree-sitter generated parser of Python language.
//
// TODO: try to re-use the Parser, Query and QueryCursor as much as possible
// instead of creating a new one each time.
func getTreeSitterBzlFileLoads(path string) ([]string, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadFile(%q) error: %v", path, err)
	}

	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(python.GetLanguage())

	tree := parser.Parse(nil, f)
	defer tree.Close()

	q, err := sitter.NewQuery(loadModuleQuery, python.GetLanguage())
	if err != nil {
		return nil, fmt.Errorf("New query init error: %v", err)
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(q, tree.RootNode())

	loads := []string{}
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, c := range m.Captures {
			if q.CaptureNameForId(c.Index) != "module" {
				continue
			}

			content := strings.Trim(c.Node.Content(f), `"`)
			loads = append(loads, content)
		}
	}

	return loads, nil
}
