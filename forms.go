package trans

import (
	"encoding/json"
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

type Workspace struct {
	Href string `json:"href,omitempty"`
}

type Form struct {
	Workspace       Workspace       `json:"workspace,omitempty"`
	Title           string          `json:"title"`
	Fields          []*Field        `json:"fields"`
	ThankYouScreens []*Field        `json:"thankyou_screens,omitempty"`
	Logic           json.RawMessage `json:"logic,omitempty"`
}

type FieldTranslator struct {
	Translate bool              `json:"translate"`
	Mapping   map[string]string `json:"mapping,omitempty"`
}

type FormTranslator struct {
	Fields map[string]*FieldTranslator `json:"fields"`
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

func ExtractLabels(options string) ([]*Answer, error) {
	character := `[A-Z]` // [\p{L}] for unicode? Only caps?
	base := `(?:^|\n)(?:- ?(%s)(?:[^\S\r\n]|[\p{Pd}-\.\)])+|(%s)[\p{Pd}-\.\)]+[^\S\r\n]?)([^\n]+)`
	r, _ := regexp.Compile(fmt.Sprintf(base, character, character))
	matches := r.FindAllStringSubmatch(options, -1)

	answers := []*Answer{}

	for _, match := range matches {
		if len(match) != 4 {
			return answers, fmt.Errorf("Could not make labels from options: %s", options)
		}
		if match[1] == "" && match[2] == "" {
			return answers, fmt.Errorf("Could not make labels from options: %s", options)
		}
		if match[1] != "" {
			answers = append(answers, &Answer{match[1], match[3]})
		}
		if match[2] != "" {
			answers = append(answers, &Answer{match[2], match[3]})
		}

	}

	return answers, nil
}

func mapResponse(ans []*Answer) []string {
	res := make([]string, len(ans))
	for i, a := range ans {
		res[i] = a.Response
	}
	return res
}

func compare(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, bb := range b {
		if a[i] != bb {
			return false
		}
	}

	return true
}

func ExtractAnswers(field *Field) ([]*Answer, error) {
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

	// another option - try to extractlabels and if it fails with an error
	// make it none? This is elegant but doesn't help formatting issues.
	shortened := labels[0] == "A"
	answers := make([]*Answer, N)

	if shortened {
		a, err := ExtractLabels(field.Title)
		if err != nil {
			return answers, err
		}

		if !compare(labels, mapResponse(a)) {
			return answers, fmt.Errorf("Problem extracting values for label: %s from question: %s", labels, field.Title)

		}

		return a, nil
	}

	for i, label := range labels {
		answers[i] = &Answer{label, label}
	}
	return answers, nil

}

func MakeMCTranslator(src *Field, dst *Field) (map[string]string, error) {
	fields := []*Field{src, dst}
	ans := make([][]*Answer, len(fields))

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

func findField(ref string, form *Form) (*Field, error) {
	for _, f := range form.Fields {
		if f.Ref == ref {
			return f, nil
		}
	}
	return nil, &FormTranslationError{fmt.Sprintf("Could not find field ref %v in form titled %v", ref, form.Title)}
}

func prepForms(a, b *Form) {
	a.Fields = append(a.Fields, a.ThankYouScreens...)
	b.Fields = append(b.Fields, b.ThankYouScreens...)
}

func makeTranslator(form, destForm *Form, byRef bool) (*FormTranslator, error) {
	prepForms(form, destForm)

	if !byRef && len(form.Fields) != len(destForm.Fields){
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

func MakeTranslatorByShape(form, destForm *Form) (*FormTranslator, error) {
	return makeTranslator(form, destForm, false)
}

func MakeTranslatorByRef(form, destForm *Form) (*FormTranslator, error) {
	return makeTranslator(form, destForm, true)
}
