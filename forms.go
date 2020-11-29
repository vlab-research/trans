package trans

import (
	"fmt"
	"regexp"
)

type FieldChoice struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
	Ref   string `json:"ref,omitempty"`
}

type FieldProperties struct {
	Choices     []FieldChoice `json:"choices,omitempty"`
	Description string        `json:"description,omitempty"`
}

type Field struct {
	ID         string          `json:"id,omitempty"`
	Type       string          `json:"type,omitempty"`
	Title      string          `json:"title,omitempty"`
	Ref        string          `json:"ref,omitempty"`
	Properties FieldProperties `json:"properties,omitempty"`
}

type FormJson struct {
	Title  string   `json:"title"`
	Fields []*Field `json:"fields"`
}

type FieldTranslator struct {
	Translate bool              `json:"translate"`
	Mapping   map[string]string `json:"mapping,omitempty"`
}

type FormTranslator struct {
	Surveyid string                      `json:"surveyid"`
	Fields   map[string]*FieldTranslator `json:"fields"`
}

type Answer struct {
	Response string
	Value    string
}

type FormTranslationError struct {
	Message string
}

func (e *FormTranslationError) Error() string {
	return e.Message
}

func GetValue(field *Field, label string) (string, error) {
	r, err := regexp.Compile(`\n-? ?` + label + `.? ([^\n]+)`)

	if err != nil {
		return "", err
	}

	matches := r.FindAllStringSubmatch(field.Title, -1)
	if len(matches) != 1 {
		return "", &FormTranslationError{fmt.Sprintf("Could not find label %v in field text %v", label, field.Title)}
	}

	res := matches[0]
	if len(res) != 2 {
		return "", &FormTranslationError{fmt.Sprintf("Could not find label %v in field text %v", label, field.Title)}
	}

	return res[1], nil
}

func ExtractAnswers(field *Field) ([]Answer, error) {
	choices := field.Properties.Choices
	N := len(choices)

	if N == 0 {
		return nil, &FormTranslationError{fmt.Sprintf("Multiple Choice question with no answer options! Ref: %v", field.Ref)}
	}

	labels := make([]string, N)
	for i, c := range field.Properties.Choices {
		labels[i] = c.Label
	}

	// NOTE: this is how we test if we should search the
	// text for the answer. It's pretty hokey!
	// add metadata to description?
	shortened := labels[0] == "A"

	ans := make([]Answer, N)
	for i, l := range labels {
		if shortened {
			v, err := GetValue(field, l)
			if err != nil {
				return []Answer{}, err
			}

			ans[i] = Answer{l, v}
		} else {
			ans[i] = Answer{l, l}
		}
	}

	return ans, nil
}

func MakeMCTranslator(src *Field, dst *Field) (map[string]string, error) {
	fields := []*Field{src, dst}
	ans := make([][]Answer, len(fields))

	for i, f := range fields {
		a, err := ExtractAnswers(f)
		if err != nil {

			// TODO: keep old error - multierr
			e := &FormTranslationError{fmt.Sprintf("Could not create translator for field %v to field %v. Had error: %v", src.Ref, dst.Ref, err.Error())}
			return nil, e
		}
		ans[i] = a
	}

	if len(ans[0]) != len(ans[1]) {
		return nil, &FormTranslationError{fmt.Sprintf("Could not create translator for field %v to field %v. They had different length answers!", src.Ref, dst.Ref)}
	}

	m := make(map[string]string)

	for i, sa := range ans[0] {
		da := ans[1][i]
		m[sa.Response] = da.Value
	}

	return m, nil
}

var translatorMakers = map[string]func(*Field, *Field) (map[string]string, error){
	"multiple_choice": MakeMCTranslator,
	"opinion_scale":   MakeMCTranslator,
	"rating":          MakeMCTranslator,
}

func MakeFieldTranslator(field, destField *Field) (*FieldTranslator, error) {
	tm, ok := translatorMakers[field.Type]
	if ok {
		translator, err := tm(field, destField)
		if err != nil {
			return nil, err
		}
		return &FieldTranslator{true, translator}, nil
	}

	// NOTE: unrecognized types not dealt with here.
	return &FieldTranslator{false, nil}, nil
}

func findField(ref string, form *FormJson) (*Field, error) {
	for _, f := range form.Fields {
		if f.Ref == ref {
			return f, nil
		}
	}
	return nil, &FormTranslationError{fmt.Sprintf("Could not find field ref %v in form titled %v", ref, form.Title)}
}

func makeTranslator(form, destForm *FormJson, byRef bool) (*FormTranslator, error) {
	if len(form.Fields) != len(destForm.Fields) {
		return nil, &FormTranslationError{"Forms have different lengths!"}
	}

	formTranslator := &FormTranslator{Fields: map[string]*FieldTranslator{}}

	for i, f := range form.Fields {
		var df *Field
		var err error

		if byRef {
			df, err = findField(f.Ref, destForm)
			if err != nil {
				return nil, err
			}
		} else {
			df = destForm.Fields[i]
		}

		ft, err := MakeFieldTranslator(f, df)
		if err != nil {
			return nil, err
		}
		formTranslator.Fields[f.Ref] = ft
	}

	return formTranslator, nil
}

func MakeTranslatorByShape(form, destForm *FormJson) (*FormTranslator, error) {
	return makeTranslator(form, destForm, false)
}

func MakeTranslatorByRef(form, destForm *FormJson) (*FormTranslator, error) {
	return makeTranslator(form, destForm, true)
}
