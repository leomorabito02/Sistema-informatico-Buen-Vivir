package repositories

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawString(t *testing.T) {
	fields := map[string]json.RawMessage{
		"title": json.RawMessage(`"Curso 1"`),
		"nil":   nil,
		"bad":   json.RawMessage(`123`),
	}

	assert.Equal(t, "Curso 1", rawString(fields, "title"))
	assert.Equal(t, "", rawString(fields, "ausente"))
	assert.Equal(t, "", rawString(fields, "nil"))
	assert.Equal(t, "", rawString(fields, "bad"))
}

func TestRawLink(t *testing.T) {
	fields := map[string]json.RawMessage{
		"trainingArea": json.RawMessage(`{"sys":{"type":"Link","linkType":"Entry","id":"area-1"}}`),
		"empty":        json.RawMessage(`{"sys":{"id":""}}`),
		"noJSON":       json.RawMessage(`"x"`),
	}

	id, ok := rawLink(fields, "trainingArea")
	assert.True(t, ok)
	assert.Equal(t, "area-1", id)

	_, ok = rawLink(fields, "ausente")
	assert.False(t, ok)

	_, ok = rawLink(fields, "empty")
	assert.False(t, ok)

	_, ok = rawLink(fields, "noJSON")
	assert.False(t, ok)
}

func TestRawLinks(t *testing.T) {
	fields := map[string]json.RawMessage{
		"material": json.RawMessage(`[
			{"sys":{"type":"Link","linkType":"Entry","id":"m1"}},
			{"sys":{"type":"Link","linkType":"Entry","id":""}},
			{"sys":{"type":"Link","linkType":"Entry","id":"m2"}}
		]`),
		"notArray": json.RawMessage(`{"sys":{"id":"x"}}`),
	}

	ids := rawLinks(fields, "material")
	assert.Equal(t, []string{"m1", "m2"}, ids, "ignora los links sin id")

	assert.Nil(t, rawLinks(fields, "ausente"))
	assert.Nil(t, rawLinks(fields, "notArray"))
}

func TestUnwrapLocale(t *testing.T) {
	assert.Nil(t, unwrapLocale(nil, "en-US"))
	assert.Nil(t, unwrapLocale(json.RawMessage(`"no es un objeto"`), "en-US"))

	raw := json.RawMessage(`{"en-US":"Hola"}`)
	assert.JSONEq(t, `"Hola"`, string(unwrapLocale(raw, "en-US")))

	// locale pedido ausente: cae al primero disponible
	fallback := json.RawMessage(`{"es-AR":"Chau"}`)
	assert.JSONEq(t, `"Chau"`, string(unwrapLocale(fallback, "en-US")))
}

func TestFlattenCMAFields(t *testing.T) {
	fields := map[string]json.RawMessage{
		"title": json.RawMessage(`{"en-US":"Curso"}`),
		"mode":  json.RawMessage(`{"en-US":"Virtual"}`),
		"vacio": json.RawMessage(`{}`),
	}

	flat := flattenCMAFields(fields, "en-US")
	assert.JSONEq(t, `"Curso"`, string(flat["title"]))
	assert.JSONEq(t, `"Virtual"`, string(flat["mode"]))
	_, ok := flat["vacio"]
	assert.False(t, ok, "los campos que quedan nil tras desenvolver se descartan")
}
