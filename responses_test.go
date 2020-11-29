package trans

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTranslateWorksWithGoodData(t *testing.T) {

	ft := &FormTranslator{map[string]*FieldTranslator{
		"foo": {true, map[string]string{
			"A": "Makin that monay",
		}},
		"bar": {true, map[string]string{
			"man": "hombre",
		}},
		"baz": {false, map[string]string{}},
	}}

	res, err := Translate("foo", "A", ft)
	assert.Nil(t, err)
	assert.Equal(t, "Makin that monay", *res)

	res, err = Translate("bar", "man", ft)
	assert.Nil(t, err)
	assert.Equal(t, "hombre", *res)

	res, err = Translate("baz", "anything", ft)
	assert.Nil(t, err)
	assert.Equal(t, "anything", *res)
}

func TestTranslateReturnsNilIfInvalidAnswer(t *testing.T) {
	ft := &FormTranslator{map[string]*FieldTranslator{
		"foo": {true, map[string]string{
			"A": "Makin that monay",
		}},
	}}

	res, err := Translate("foo", "B", ft)
	assert.Nil(t, err)
	assert.Nil(t, res)
}

func TestTranslateErrorsIfImpossibleRef(t *testing.T) {
	ft := &FormTranslator{map[string]*FieldTranslator{
		"foo": {true, map[string]string{
			"A": "Makin that monay",
		}},
	}}

	res, err := Translate("baz", "B", ft)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}
