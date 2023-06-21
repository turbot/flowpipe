package pipeline_hcl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"sigs.k8s.io/yaml"
)

// PipelineParser interface defines the Parse method.
type PipelineParser interface {
	Parse(filePath string) (*hcl.File, hcl.Diagnostics)
}

// PipelineYAMLParser struct implements the PipelineParser interface for YAML parsing.
type PipelineYAMLParser struct{}

func (p *PipelineYAMLParser) Parse(filePath string) (*hcl.File, hcl.Diagnostics) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to open file",
				Detail:   fmt.Sprintf("The file %q could not be opened.", filePath),
			},
		}
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The file %q was opened, but an error occured while reading it.", filePath),
			},
		}
	}
	jsonData, err := yaml.YAMLToJSON(src)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read convert YAML to JSON",
				Detail:   fmt.Sprintf("The file %q was opened, but an error occured while converting it to JSON.", filePath),
			},
		}
	}
	return json.Parse(jsonData, filePath)
}

// PipelineHCLParser struct implements the PipelineParser interface for HCL parsing.
type PipelineHCLParser struct {
	files map[string]*hcl.File
}

func (p *PipelineHCLParser) Parse(filePath string) (*hcl.File, hcl.Diagnostics) {
	if existing := p.files[filePath]; existing != nil {
		return existing, nil
	}

	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The configuration file %q could not be read.", filePath),
			},
		}
	}

	return p.parseHCL(src, filePath)
}

func (p *PipelineHCLParser) parseHCL(src []byte, filename string) (*hcl.File, hcl.Diagnostics) {
	if existing := p.files[filename]; existing != nil {
		return existing, nil
	}

	file, diags := hclsyntax.ParseConfig(src, filename, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	p.files[filename] = file
	return file, diags
}

// PipelineParserFactory struct creates the appropriate parser based on the file extension.
type PipelineParserFactory struct{}

// CreateParser creates a parser based on the file extension.
func (f *PipelineParserFactory) CreateParser(filePath string) (PipelineParser, error) {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".yaml", ".yml":
		return &PipelineYAMLParser{}, nil
	case ".hcl":
		return &PipelineHCLParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}
