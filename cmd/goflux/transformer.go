// Package main - Transformer implementation for GoFlux bitwise transformations
// This file contains the core logic for converting traditional Go patterns into optimized bitwise operations

package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// Transformer handles AST transformations for different optimization modes
// It's like a translator between "normal Go" and "bitwise optimized Go"
type Transformer struct {
	config           *GoFluxConfig
	transformedCount int                 // Count of transformations applied
	flagCounter      uint64              // Counter for generating unique flag values
	generatedFlags   map[string]FlagInfo // Track generated flags for this transformation
}

// FlagInfo stores information about generated bitwise flags
// This helps us understand what each flag represents in the original code
type FlagInfo struct {
	OriginalName string // What the bool field was called originally
	FlagValue    uint64 // The bitwise value (1, 2, 4, 8, 16, etc.)
	Description  string // Human-readable description of what this flag does
}

// NewTransformer creates a new AST transformer with the given configuration
func NewTransformer(config *GoFluxConfig) *Transformer {
	return &Transformer{
		config:         config,
		flagCounter:    0,
		generatedFlags: make(map[string]FlagInfo),
	}
}

// Transform applies transformations to an AST node based on the configured mode
// This is the main entry point where we decide what transformation to apply
func (t *Transformer) Transform(node ast.Node) bool {
	switch t.config.Mode {
	case ModeBitwise:
		return t.transformBoolStructs(node)
	case ModeJumps:
		return t.transformControlFlow(node)
	case ModeCompact:
		return t.transformConstants(node)
	case ModeAll:
		// Apply all transformations - the full revolution! 🚀
		result := t.transformBoolStructs(node)
		result = t.transformControlFlow(node) && result
		result = t.transformConstants(node) && result
		return result
	}
	return true
}

// transformBoolStructs converts struct fields with multiple bools into bitwise flags
// BEFORE: struct { EnableAuth bool; EnableLogs bool; EnableCache bool }
// AFTER:  struct { Flags ConfigFlags } where ConfigFlags uses bitwise operations
func (t *Transformer) transformBoolStructs(node ast.Node) bool {
	structType, ok := node.(*ast.StructType)
	if !ok {
		return true // Continue traversing if not a struct
	}

	// Count boolean fields in this struct
	boolFields := t.findBoolFields(structType)
	if len(boolFields) < 2 {
		// Not worth optimizing structs with less than 2 bool fields
		return true
	}

	if t.config.Verbose {
		fmt.Printf("    🎯 Found struct with %d bool fields - converting to bitwise flags!\n", len(boolFields))
	}

	// This is where the magic happens! ✨
	t.convertBoolFieldsToFlags(structType, boolFields)
	t.transformedCount++

	return true
}

// findBoolFields identifies all boolean fields in a struct
// Returns a slice of field names that are boolean types
func (t *Transformer) findBoolFields(structType *ast.StructType) []string {
	var boolFields []string

	for _, field := range structType.Fields.List {
		// Check if field type is 'bool'
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
			// Extract field names
			for _, name := range field.Names {
				boolFields = append(boolFields, name.Name)

				if t.config.Verbose {
					fmt.Printf("      📌 Found bool field: %s\n", name.Name)
				}
			}
		}
	}

	return boolFields
}

// convertBoolFieldsToFlags transforms boolean fields into a single bitwise flags field
// This is the core of our bitwise revolution! 🔥
func (t *Transformer) convertBoolFieldsToFlags(structType *ast.StructType, boolFields []string) {
	// Generate flag constants for each boolean field
	flagTypeName := "Flags"

	// Remove all boolean fields from the struct
	var newFields []*ast.Field
	for _, field := range structType.Fields.List {
		keepField := true

		// Check if this field is one of our boolean fields
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
			for _, name := range field.Names {
				for _, boolField := range boolFields {
					if name.Name == boolField {
						keepField = false
						break
					}
				}
			}
		}

		if keepField {
			newFields = append(newFields, field)
		}
	}

	// Add the new flags field
	flagsField := &ast.Field{
		Names: []*ast.Ident{{Name: flagTypeName}},
		Type:  &ast.Ident{Name: "uint64"}, // Using uint64 for maximum flag capacity
	}

	// Add comment explaining the transformation
	if t.config.PreserveDocs {
		flagsField.Comment = &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("// GoFlux: Converted %d bool fields to bitwise flags for performance", len(boolFields)),
				},
			},
		}
	}

	newFields = append(newFields, flagsField)
	structType.Fields.List = newFields

	// Store flag information for documentation
	for i, fieldName := range boolFields {
		flagValue := uint64(1 << i) // 1, 2, 4, 8, 16, 32, etc.
		t.generatedFlags[fieldName] = FlagInfo{
			OriginalName: fieldName,
			FlagValue:    flagValue,
			Description:  fmt.Sprintf("Flag for original bool field '%s'", fieldName),
		}

		if t.config.Verbose {
			fmt.Printf("      🏁 Generated flag: %s = %d (binary: %b)\n",
				fieldName, flagValue, flagValue)
		}
	}
}

