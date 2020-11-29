package trans

import "fmt"

type TranslationError struct {
	Message string
}

func (e *TranslationError) Error() string {
	return e.Message
}

func Translate(qr, response string, ft *FormTranslator) (*string, error) {

	fieldTranslator, ok := ft.Fields[qr]
	if !ok {
		return nil, &TranslationError{fmt.Sprintf("Ref %v not found in translation mapping!", qr)}
	}

	// If not translate, return original message
	if !fieldTranslator.Translate {
		return &response, nil
	}

	// If not valid answer, dont error, just dont translate
	translated, ok := fieldTranslator.Mapping[response]
	if !ok {
		return nil, nil
	}

	return &translated, nil
}
