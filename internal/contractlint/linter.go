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

type Config struct {
	Root         string
	IncludeTests bool
	ExportedOnly bool
}

type EntityKind string

const (
	EntityFunc            EntityKind = "func"
	EntityMethod          EntityKind = "method"
	EntityType            EntityKind = "type"
	EntityInterfaceMethod EntityKind = "interface-method"
)

type Entity struct {
	Kind        EntityKind
	Name        string
	File        string
	Line        int
	ParamsCount int
	ReturnCount int
	Comment     string
}

type Severity string

const (
	SeverityError Severity = "error"
	SeverityWarn  Severity = "warn"
)

type Finding struct {
	Severity Severity
	Code     string
	Message  string
}

type EntityReport struct {
	Entity        Entity
	RequiredMiss  []string
	OptionalNotes []string
	Findings      []Finding
}

type Report struct {
	Entities     []EntityReport
	ScannedCount int
	Violations   int
	ErrorCount   int
	WarningCount int
}

var tagRegex = regexp.MustCompile(`(?m)@([a-zA-Z][a-zA-Z0-9_-]*)`)

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

func commentText(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	return strings.TrimSpace(cg.Text())
}

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
