package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run countstatements_dir.go <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]
	var totalStmtCount int

	// Walk the directory tree, looking for *.go files.
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // e.g. permissions error
		}

		// Only parse regular files with a .go extension
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			stmtCount, parseErr := countStatements(path)
			if parseErr != nil {
				// You could return the error or just log it and continue
				return parseErr
			}

			// Print the count for this file and add to total
			fmt.Printf("%s: %d statements\n", path, stmtCount)
			totalStmtCount += stmtCount
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total statements in directory '%s': %d\n", dir, totalStmtCount)
}

// countStatements parses a single .go file and returns the statement count, excluding import declarations.
func countStatements(filename string) (int, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return 0, err
	}

	var stmtCount int
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		switch stmt := n.(type) {
		case *ast.DeclStmt:
			// Skip imports
			if gen, ok := stmt.Decl.(*ast.GenDecl); ok {
				if gen.Tok.String() != "import" {
					stmtCount++
				}
			}
		case *ast.AssignStmt,
			*ast.ExprStmt,
			*ast.ReturnStmt,
			*ast.IncDecStmt,
			*ast.GoStmt,
			*ast.DeferStmt,
			*ast.BranchStmt,
			*ast.RangeStmt,
			*ast.IfStmt,
			*ast.ForStmt,
			*ast.CaseClause,
			*ast.SwitchStmt,
			*ast.TypeSwitchStmt,
			*ast.SelectStmt,
			*ast.CommClause,
			*ast.LabeledStmt,
			*ast.BlockStmt:
			stmtCount++
		}
		return true
	})

	return stmtCount, nil
}
