package pipeparser

import "github.com/turbot/go-kit/helpers"

const (
	PipelineExtension      = ".fp"
	SqlExtension           = ".sql"
	MarkdownExtension      = ".md"
	VariablesExtension     = ".fpvars"
	AutoVariablesExtension = ".auto.fpvars"
	JsonExtension          = ".json"
	CsvExtension           = ".csv"
	TextExtension          = ".txt"
)

var YamlExtensions = []string{".yml", ".yaml"}

func IsYamlExtension(ext string) bool {
	return helpers.StringSliceContains(YamlExtensions, ext)
}
