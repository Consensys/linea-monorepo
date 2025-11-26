package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println(extractSymbolsFromPackage(".."))
}

// Extract all symbols defined in the client package
func extractSymbolsFromPackage(path string) []string {
	var symbols []string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(filePath, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
		if err != nil {
			fmt.Println("Error parsing file:", err)
			return nil
		}

		// Traverse AST to collect defined symbols
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				symbols = append(symbols, x.Name.Name)
			case *ast.GenDecl:
				for _, spec := range x.Specs {
					if valSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valSpec.Names {
							symbols = append(symbols, name.Name)
						}
					} else if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						symbols = append(symbols, typeSpec.Name.Name)
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		fmt.Println("Error walking package path:", err)
	}
	return symbols
}