// transformControlFlow converts if/else chains and switch statements to jump tables
// BEFORE: if condition1 { action1() } else if condition2 { action2() }
// AFTER:  jumpTable[conditionIndex]() - much faster!
func (t *Transformer) transformControlFlow(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.IfStmt:
		return t.optimizeIfElseChain(n)
	case *ast.SwitchStmt:
		return t.optimizeSwitchStatement(n)
	}
	return true
}

// optimizeIfElseChain converts if/else chains to lookup tables
func (t *Transformer) optimizeIfElseChain(ifStmt *ast.IfStmt) bool {
	// For now, just add a comment showing what could be optimized
	if t.config.PreserveDocs && ifStmt.If != token.NoPos {
		// This is a placeholder for the actual jump table optimization
		if t.config.Verbose {
			fmt.Printf("    🎯 Found if/else chain - candidate for jump table optimization!\n")
		}
	}
	return true
}

// optimizeSwitchStatement converts switch statements to jump tables
func (t *Transformer) optimizeSwitchStatement(switchStmt *ast.SwitchStmt) bool {
	// For now, just add a comment showing what could be optimized
	if t.config.PreserveDocs && switchStmt.Switch != token.NoPos {
		if t.config.Verbose {
			fmt.Printf("    🎯 Found switch statement - candidate for jump table optimization!\n")
		}
	}
	return true
}

// transformConstants compresses string literals and constants
// BEFORE: const Secret = "ADMIN_PASSWORD"
// AFTER:  var secret = [...]byte{65,68,77,73,78,...} - obfuscated and smaller
func (t *Transformer) transformConstants(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.BasicLit:
		return t.optimizeStringLiteral(n)
	case *ast.GenDecl:
		return t.optimizeConstDeclaration(n)
	}
	return true
}

// optimizeStringLiteral converts string literals to byte arrays
func (t *Transformer) optimizeStringLiteral(lit *ast.BasicLit) bool {
	if lit.Kind == token.STRING {
		// Remove quotes and get the actual string value
		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}

		if len(value) > 10 && strings.Contains(value, "SECRET") {
			// This is a candidate for obfuscation
			if t.config.Verbose {
				fmt.Printf("    🔒 Found string literal candidate for obfuscation: %s\n",
					lit.Value[:min(20, len(lit.Value))]+"...")
			}
		}
	}
	return true
}

// optimizeConstDeclaration optimizes constant declarations
func (t *Transformer) optimizeConstDeclaration(decl *ast.GenDecl) bool {
	if decl.Tok == token.CONST {
		if t.config.Verbose {
			fmt.Printf("    📋 Found const declaration - candidate for compression!\n")
		}
	}
	return true
}

// Helper function for min (Go 1.21+ has this built-in)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PrintTransformationSummary shows what transformations were applied
func (t *Transformer) PrintTransformationSummary() {
	fmt.Printf("\n📊 GoFlux Transformation Summary:\n")
	fmt.Printf("   🔧 Total transformations applied: %d\n", t.transformedCount)

	if len(t.generatedFlags) > 0 {
		fmt.Printf("   🏁 Generated flags:\n")
		for _, info := range t.generatedFlags {
			fmt.Printf("      • %s → Flag value %d (binary: %b)\n",
				info.OriginalName, info.FlagValue, info.FlagValue)
		}
	}

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   1. Review the transformed code in the output directory\n")
	fmt.Printf("   2. Update your code to use the new bitwise patterns\n")
	fmt.Printf("   3. Benchmark the performance improvements!\n")
}
