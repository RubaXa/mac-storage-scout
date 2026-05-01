package contractlint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Config defines linter execution options.
//
// @purpose Configure scope and filtering behavior for contract lint runs.
// @consumer Run entrypoint.
type Config struct {
	Root         string
	IncludeTests bool
	ExportedOnly bool
}

// EntityKind identifies parsed entity category.
//
// @purpose Distinguish functions, methods, types, and interface methods.
// @consumer Entity indexing and report rendering.
type EntityKind string

const (
	EntityFunc            EntityKind = "func"
	EntityMethod          EntityKind = "method"
	EntityType            EntityKind = "type"
	EntityInterfaceMethod EntityKind = "interface-method"
)

// Entity describes one parsed top-level declaration.
//
// @purpose Carry AST-derived metadata used for contract validation.
// @consumer checkEntity validator and report output formatters.
type Entity struct {
	Kind        EntityKind
	Name        string
	File        string
	Line        int
	ParamsCount int
	ReturnCount int
	Comment     string
}

// Severity defines finding severity level.
//
// @purpose Distinguish blocking errors from advisory warnings.
// @consumer Report rendering and exit-code policy.
type Severity string

const (
	SeverityError Severity = "error"
	SeverityWarn  Severity = "warn"
)

// Finding represents one contract validation issue.
//
// @purpose Store machine-readable lint violation details.
// @consumer Entity report output.
type Finding struct {
	Severity Severity
	Code     string
	Message  string
}

// EntityReport contains validation details for one entity.
//
// @purpose Preserve required/present/missing tags and findings per entity.
// @consumer Final lint report and detailed mode.
type EntityReport struct {
	Entity        Entity
	PresentTags   []string
	RequiredTags  []string
	RequiredMiss  []string
	OptionalNotes []string
	Findings      []Finding
}

// Report is aggregate output of one lint run.
//
// @purpose Summarize lint counters and per-entity findings.
// @consumer CLI output and automation exit handling.
type Report struct {
	Entities     []EntityReport
	ScannedCount int
	Violations   int
	ErrorCount   int
	WarningCount int
}

var tagRegex = regexp.MustCompile(`(?m)@([a-zA-Z][a-zA-Z0-9_-]*)`)

// Run executes contract lint analysis.
//
// @purpose Analyze top-level entities and validate required contract tags.
// @consumer cmd/mss-contract-lint CLI.
// @param cfg Linter configuration.
// @returns Lint report and optional execution error.
func Run(cfg Config) (Report, error) {
	root := strings.TrimSpace(cfg.Root)
	if root == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Report{}, fmt.Errorf("resolve root: %w", err)
	}

	entities, err := collectEntities(absRoot, cfg.IncludeTests)
	if err != nil {
		return Report{}, err
	}

	reports := make([]EntityReport, 0, len(entities))
	errCount := 0
	warnCount := 0
	violations := 0

	for _, e := range entities {
		if cfg.ExportedOnly && !isExportedEntity(e) {
			continue
		}
		er := checkEntity(e)
		if len(er.Findings) > 0 {
			violations++
		}
		for _, f := range er.Findings {
			if f.Severity == SeverityError {
				errCount++
			} else if f.Severity == SeverityWarn {
				warnCount++
			}
		}
		reports = append(reports, er)
	}

	sort.Slice(reports, func(i, j int) bool {
		if reports[i].Entity.File == reports[j].Entity.File {
			if reports[i].Entity.Line == reports[j].Entity.Line {
				return reports[i].Entity.Name < reports[j].Entity.Name
			}
			return reports[i].Entity.Line < reports[j].Entity.Line
		}
		return reports[i].Entity.File < reports[j].Entity.File
	})

	return Report{
		Entities:     reports,
		ScannedCount: len(reports),
		Violations:   violations,
		ErrorCount:   errCount,
		WarningCount: warnCount,
	}, nil
}

