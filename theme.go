package forms

import (
	"reflect"

	"github.com/jaredtmartin/bolt-go"
)

// type ComponentLibrary map[string]Component

type Theme struct {
	base      Component
	fieldName map[string]Component
	fieldType map[string]Component
	dataType  map[string]Component
}

func (t *Theme) GetComponent(meta reflect.StructField) Component {
	// Get by Tag Name
	if componentName := meta.Tag.Get("element"); componentName != "" {
		if el, ok := t.fieldType[componentName]; ok {
			return el
		}
	}
	// Get by Field Name
	if el, ok := t.fieldName[meta.Name]; ok {
		return el
	}
	dataType := meta.Type.String()
	// Get by Data Type
	if el, ok := t.dataType[dataType]; ok {
		return el
	}
	if t.base != nil {
		return t.base
	}
	return missingComponent(meta.Name)
}

func missingComponent(componentName string) Component {
	return func(name, label, value string) *bolt.Field {
		field := &bolt.Field{DefaultElement: bolt.NewDefaultElement("p")}
		field.DefaultElement.Text("Missing Component: " + componentName)
		return field
	}
}
func NewTheme(defaultComponent Component) *Theme {
	return &Theme{
		base:      defaultComponent,
		fieldName: map[string]Component{},
		fieldType: map[string]Component{},
		dataType:  map[string]Component{},
	}
}
func (t *Theme) FieldName(fieldName string, component Component) *Theme {
	t.fieldName[fieldName] = component
	return t
}
func (t *Theme) FieldType(fieldType string, component Component) *Theme {
	t.fieldType[fieldType] = component
	return t
}
func (t *Theme) DataType(dataType string, component Component) *Theme {
	t.dataType[dataType] = component
	return t
}
