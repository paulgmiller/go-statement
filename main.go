package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run countstatements_dir.go <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]

	// Compute statements per directory (direct and recursive)
	var dirDirect = make(map[string]int)
	var dirsSet = make(map[string]bool)
	// Walk the directory tree, collecting direct counts and directories
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirsSet[path] = true
			return nil
		}
		if filepath.Ext(path) == ".go" {
			count, err := countStatements(path)
			if err != nil {
				return err
			}
			d := filepath.Dir(path)
			dirDirect[d] += count
			dirsSet[d] = true
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the directory: %v\n", err)
		os.Exit(1)
	}

	// Gather and sort directories
	dirs := make([]string, 0, len(dirsSet))
	for d := range dirsSet {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	// Compute recursive counts bottom-up
	dirRecursive := make(map[string]int)
	for i := len(dirs) - 1; i >= 0; i-- {
		d := dirs[i]
		dirRecursive[d] = dirDirect[d]
		for j := i + 1; j < len(dirs); j++ {
			other := dirs[j]
			if filepath.Dir(other) == d {
				dirRecursive[d] += dirRecursive[other]
			}
		}
	}

	// Print directory stats
	for _, d := range dirs {
		fmt.Printf("%s: %d statements (direct), %d statements (recursive)\n", d, dirDirect[d], dirRecursive[d])
	}
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