// collectEntities parses Go files and indexes top-level entities.
//
// @purpose Build analyzable entity list from AST declarations.
// @consumer Run analysis pipeline.
// @param root Project root directory.
// @returns Indexed entities and optional parse error.
func collectEntities(root string, includeTests bool) ([]Entity, error) {
	fset := token.NewFileSet()
	entities := make([]Entity, 0, 128)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") && path != root {
				return filepath.SkipDir
			}
			if name == "bin" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if !includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileNode, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}

		for _, decl := range fileNode.Decls {
			switch dcl := decl.(type) {
			case *ast.FuncDecl:
				name := dcl.Name.Name
				kind := EntityFunc
				if dcl.Recv != nil {
					kind = EntityMethod
					name = fmt.Sprintf("%s.%s", receiverName(dcl.Recv), dcl.Name.Name)
				}
				entities = append(entities, Entity{
					Kind:        kind,
					Name:        name,
					File:        rel,
					Line:        fset.Position(dcl.Pos()).Line,
					ParamsCount: fieldsCount(dcl.Type.Params),
					ReturnCount: fieldsCount(dcl.Type.Results),
					Comment:     commentText(dcl.Doc),
				})

			case *ast.GenDecl:
				if dcl.Tok != token.TYPE {
					continue
				}
				for _, spec := range dcl.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					comment := commentText(ts.Doc)
					if comment == "" {
						comment = commentText(dcl.Doc)
					}
					entities = append(entities, Entity{
						Kind:    EntityType,
						Name:    ts.Name.Name,
						File:    rel,
						Line:    fset.Position(ts.Pos()).Line,
						Comment: comment,
					})

					it, ok := ts.Type.(*ast.InterfaceType)
					if !ok || it.Methods == nil {
						continue
					}
					for _, m := range it.Methods.List {
						if len(m.Names) == 0 {
							continue
						}
						fn, ok := m.Type.(*ast.FuncType)
						if !ok {
							continue
						}
						for _, n := range m.Names {
							methodComment := commentText(m.Doc)
							if methodComment == "" {
								methodComment = commentText(m.Comment)
							}
							entities = append(entities, Entity{
								Kind:        EntityInterfaceMethod,
								Name:        fmt.Sprintf("%s.%s", ts.Name.Name, n.Name),
								File:        rel,
								Line:        fset.Position(m.Pos()).Line,
								ParamsCount: fieldsCount(fn.Params),
								ReturnCount: fieldsCount(fn.Results),
								Comment:     methodComment,
							})
						}
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entities, nil
}

// checkEntity validates required tags for one entity.
//
// @purpose Apply mandatory and conditional contract checks to entity comments.
// @consumer Run analysis pipeline.
// @param e One parsed entity.
// @returns Entity report with findings.
func checkEntity(e Entity) EntityReport {
	r := EntityReport{Entity: e}
	tags := extractTags(e.Comment)

	required := []string{"purpose", "consumer"}
	if e.ParamsCount > 0 {
		required = append(required, "param")
	}
	if e.ReturnCount > 0 {
		required = append(required, "returns")
	}
	r.RequiredTags = append(r.RequiredTags, required...)

	allTagNames := sortedTagNames(tags)
	r.PresentTags = append(r.PresentTags, allTagNames...)

	for _, need := range required {
		if tags[need] {
			continue
		}
		r.RequiredMiss = append(r.RequiredMiss, need)
		r.Findings = append(r.Findings, Finding{
			Severity: SeverityError,
			Code:     "missing-required-tag",
			Message:  fmt.Sprintf("missing required @%s", need),
		})
	}

	if len(r.RequiredMiss) == 0 {
		return r
	}

	// Extended recheck runs only when required tags are missing.
	optional := []string{"pre", "post", "invariant", "implements", "see"}
	for _, opt := range optional {
		if tags[opt] {
			continue
		}
		r.OptionalNotes = append(r.OptionalNotes, opt)
	}
	if len(r.OptionalNotes) > 0 {
		r.Findings = append(r.Findings, Finding{
			Severity: SeverityWarn,
			Code:     "recheck-full-contract",
			Message:  fmt.Sprintf("required tags missing; recheck optional tags: %s", strings.Join(r.OptionalNotes, ", ")),
		})
	}

	return r
}

// fieldsCount counts function parameter/result fields.
//
// @purpose Determine whether conditional tags @param/@returns are required.
// @consumer Entity parser and validator.
// @param fl Field list to count.
// @returns Number of fields.
func fieldsCount(fl *ast.FieldList) int {
	if fl == nil {
		return 0
	}
	count := 0
	for _, f := range fl.List {
		if len(f.Names) == 0 {
			count++
			continue
		}
		count += len(f.Names)
	}
	return count
}

// receiverName extracts receiver type label.
//
// @purpose Build method entity naming in Receiver.Method form.
// @consumer collectEntities parser.
// @param fl Receiver field list.
// @returns Receiver type name or fallback.
func receiverName(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return "unknown"
	}
	f := fl.List[0]
	switch t := f.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	}
	return "unknown"
}

// commentText normalizes doc comment group text.
//
// @purpose Provide stable comment input for tag extraction.
// @consumer collectEntities parser.
// @param cg AST comment group.
// @returns Trimmed comment text.
func commentText(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	return strings.TrimSpace(cg.Text())
}

// extractTags parses @tag names from comments.
//
// @purpose Detect contract tags for required-field checks.
// @consumer checkEntity validator.
// @param comment Raw comment text.
// @returns Set of discovered tags.
func extractTags(comment string) map[string]bool {
	out := map[string]bool{}
	if strings.TrimSpace(comment) == "" {
		return out
	}
	matches := tagRegex.FindAllStringSubmatch(strings.ToLower(comment), -1)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		out[m[1]] = true
	}
	return out
}

// sortedTagNames returns deterministic tag listing.
//
// @purpose Produce stable reporting for present tags.
// @consumer checkEntity and report formatters.
// @param tags Tag set.
// @returns Sorted tag names.
func sortedTagNames(tags map[string]bool) []string {
	if len(tags) == 0 {
		return nil
	}
	out := make([]string, 0, len(tags))
	for k := range tags {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// isExportedEntity checks exported visibility by naming convention.
//
// @purpose Support exported-only linter mode.
// @consumer Run filtering stage.
// @param e Parsed entity.
// @returns True when entity is exported.
func isExportedEntity(e Entity) bool {
	name := e.Name
	if dot := strings.LastIndex(name, "."); dot >= 0 {
		name = name[dot+1:]
	}
	if name == "" {
		return false
	}
	r := rune(name[0])
	return r >= 'A' && r <= 'Z'
}
