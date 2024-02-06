package types

import (
	"fmt"
	"github.com/google/martian/log"
	"github.com/logrusorgru/aurora"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
	"reflect"
	"strings"
	"time"
)

type Showable interface {
	GetAsTable() printers.Table
}

func Show(resource Showable, opts sanitize.RenderOptions) (string, error) {
	data := resource.GetAsTable()
	au := aurora.NewAurora(opts.ColorEnabled)
	if len(data.Rows) != 1 {
		return "", fmt.Errorf("expected 1 row, got %d", len(data.Rows))
	}
	row := data.Rows[0]
	if len(data.Columns) != len(row.Cells) {
		return "", fmt.Errorf("expected %d columns, got %d", len(data.Columns), len(data.Rows[0].Cells))
	}

	var b strings.Builder

	/* we print primitive types as follows
	<TitleFormat(Title)>:<padding><value>
	*/
	// calc the padding
	// the padding is such that the value is aligned with the longest title
	var maxTitleLength int
	for _, c := range data.Columns {
		if len(c.Name) > maxTitleLength {
			maxTitleLength = len(c.Name)
		}
	}
	// add 2 for the colon and space
	maxTitleLength += 2

	for idx, c := range data.Columns {

		// title
		padFormat := fmt.Sprintf("%%-%ds", maxTitleLength)
		// todo bold only for top level
		title := fmt.Sprintf(padFormat, au.Blue(fmt.Sprintf("%s:", c.Name)).Bold())

		// value

		// TODO handle map, struct and array

		columnVal := row.Cells[idx]

		val := reflect.ValueOf(columnVal)

		switch val.Kind() {
		case reflect.Slice:
			b.WriteString(fmt.Sprintf("%s\n", title))

			s, err := showSlice(val, opts)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		default:

			switch vt := columnVal.(type) {
			case string, int, float64, bool:
				b.WriteString(fmt.Sprintf("%s%v\n", title, vt))
			case time.Time:
				b.WriteString(fmt.Sprintf("%s%v\n", title, vt.Format(time.RFC3339)))
			}

		}
	}

	return b.String(), nil
}

func showSlice(val reflect.Value, opts sanitize.RenderOptions) (string, error) {
	var b strings.Builder

	for i := 0; i < val.Len(); i++ {
		// Retrieve each element in the slice
		elem := dereferencePointer(val.Index(i).Interface())
		var elemString string
		var err error

		if s, ok := elem.(Showable); ok {
			elemString, err = Show(s, opts)
			if err != nil {
				return "", err
			}

		} else {
			elemVal := reflect.ValueOf(elem)
			switch elemVal.Kind() {
			case reflect.Slice:
				elemString, err = showSlice(val, opts)
				if err != nil {
					return "", err
				}
			default:
				elemString = fmt.Sprintf("%v\n", elem)
			}
		}

		soFar := b.String()
		log.Debugf("soFar: %s", soFar)
		elemString = addBullet(elemString)
		// now add the "- " to the first line and indenr all other lines
		b.WriteString(elemString)

	}
	return b.String(), nil
}

// addBullet takes a multiline string.
// It adds "- " to the start of the first line and indents all other lines to align with the bullet.
func addBullet(s string) string {
	lines := strings.Split(s, "\n")

	// Process first line with bullet
	if len(lines) > 0 {
		lines[0] = "- " + lines[0]
	}

	// Process remaining lines
	indent := strings.Repeat(" ", len("- "))
	for i := 1; i < len(lines); i++ {
		if len(lines[i]) > 0 {
			lines[i] = indent + lines[i]
		}
	}

	return strings.Join(lines, "\n")

}

func renderField(key string, value any, level int, au aurora.Aurora) string {
	if !helpers.IsNil(value) {
		return fmt.Sprintf("%s%s\n", au.Blue(key+":").Bold(), value)
	}
	return ""
}

func AddField(key string, value any, table *printers.Table) {
	if helpers.IsNil(value) {
		return
	}

	value = dereferencePointer(value)

	table.Columns = append(table.Columns, printers.TableColumnDefinition{
		Name: key,
	})
	table.Rows[0].Cells = append(table.Rows[0].Cells, value)

}

func dereferencePointer(value any) any {
	val := reflect.ValueOf(value)

	// Check if the value is a pointer
	if val.Kind() == reflect.Ptr {
		// Dereference the pointer and update the value
		value = val.Elem().Interface()

	}
	return value
}
