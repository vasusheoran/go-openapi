package scan

import (
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"strings"
)

type tagOptions map[string]bool

func (opts tagOptions) Contains(key string) bool {
	_, ok := opts[key]
	return ok
}

// Parse the given comments and extract OpenAPI info
func (p *Parser) extractOpenAPIInfo(cg *ast.CommentGroup) {
	if cg == nil || len(cg.List) == 0 {
		p.logger.Debug("no openapi info comments found for file %s", p.file.Name.Name)
		return
	}

	for i := 0; i < len(cg.List); i++ {
		text := cg.List[i].Text
		if strings.HasPrefix(text, "// openapi:info") {
			fields := strings.Fields(text[11:])
			switch fields[0] {
			case "info":
				switch fields[1] {
				case "title": // join " " and trim \"
					p.spec.Info.Description = strings.Trim(strings.TrimPrefix(cg.List[i].Text, "// openapi:info title "), "\"")
				case "description":
					start := i + 1
					if fields[2] == "start" {
						endIdx := -1
						for i = i + 1; i < len(cg.List); i++ {
							if strings.Contains(cg.List[i].Text, "// openapi:") {
								endIdx = i
								break
							}
						}
						if endIdx == -1 {
							continue
						}
						var sb strings.Builder
						for i = start; i < endIdx; i++ {
							sb.WriteString(strings.TrimSpace(cg.List[i].Text[2:]))
							sb.WriteString("\n")
						}
						p.spec.Info.Description = sb.String()
					} else {
						p.spec.Info.Description = strings.Trim(strings.TrimPrefix(cg.List[i].Text, "// openapi:info description "), "\"")
					}
				case "version":
					p.spec.Info.Version = strings.Trim(fields[2], "\"")
				case "oas":
					p.spec.OpenAPI = strings.Trim(fields[2], "\"")
				case "server":
					p.spec.Servers = openapi3.Servers{}
					serverList := strings.Split(strings.Trim(strings.TrimPrefix(cg.List[i].Text, "// openapi:info server "), "\""), " ")
					for _, url := range serverList {
						s := &openapi3.Server{
							URL:         url,
							Description: "",
							Variables:   nil,
						}
						p.spec.Servers = append(p.spec.Servers, s)
					}
				}
			}
		} else if strings.HasPrefix(text, "// openapi:tag") {
			description, name := extractContentFromSting(text, "[", "]")
			name = strings.TrimSpace(strings.TrimPrefix(name, "// openapi:tag "))
			p.spec.Tags = append(p.spec.Tags, &openapi3.Tag{
				Name:        name,
				Description: description,
			})
		}
	}
}
