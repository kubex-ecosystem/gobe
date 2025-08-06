// Package main implements GoFlux - The Go AST transformation engine for bitwise optimization
// GoFlux transforms traditional Go code into highly optimized bitwise operations
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Transformation modes - what kind of magic we're applying
	ModeBitwise = "bitwise" // Convert bool structs to bitwise flags
	ModeJumps   = "jumps"   // Convert if/switch chains to jump tables
	ModeCompact = "compact" // Compress strings and constants
	ModeAll     = "all"     // Apply all transformations
)

// GoFluxConfig holds configuration for AST transformations
type GoFluxConfig struct {
	InputDir     string // Where to read original Go files
	OutputDir    string // Where to write transformed Go files
	Mode         string // What transformation to apply
	PreserveDocs bool   // Keep original comments for learning
	Verbose      bool   // Show detailed transformation logs
}

func main() {
	fmt.Println("🚀 GoFlux - The Go Bitwise Revolution Engine")
	fmt.Println("   Transforming traditional Go → Ultra-optimized bitwise Go")
	fmt.Println()

	// Parse command line flags with intuitive names
	config := parseFlags()

	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	// Transform the code!
	if err := transformGoCode(config); err != nil {
		log.Fatalf("❌ Transformation failed: %v", err)
	}

	fmt.Printf("✅ GoFlux transformation completed successfully!\n")
	fmt.Printf("   📂 Output: %s\n", config.OutputDir)
	fmt.Printf("   🎯 Mode: %s\n", config.Mode)
}

// parseFlags parses command line arguments with clear descriptions
func parseFlags() *GoFluxConfig {
	var config GoFluxConfig

	flag.StringVar(&config.InputDir, "in", ".",
		"Input directory containing Go files to transform")

	flag.StringVar(&config.OutputDir, "out", "./goflux_output",
		"Output directory for transformed Go files")

	flag.StringVar(&config.Mode, "mode", ModeBitwise,
		fmt.Sprintf("Transformation mode (%s|%s|%s|%s)",
			ModeBitwise, ModeJumps, ModeCompact, ModeAll))

	flag.BoolVar(&config.PreserveDocs, "preserve-docs", true,
		"Keep original comments to understand transformations")

	flag.BoolVar(&config.Verbose, "verbose", false,
		"Show detailed logs of each transformation")

	flag.Parse()
	return &config
}

// validateConfig ensures configuration makes sense
func validateConfig(config *GoFluxConfig) error {
	validModes := []string{ModeBitwise, ModeJumps, ModeCompact, ModeAll}
	for _, mode := range validModes {
		if config.Mode == mode {
			return nil
		}
	}
	return fmt.Errorf("invalid mode '%s', must be one of: %s",
		config.Mode, strings.Join(validModes, ", "))
}

// transformGoCode performs the actual AST transformation magic
func transformGoCode(config *GoFluxConfig) error {
	fmt.Printf("🔍 Scanning Go files in: %s\n", config.InputDir)

	// Create token file set for tracking source positions
	fileSet := token.NewFileSet()

	// Parse all Go files in input directory
	packages, err := parser.ParseDir(fileSet, config.InputDir, nil,
		parser.ParseComments) // Keep comments for learning
	if err != nil {
		return fmt.Errorf("failed to parse Go files: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Transform each package
	for packageName, pkg := range packages {
		fmt.Printf("📦 Processing package: %s\n", packageName)

		for fileName, file := range pkg.Files {
			if err := transformFile(fileSet, file, fileName, config); err != nil {
				return fmt.Errorf("failed to transform %s: %w", fileName, err)
			}
		}
	}

	return nil
}

// transformFile applies AST transformations to a single Go file
func transformFile(fileSet *token.FileSet, file *ast.File, originalPath string, config *GoFluxConfig) error {
	// Extract just the filename for output
	fileName := filepath.Base(originalPath)

	if config.Verbose {
		fmt.Printf("  📄 Transforming: %s\n", fileName)
	}

	// Create transformer based on mode
	transformer := NewTransformer(config)

	// Apply transformations to the AST
	// This is where the magic happens! 🎩✨
	ast.Inspect(file, func(node ast.Node) bool {
		return transformer.Transform(node)
	})

	// Write transformed file to output directory
	outputPath := filepath.Join(config.OutputDir, fileName)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Format and write the transformed AST
	if err := format.Node(outputFile, fileSet, file); err != nil {
		return fmt.Errorf("failed to format transformed code: %w", err)
	}

	if config.Verbose {
		fmt.Printf("    ✅ Saved: %s\n", outputPath)
	}

	return nil
}
