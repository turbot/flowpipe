package templates

import (
	"embed"
	"fmt"
)

//go:embed files/*.html
var templatesFs embed.FS

// Get HTML templates
func HTMLTemplate(name string) ([]byte, error) {
	content, err := templatesFs.ReadFile(fmt.Sprintf("files/%s", name))
	if err != nil {
		return nil, err
	}

	return content, nil
}
