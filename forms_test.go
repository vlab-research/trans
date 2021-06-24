package trans

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractAnswersGetsSimpleLabels(t *testing.T) {
	field := `{
        "title": "What is your gender? ",
 		"ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
		"properties": {
		"choices": [{"label": "Male"},
			{"label": "Female"},
			{"label": "Other"}]},
		"type": "multiple_choice"}`

	f := new(Field)
	json.Unmarshal([]byte(field), f)

	res, err := ExtractAnswers(f)
	assert.Nil(t, err)
	assert.Equal(t, []Answer{{"Male", "Male"}, {"Female", "Female"}, {"Other", "Other"}}, res)
}

func TestExtractAnswersDoesntFailIfNoAnswers(t *testing.T) {
	field := `{
        "title": "What is your gender? ",
 		"ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
		"properties": {
		"choices": []},
		"type": "multiple_choice"}`

	f := new(Field)
	json.Unmarshal([]byte(field), f)

	res, err := ExtractAnswers(f)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestDoesntTranslateCertainFieldTypesLikeOpinionScale(t *testing.T) {
	fields := []string{
		`{"id": "YmJEQUEqh0h1", "properties": {"labels": {"left": "Not at all concerned", "right": "Very concerned"}, "start_at_one": true, "steps": 5},
                  "ref": "9ddb9864-e684-4c69-8dfe-24648ce5a6a0",
                  "title": "How concerned are you about getting infected with COVID-19?",
                  "type": "opinion_scale",
                  "validations": {"required": false}}`,
	}

	for _, field := range fields {
		f := new(Field)
		json.Unmarshal([]byte(field), f)

		res, err := MakeFieldTranslator(f, f)

		assert.Nil(t, err)
		assert.False(t, res.Translate)
	}

}

func TestExtractAnswersGetsFromText(t *testing.T) {
	fields := []string{
		// Dash before
		`{"title": "Which state do you currently live in?\n- A. foo 91  bar\n- B. Jharkhand\n- C. Odisha\n- D. Uttar Pradesh",
		"ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
		"properties": {
		"choices": [{"label": "A"},
			{"label": "B"},
			{"label": "C"},
			{"label": "D"}]},
		"type": "multiple_choice"}`,

		// No dash before, followed by . symbol
		`{"title": "Which state do you currently live in?\nA. foo 91  bar\nB. Jharkhand\nC. Odisha\nD. Uttar Pradesh",
         "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
         "properties": {
             "choices": [{"label": "A"},
                         {"label": "B"},
                         {"label": "C"},
                         {"label": "D"}]},
         "type": "multiple_choice"}`,

		// followed by - symbol
		`{"title": "Which state do you currently live in?\nA- foo 91  bar\nB- Jharkhand\nC- Odisha\nD- Uttar Pradesh",
         "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
         "properties": {
             "choices": [{"label": "A"},
                         {"label": "B"},
                         {"label": "C"},
                         {"label": "D"}]},
         "type": "multiple_choice"}`,

		// tabs instead of spaces
		`{"title": "Which state do you currently live in?\nA-\tfoo 91  bar\nB-\tJharkhand\nC-\tOdisha\nD-\tUttar Pradesh",
         "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
         "properties": {
             "choices": [{"label": "A"},
                         {"label": "B"},
                         {"label": "C"},
                         {"label": "D"}]},
         "type": "multiple_choice"}`,

		// No space
		`{"title": "Which state do you currently live in?\nA-foo 91  bar\nB-Jharkhand\nC-Odisha\nD-Uttar Pradesh",
         "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
         "properties": {
             "choices": [{"label": "A"},
                         {"label": "B"},
                         {"label": "C"},
                         {"label": "D"}]},
         "type": "multiple_choice"}`,

		// Lots of symbols
		`{"title": "Which state do you currently live in?\n- A.. foo 91  bar\n- B.) Jharkhand\n- C. Odisha\n- D. Uttar Pradesh",
		"ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
		"properties": {
		"choices": [{"label": "A"},
			{"label": "B"},
			{"label": "C"},
			{"label": "D"}]},
		"type": "multiple_choice"}`,
	}

	for _, field := range fields {
		f := new(Field)
		json.Unmarshal([]byte(field), f)

		res, _ := ExtractAnswers(f)
		expected := []Answer{{"A", "foo 91  bar"}, {"B", "Jharkhand"}, {"C", "Odisha"}, {"D", "Uttar Pradesh"}}
		assert.Equal(t, expected, res)
	}
}

func TestExtractAnswersErrorsWhenBadFormat(t *testing.T) {
	fields := []string{

		// no \n newline character before letters
		`{"title": "Which state do you currently live in?A. foo 91  bar\nB. Jharkhand\nC. Odisha\nD. Uttar Pradesh",
		"ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
		"properties": {
		    "choices": [{"label": "A"},
		                {"label": "B"},
		                {"label": "C"},
		                {"label": "D"}]},
		"type": "multiple_choice"}`,

		// Multiple matches for a character
		`{"title": "Which state do you currently live in?\nA. foo 91  bar\nB. Jharkhand\nC. Odisha\nA. Uttar Pradesh",
		"ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
		"properties": {
		    "choices": [{"label": "A"},
		                {"label": "B"},
		                {"label": "C"},
		                {"label": "D"}]},
		"type": "multiple_choice"}`,

		// Missing one label
		`{"title": "Which state do you currently live in?\nA- foo 91  bar\nB- Jharkhand\nC- Odisha\n",
		"ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
		"properties": {
		    "choices": [{"label": "A"},
		                {"label": "B"},
		                {"label": "C"},
		                {"label": "D"}]},
		"type": "multiple_choice"}`,

		// no space or symbol
		`{"title": "Which state do you currently live in?\nAfoo 91  bar\nB- Jharkhand\nC- Odisha\nD- Uttar Pradesh",
		"ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
		"properties": {
		    "choices": [{"label": "A"},
		                {"label": "B"},
		                {"label": "C"},
		                {"label": "D"}]},
		"type": "multiple_choice"}`,
	}

	for _, field := range fields {
		f := new(Field)
		json.Unmarshal([]byte(field), f)

		_, err := ExtractAnswers(f)
		assert.NotNil(t, err)
	}
}

func TestMakeMCTranslatorTranslatesLanguages(t *testing.T) {
	jsons := []string{
		`{"id": "vjl6LihKMtcX",
         "title": "आपका लिंग क्या है? ",
         "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
         "properties": {"choices": [{"label": "पुरुष"},
                                    {"label": "महिला"},
                                    {"label": "अन्य"}]},
         "type": "multiple_choice"}`,

		`{"title": "What is your gender? ",
          "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
          "properties": {
              "choices": [{"label": "Male"},
                          {"label": "Female"},
                          {"label": "Other"}]},
          "type": "multiple_choice"}`}

	fields := []*Field{}
	for _, j := range jsons {
		f := new(Field)
		json.Unmarshal([]byte(j), f)
		t.Log(f)
		fields = append(fields, f)
	}

	tr, err := MakeMCTranslator(fields[0], fields[1])
	assert.Nil(t, err)
	res, ok := tr["महिला"]
	assert.True(t, ok)
	assert.Equal(t, "Female", res)

	tr, err = MakeMCTranslator(fields[1], fields[0])
	assert.Nil(t, err)
	res, ok = tr["Female"]
	assert.True(t, ok)
	assert.Equal(t, "महिला", res)
}

func TestMakeMCTranslatosErrorsOnDifferentNumberOfAnswers(t *testing.T) {
	jsons := []string{
		`{"id": "vjl6LihKMtcX",
         "title": "आपका लिंग क्या है? ",
         "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
         "properties": {"choices": [{"label": "पुरुष"},
                                    {"label": "महिला"},
                                    {"label": "अन्य"}]},
         "type": "multiple_choice"}`,

		`{"title": "What is your gender? ",
          "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
          "properties": {
              "choices": [{"label": "Male"},
                          {"label": "Female"}]},
          "type": "multiple_choice"}`}

	fields := []*Field{}
	for _, j := range jsons {
		f := new(Field)
		json.Unmarshal([]byte(j), f)
		t.Log(f)
		fields = append(fields, f)
	}

	tr, err := MakeMCTranslator(fields[0], fields[1])
	assert.NotNil(t, err)
	assert.Nil(t, tr)
}

func TestMakeMCTranslatorReturnsTranslatorThatErrorsOnBadAnswer(t *testing.T) {
	jsons := []string{
		`{"id": "vjl6LihKMtcX",
         "title": "आपका लिंग क्या है? ",
         "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
         "properties": {"choices": [{"label": "पुरुष"},
                                    {"label": "महिला"}]},
         "type": "multiple_choice"}`,

		`{"title": "What is your gender? ",
          "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
          "properties": {
              "choices": [{"label": "Male"},
                          {"label": "Female"}]},
          "type": "multiple_choice"}`}

	fields := []*Field{}
	for _, j := range jsons {
		f := new(Field)
		json.Unmarshal([]byte(j), f)
		t.Log(f)
		fields = append(fields, f)
	}

	tr, err := MakeMCTranslator(fields[0], fields[1])
	assert.Nil(t, err)
	res, ok := tr["foo"]

	assert.False(t, ok)
	assert.Equal(t, "", res)
}

func TestMakeMCTranslatorTranslatesLanguagesWithLabels(t *testing.T) {
	jsons := []string{
		`{"id": "mdUpJMSY8Lct",
           "title": "वर्तमान में आप किस राज्य में रहते हैं?\n- A. छत्तीसगढ़\n- B. झारखंड\n- C. ओडिशा\n- D. उत्तर प्रदेश",
           "ref": "e959559b-092a-434f-b67f-dca329fab50a",
           "properties": {"choices": [{"label": "A"},
                                      {"label": "B"},
                                      {"label": "C"},
                                      {"label": "D"}]},
           "type": "multiple_choice"}`,

		`{"title": "Which state do you currently live in?\n- A. foo 91  bar\n- B. Jharkhand\n- C. Odisha\n- D. Uttar Pradesh",
           "ref": "20218ad0-96c8-4799-bdfe-90c689c5c206",
           "properties": {"choices": [{"label": "A"},
                                      {"label": "B"},
                                      {"label": "C"},
                                      {"label": "D"}]},
           "type": "multiple_choice"}`}

	fields := []*Field{}
	for _, j := range jsons {
		f := new(Field)
		json.Unmarshal([]byte(j), f)
		t.Log(f)
		fields = append(fields, f)
	}

	tr, err := MakeMCTranslator(fields[0], fields[1])
	assert.Nil(t, err)
	res, ok := tr["B"]
	assert.True(t, ok)
	assert.Equal(t, "Jharkhand", res)

	tr, err = MakeMCTranslator(fields[1], fields[0])
	assert.Nil(t, err)
	res, ok = tr["B"]
	assert.True(t, ok)
	assert.Equal(t, "झारखंड", res)
}

func TestMakeFormTranslatorByShape(t *testing.T) {

	jsons := []string{
		`{"fields": [
          {"id": "vjl6LihKMtcX",
          "title": "आपका लिंग क्या है? ",
          "ref": "foo",
          "properties": {"choices": [{"label": "पुरुष"},
                                    {"label": "महिला"},
                                    {"label": "अन्य"}]},
          "type": "multiple_choice"},
          {"id": "mdUpJMSY8Lct",
           "title": "वर्तमान में आप किस राज्य में रहते हैं?\n- A. छत्तीसगढ़\n- B. झारखंड\n- C. ओडिशा\n- D. उत्तर प्रदेश",
           "ref": "bar",
           "properties": {"choices": [{"label": "A"},
                                      {"label": "B"},
                                      {"label": "C"},
                                      {"label": "D"}]},
           "type": "multiple_choice"},
          {"id": "mdUpJMSY8Lct",
           "title": "वर्तमान में आप किस राज्य में रहते हैं?",
           "ref": "baz",
           "properties": {},
           "type": "number"}],
         "thankyou_screens": [
	  {
	    "id": "DefaultTyScreen",
	    "ref": "default_tys",
	    "title": "Done! Your information was sent perfectly.",
	    "properties": {
	      "show_button": false,
	      "share_icons": false
	    }
	  }
	]}`,

		`{"fields": [
          {"title": "What is your gender? ",
           "ref": "eng_foo",
           "properties": {
              "choices": [{"label": "Male"},
                          {"label": "Female"},
                          {"label": "Other"}]},
           "type": "multiple_choice"},
          {"title": "Which state do you currently live in?\n- A. foo 91  bar\n- B. Jharkhand\n- C. Odisha\n- D. Uttar Pradesh",
           "ref": "eng_bar",
           "properties": {"choices": [{"label": "A"},
                                      {"label": "B"},
                                      {"label": "C"},
                                      {"label": "D"}]},
           "type": "multiple_choice"},
           {"title": "How old are you?",
           "ref": "eng_baz",
           "properties": {},
           "type": "number"}],
         "thankyou_screens": [
	  {
	    "id": "DefaultTyScreen",
	    "ref": "default_tys",
	    "title": "Done! Your information was sent perfectly.",
	    "properties": {
	      "show_button": false,
	      "share_icons": false
	    }
	  }
	]}`,
	}

	forms := []FormJson{}
	for _, j := range jsons {
		f := new(FormJson)
		json.Unmarshal([]byte(j), f)
		forms = append(forms, *f)
	}

	ft, err := MakeTranslatorByShape(&forms[0], &forms[1])
	assert.Nil(t, err)
	assert.Equal(t, "Uttar Pradesh", ft.Fields["bar"].Mapping["D"])
	assert.Equal(t, "Male", ft.Fields["foo"].Mapping["पुरुष"])
	assert.Equal(t, false, ft.Fields["baz"].Translate)
	assert.Equal(t, 0, len(ft.Fields["baz"].Mapping))

	assert.Equal(t, false, ft.Fields["default_tys"].Translate)
	assert.Equal(t, 0, len(ft.Fields["default_tys"].Mapping))
}

func TestMakeFormTranslatorByRef(t *testing.T) {
	jsons := []string{
		`{"fields": [
          {"id": "vjl6LihKMtcX",
          "title": "आपका लिंग क्या है? ",
          "ref": "foo",
          "properties": {"choices": [{"label": "पुरुष"},
                                    {"label": "महिला"},
                                    {"label": "अन्य"}]},
          "type": "multiple_choice"},
          {"id": "mdUpJMSY8Lct",
           "title": "वर्तमान में आप किस राज्य में रहते हैं?\n- A. छत्तीसगढ़\n- B. झारखंड\n- C. ओडिशा\n- D. उत्तर प्रदेश",
           "ref": "bar",
           "properties": {"choices": [{"label": "A"},
                                      {"label": "B"},
                                      {"label": "C"},
                                      {"label": "D"}]},
           "type": "multiple_choice"},
          {"id": "mdUpJMSY8Lct",
           "title": "वर्तमान में आप किस राज्य में रहते हैं?",
           "ref": "baz",
           "properties": {},
           "type": "number"}]}`,

		`{"fields": [
          {"title": "Which state do you currently live in?\n- A. foo 91  bar\n- B. Jharkhand\n- C. Odisha\n- D. Uttar Pradesh",
           "ref": "bar",
           "properties": {"choices": [{"label": "A"},
                                      {"label": "B"},
                                      {"label": "C"},
                                      {"label": "D"}]},

           "type": "multiple_choice"},
           {"title": "How old are you?",
           "ref": "baz",
           "properties": {},
           "type": "number"},
           {"title": "What is your gender? ",
           "ref": "foo",
           "properties": {
              "choices": [{"label": "Male"},
                          {"label": "Female"},
                          {"label": "Other"}]},
           "type": "multiple_choice"}]}`}

	forms := []FormJson{}
	for _, j := range jsons {
		f := new(FormJson)
		json.Unmarshal([]byte(j), f)
		forms = append(forms, *f)
	}

	ft, err := MakeTranslatorByRef(&forms[0], &forms[1])
	assert.Nil(t, err)
	assert.Equal(t, "Uttar Pradesh", ft.Fields["bar"].Mapping["D"])
	assert.Equal(t, "Male", ft.Fields["foo"].Mapping["पुरुष"])
	assert.Equal(t, false, ft.Fields["baz"].Translate)
	assert.Equal(t, 0, len(ft.Fields["baz"].Mapping))
}

// DEAL WITH default_tys!!!
