package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/events"
)

type fakeContenidoPublisher struct {
	ids     []string
	payload events.ContenidoPayload
	err     error
}

func (f *fakeContenidoPublisher) Contenido(_ context.Context, eventID, _ string, payload events.ContenidoPayload) error {
	f.ids = append(f.ids, eventID)
	f.payload = payload
	return f.err
}

const webhookBody = `{"sys":{"id":"entry1","revision":2,"contentType":{"sys":{"id":"course"}}}}`

func webhookRequest(method, secret, topic, body string) *http.Request {
	r := httptest.NewRequest(method, "/webhooks/contentful", strings.NewReader(body))
	if secret != "" {
		r.Header.Set(HeaderWebhookSecret, secret)
	}
	r.Header.Set("X-Contentful-Topic", topic)
	return r
}

func TestContentfulWebhook_PublicaYResponde202(t *testing.T) {
	pub := &fakeContenidoPublisher{}
	c := NewContentfulWebhookController("s3cr3t", pub)

	rec := httptest.NewRecorder()
	c.Receive(rec, webhookRequest(http.MethodPost, "s3cr3t", "ContentManagement.Entry.publish", webhookBody))
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Len(t, pub.ids, 1)
	assert.Equal(t, "entry1", pub.payload.EntryID)
	assert.Contains(t, rec.Body.String(), pub.ids[0])
}

func TestContentfulWebhook_SecretoInvalido401(t *testing.T) {
	pub := &fakeContenidoPublisher{}
	for _, secret := range []string{"", "otro", "s3cr3t-mas-largo"} {
		rec := httptest.NewRecorder()
		NewContentfulWebhookController("s3cr3t", pub).Receive(rec, webhookRequest(http.MethodPost, secret, "ContentManagement.Entry.publish", webhookBody))
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "secreto %q", secret)
	}
	// Sin secreto configurado no acepta nada, ni siquiera un header vacío.
	rec := httptest.NewRecorder()
	NewContentfulWebhookController("", pub).Receive(rec, webhookRequest(http.MethodPost, "", "ContentManagement.Entry.publish", webhookBody))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, pub.ids)
}

func TestContentfulWebhook_OtrosCasos(t *testing.T) {
	pub := &fakeContenidoPublisher{}
	c := NewContentfulWebhookController("s3cr3t", pub)

	rec := httptest.NewRecorder()
	c.Receive(rec, webhookRequest(http.MethodGet, "s3cr3t", "", ""))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)

	rec = httptest.NewRecorder()
	c.Receive(rec, webhookRequest(http.MethodPost, "s3cr3t", "ContentManagement.Entry.save", webhookBody))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"ignorado":true`)

	rec = httptest.NewRecorder()
	c.Receive(rec, webhookRequest(http.MethodPost, "s3cr3t", "ContentManagement.Entry.publish", `{"sys":{}}`))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Empty(t, pub.ids)

	pub.err = errors.New("firestore caído")
	rec = httptest.NewRecorder()
	c.Receive(rec, webhookRequest(http.MethodPost, "s3cr3t", "ContentManagement.Entry.publish", webhookBody))
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "5xx para que Contentful reintente")
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("corte") }

func TestContentfulWebhook_CuerpoIlegible(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/webhooks/contentful", errReader{})
	r.Header.Set(HeaderWebhookSecret, "s3cr3t")
	rec := httptest.NewRecorder()
	NewContentfulWebhookController("s3cr3t", &fakeContenidoPublisher{}).Receive(rec, r)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
