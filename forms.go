package forms

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jaredtmartin/bolt-go"
	"github.com/jaredtmartin/hound"
)

type RenderMethod string

const (
	RenderByName      RenderMethod = "RenderByName"
	RenderByFieldType RenderMethod = "RenderByFieldType"
	RenderByDataType  RenderMethod = "RenderByDataType"
)

func (m RenderMethod) String() string {
	return string(m)
}

type Model struct {
	hound.Model `bson:",inline"`
	Theme       *Theme                  `json:"-" bson:"-"`
	fieldConfig map[string]*FieldConfig `json:"-" bson:"-"`
	formatters  map[string]Formatter    `json:"-" bson:"-"`
}
type FieldConfig struct {
	name      string
	label     string
	component Component
	formatter Formatter
}

// type FieldBuilder struct {
// 	config *FieldConfig
// }

func NewModel(collectionName string, id ...string) Model {
	return Model{
		Model:       hound.NewModel(collectionName, id...),
		fieldConfig: map[string]*FieldConfig{
			// "Model": {component: HiddenIdField},
		},
		formatters: map[string]Formatter{
			"int":   IntFormatter,
			"int32": IntFormatter,
			"int64": IntFormatter,
			"bool":  BoolFormatter,
		},
	}
}

// Three ways to override fields:
//  1. Have a custom render method on the struct for that field
//  2. Provide a tag to specify what type of field you want.
//  3. Have a custom render method on the value type

type Component func(name, label, value string) *bolt.Field
type Formatter func(value reflect.Value) string

//	func (m *Model) UseComponent(fieldType string, component Component) {
//		m.Theme[fieldType] = component
//	}
func (m *Model) FieldConfig(name string) *FieldConfig {
	config := &FieldConfig{}
	m.fieldConfig[name] = config
	return config
}
func (c *FieldConfig) Name(name string) {
	c.name = name
}
func (c *FieldConfig) Label(label string) {
	c.label = label
}
func (c *FieldConfig) Component(component Component) {
	c.component = component
}
func (c *FieldConfig) Formatter(formatter Formatter) {
	c.formatter = formatter
}
func getReflectTypeAndValue(obj any) (reflect.Type, reflect.Value) {
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)
	// Dereference if pointer
	if objType.Kind() == reflect.Pointer {
		objType = objType.Elem()
		objValue = objValue.Elem()
	}
	return objType, objValue
}
func (m *Model) renderField(meta reflect.StructField, value reflect.Value, prefix string, themes ...*Theme) *bolt.Field {
	theme := m.getTheme(themes...)
	var component Component
	if config, ok := m.fieldConfig[meta.Name]; ok && config.component != nil {
		component = config.component
	} else {
		component = theme.GetComponent(meta)
	}
	field := component(prefix+m.getName(meta), m.getLabel(meta), m.getValue(meta, value))
	log.Printf(`rendered field for %s: %s`, meta.Name, field.Render())
	return field
}
func errorField(err error) *bolt.Field {
	field := &bolt.Field{DefaultElement: bolt.NewDefaultElement("p")}
	field.DefaultElement.Text(err.Error()).Class("text-red-500")
	return field
}
func (m *Model) getTheme(themes ...*Theme) *Theme {
	if len(themes) > 0 {
		return themes[0]
	}
	if m.Theme != nil {
		return m.Theme
	} else {
		log.Println("No theme provided.")
		return NewTheme(defaultComponent)
	}
}
func (m *Model) Field(name string, obj any, theme ...*Theme) *bolt.Field {
	objType, objValue := getReflectTypeAndValue(obj)
	// Get the struct field by name (on the TYPE, not the value)
	meta, ok := objType.FieldByName(name)
	if !ok {
		log.Printf("Field with name %s not found\n", name)
		return errorField(fmt.Errorf("Field with name %s not found\n", name))
	}
	// Get the value
	value := objValue.FieldByName(name)
	if !value.IsValid() {
		return errorField(fmt.Errorf("Value for with name %s is not valid\n", name))
	}
	return m.renderField(meta, value, "", theme...)
}
func (m *Model) getFormatter(meta reflect.StructField) Formatter {
	if config, ok := m.fieldConfig[meta.Name]; ok && config.formatter != nil {
		return config.formatter
	}
	format := meta.Tag.Get("format")
	if format != "" {
		if formatter, ok := m.formatters[format]; ok {
			return formatter
		}
	}
	if meta.Type.Kind() == reflect.Int || meta.Type.Kind() == reflect.Int32 || meta.Type.Kind() == reflect.Int64 {
		return IntFormatter
	}
	if meta.Type.Kind() == reflect.Bool {
		return BoolFormatter
	}
	if meta.Type.Kind() == reflect.String && meta.Type.String() == "[]string" {
		return StringSliceFormatter
	}
	return StringFormatter
}
func (m *Model) getValue(meta reflect.StructField, value reflect.Value) string {
	formatter := m.getFormatter(meta)
	log.Println(`formatter: `, formatter)
	return formatter(value)
}
func (m *Model) getLabel(meta reflect.StructField) string {
	config, ok := m.fieldConfig[meta.Name]
	if ok {
		return config.label
	}
	label := meta.Tag.Get("label")
	if label != "" {
		return label
	}
	return meta.Name
}

func (m *Model) getName(meta reflect.StructField) string {
	log.Println(`getting name for:`, meta.Name)
	config, ok := m.fieldConfig[meta.Name]
	if ok {
		log.Printf(`found config.name for  %s: %s`, meta.Name, config.name)
		return config.name
	}
	name := meta.Tag.Get("name")
	if name != "" {
		log.Printf(`found name Tag for  %s: %s`, meta.Name, name)
		return name
	}
	log.Printf(`using field name for  %s`, meta.Name)
	return meta.Name
}
func (m *Model) Form(s any, prefix string, theme ...*Theme) bolt.Element {
	objType, objValue := getReflectTypeAndValue(s)
	log.Printf("This object has %d fields to render. ", objType.NumField())
	form := bolt.Form()
	for i := 0; i < objType.NumField(); i++ {

		meta := objType.Field(i)
		value := objValue.Field(i)

		// Skip unexported fields (reflection can't access them)
		if !meta.IsExported() {
			log.Printf(`Skipping %s of type %s becuase it is not exported and reflection can't access it`, meta.Name, meta.Type)
			continue
		}
		log.Printf(`Rendering %s of type %s`, meta.Name, meta.Type)
		form.Add(m.renderField(meta, value, prefix, theme...))
	}
	return form
}

func StringFormatter(v reflect.Value) string {
	return fmt.Sprint(v.String())
}
func IntFormatter(v reflect.Value) string {
	return fmt.Sprint(v.Int())
}
func BoolFormatter(v reflect.Value) string {
	if v.Bool() {
		return "true"
	}
	return ""
}
func IdFormatter(v reflect.Value) string {
	if modelStruct, ok := v.Interface().(Model); ok {
		// fmt.Println("ID via assertion:", modelStruct.Id)
		return modelStruct.Id
	}
	log.Println("Unable to get Id from Model")
	return ""
}
func StringSliceFormatter(v reflect.Value) string {
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.String {
		return strings.Join(v.Interface().([]string), ",")
	}
	return fmt.Sprint(v.Interface())
}
func defaultComponent(name, label, value string) *bolt.Field {
	return bolt.TextField(name, label, value)
}
func (m *Model) Element(name ...string) bolt.Element {
	nameFormat := "Id"
	if len(name) > 0 {
		nameFormat = fmt.Sprintf(name[0], nameFormat)
	}
	return bolt.HiddenInput(nameFormat, m.Id).Attr("special", "ed")
}
