package main

import (
	"flag"
	"fmt"
	"mac-storage-scout/internal/contractlint"
	"os"
	"strings"
)

func main() {
	root := flag.String("root", ".", "project root to scan")
	includeTests := flag.Bool("include-tests", false, "include *_test.go files")
	exportedOnly := flag.Bool("exported-only", false, "check exported entities only")
	format := flag.String("format", "text", "output format: text|markdown")
	flag.Parse()

	report, err := contractlint.Run(contractlint.Config{
		Root:         *root,
		IncludeTests: *includeTests,
		ExportedOnly: *exportedOnly,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "contract-lint failed: %v\n", err)
		os.Exit(2)
	}

	switch strings.ToLower(*format) {
	case "markdown":
		printMarkdown(report)
	default:
		printText(report)
	}

	if report.ErrorCount > 0 {
		os.Exit(1)
	}
}

func printText(r contractlint.Report) {
	fmt.Printf("scanned entities: %d\n", r.ScannedCount)
	fmt.Printf("violations: %d | errors: %d | warnings: %d\n", r.Violations, r.ErrorCount, r.WarningCount)
	for _, er := range r.Entities {
		if len(er.Findings) == 0 {
			continue
		}
		fmt.Printf("\n- %s %s (%s:%d)\n", er.Entity.Kind, er.Entity.Name, er.Entity.File, er.Entity.Line)
		for _, f := range er.Findings {
			fmt.Printf("  [%s] %s: %s\n", strings.ToUpper(string(f.Severity)), f.Code, f.Message)
		}
	}
}

func printMarkdown(r contractlint.Report) {
	fmt.Printf("# Contract Lint Report\n\n")
	fmt.Printf("- scanned_entities: %d\n", r.ScannedCount)
	fmt.Printf("- violations: %d\n", r.Violations)
	fmt.Printf("- errors: %d\n", r.ErrorCount)
	fmt.Printf("- warnings: %d\n\n", r.WarningCount)

	fmt.Println("| Kind | Entity | File | Line | Status | Findings |")
	fmt.Println("|---|---|---|---:|---|---|")
	for _, er := range r.Entities {
		status := "ok"
		parts := make([]string, 0, len(er.Findings))
		for _, f := range er.Findings {
			parts = append(parts, fmt.Sprintf("%s:%s", f.Severity, f.Message))
		}
		if len(parts) > 0 {
			status = "violation"
		}
		findings := "-"
		if len(parts) > 0 {
			findings = strings.Join(parts, "<br>")
		}
		fmt.Printf("| %s | %s | %s | %d | %s | %s |\n", er.Entity.Kind, er.Entity.Name, er.Entity.File, er.Entity.Line, status, findings)
	}
}
