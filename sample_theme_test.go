package forms_test

import (
	"github.com/jaredtmartin/bolt-go"
	"github.com/jaredtmartin/forms"
)

func HiddenField(name, label, value string) *bolt.Field {
	inputEl := bolt.HiddenInput(name, value)
	field := &bolt.Field{
		DefaultElement: bolt.NewDefaultElement(""),
		Input:          inputEl,
	}
	field.Children(inputEl)
	return field
}
func HiddenIdField(name, label, value string) *bolt.Field {
	return HiddenField("Id", "", value)
}
func TextField(name, label, value string) *bolt.Field {
	return bolt.TextField(name, label, value)
}
func TextareaField(name, label, value string) *bolt.Field {
	return bolt.Textarea(name, label, value)
}
func NumberField(name, label, value string) *bolt.Field {
	field := bolt.TextField(name, label, value)
	field.Input.Type("number")
	return field
}
func EmailField(name, label, value string) *bolt.Field {
	input := bolt.TextField(name, label, value)
	input.Type("email")
	return input
}
func PhoneField(name, label, value string) *bolt.Field {
	input := bolt.TextField(name, label, value)
	input.Type("phone")
	return input
}
func CheckboxField(name, label, value string) *bolt.Field {
	return bolt.Checkbox(name, label, value)
}
func RadioField(name, label, value string) *bolt.Field {
	return bolt.Radio(name, label, value)
}
func GenderField(name, label, value string) *bolt.Field {
	return bolt.Select(name, label, value, []bolt.Option{
		{Label: "Male", Value: "male"},
		{Label: "Female", Value: "female"},
	})
}

var theme = forms.NewTheme(TextField).
	FieldName("Model", HiddenIdField).
	FieldName("Age", NumberField).
	FieldType("Text", TextField).
	FieldType("Hidden", HiddenField).
	FieldType("Textarea", TextareaField).
	FieldType("Number", NumberField).
	FieldType("Email", EmailField).
	FieldType("Phone", PhoneField).
	FieldType("Checkbox", CheckboxField).
	FieldType("Radio", RadioField).
	DataType("Gender", GenderField).
	DataType("int", NumberField).
	DataType("int32", NumberField).
	DataType("int64", NumberField).
	DataType("bool", CheckboxField)
