package main

import (
	"flag"
	"fmt"
	"mac-storage-scout/internal/contractlint"
	"os"
	"strings"
)

// main runs contract lint CLI workflow.
//
// @purpose Execute contract linter CLI with selected mode and format.
// @consumer Developers and AI agents running repository contract checks.
func main() {
	root := flag.String("root", ".", "project root to scan")
	includeTests := flag.Bool("include-tests", false, "include *_test.go files")
	exportedOnly := flag.Bool("exported-only", false, "check exported entities only")
	format := flag.String("format", "text", "output format: text|markdown")
	mode := flag.String("mode", "short", "report mode: short|detailed")
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
		printMarkdown(report, strings.ToLower(*mode) == "detailed")
	default:
		printText(report, strings.ToLower(*mode) == "detailed")
	}

	if report.ErrorCount > 0 {
		os.Exit(1)
	}
}

// printText renders linter report in plain text.
//
// @purpose Provide console-readable contract lint output.
// @consumer CLI users of mss-contract-lint.
// @param r Linter report data.
// @returns Writes formatted output to stdout.
func printText(r contractlint.Report, detailed bool) {
	fmt.Printf("scanned entities: %d\n", r.ScannedCount)
	fmt.Printf("violations: %d | errors: %d | warnings: %d\n", r.Violations, r.ErrorCount, r.WarningCount)
	for _, er := range r.Entities {
		if !detailed && len(er.Findings) == 0 {
			continue
		}
		fmt.Printf("\n- %s %s (%s:%d)\n", er.Entity.Kind, er.Entity.Name, er.Entity.File, er.Entity.Line)
		if detailed {
			required := "-"
			if len(er.RequiredTags) > 0 {
				required = strings.Join(er.RequiredTags, ", ")
			}
			present := "-"
			if len(er.PresentTags) > 0 {
				present = strings.Join(er.PresentTags, ", ")
			}
			missing := "-"
			if len(er.RequiredMiss) > 0 {
				missing = strings.Join(er.RequiredMiss, ", ")
			}
			status := "ok"
			if len(er.Findings) > 0 {
				status = "violation"
			}
			fmt.Printf("  status: %s\n", status)
			fmt.Printf("  required: %s\n", required)
			fmt.Printf("  present: %s\n", present)
			fmt.Printf("  missing-required: %s\n", missing)
		}
		for _, f := range er.Findings {
			fmt.Printf("  [%s] %s: %s\n", strings.ToUpper(string(f.Severity)), f.Code, f.Message)
		}
	}
}

// printMarkdown renders linter report in markdown format.
//
// @purpose Provide markdown report suitable for evidence artifacts.
// @consumer AI agents and reviewers consuming markdown evidence.
// @param r Linter report data.
// @returns Writes markdown output to stdout.
func printMarkdown(r contractlint.Report, detailed bool) {
	fmt.Printf("# Contract Lint Report\n\n")
	fmt.Printf("- scanned_entities: %d\n", r.ScannedCount)
	fmt.Printf("- violations: %d\n", r.Violations)
	fmt.Printf("- errors: %d\n", r.ErrorCount)
	fmt.Printf("- warnings: %d\n\n", r.WarningCount)

	if detailed {
		fmt.Println("| Kind | Entity | File | Line | Status | Required Tags | Present Tags | Missing Required | Findings |")
		fmt.Println("|---|---|---|---:|---|---|---|---|---|")
	} else {
		fmt.Println("| Kind | Entity | File | Line | Status | Findings |")
		fmt.Println("|---|---|---|---:|---|---|")
	}
	for _, er := range r.Entities {
		if !detailed && len(er.Findings) == 0 {
			continue
		}
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
		if detailed {
			required := "-"
			if len(er.RequiredTags) > 0 {
				required = strings.Join(er.RequiredTags, ", ")
			}
			present := "-"
			if len(er.PresentTags) > 0 {
				present = strings.Join(er.PresentTags, ", ")
			}
			missing := "-"
			if len(er.RequiredMiss) > 0 {
				missing = strings.Join(er.RequiredMiss, ", ")
			}
			fmt.Printf("| %s | %s | %s | %d | %s | %s | %s | %s | %s |\n", er.Entity.Kind, er.Entity.Name, er.Entity.File, er.Entity.Line, status, required, present, missing, findings)
			continue
		}
		fmt.Printf("| %s | %s | %s | %d | %s | %s |\n", er.Entity.Kind, er.Entity.Name, er.Entity.File, er.Entity.Line, status, findings)
	}
}
