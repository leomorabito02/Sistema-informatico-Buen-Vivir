package repositories

import (
	"errors"
	"net/http"
	"testing"

	"buenvivir/education/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestContentfulAPIError_Error(t *testing.T) {
	err := &ContentfulAPIError{StatusCode: http.StatusUnprocessableEntity, Body: `{"message":"invalid"}`}
	assert.Equal(t, `contentful respondió 422: {"message":"invalid"}`, err.Error())
}

func TestIsConflict(t *testing.T) {
	assert.True(t, IsConflict(&ContentfulAPIError{StatusCode: http.StatusConflict}))
	assert.False(t, IsConflict(&ContentfulAPIError{StatusCode: http.StatusNotFound}))
	assert.False(t, IsConflict(errors.New("otro error")))
	assert.False(t, IsConflict(nil))
}

func TestIsNotFound_RecognisesBothSources(t *testing.T) {
	assert.True(t, IsNotFound(ErrNotFound))
	assert.True(t, IsNotFound(&ContentfulAPIError{StatusCode: http.StatusNotFound}))
	assert.False(t, IsNotFound(&ContentfulAPIError{StatusCode: http.StatusConflict}))
	assert.False(t, IsNotFound(errors.New("cualquier cosa")))
}

func TestNewContentfulClient_DefaultsAndLocaleFallback(t *testing.T) {
	c := NewContentfulClient(&config.Config{
		ContentfulSpaceID:     "space-1",
		ContentfulEnvironment: config.ContentfulEnvTesting,
		ContentfulCDAToken:    "cda",
		ContentfulCMAToken:    "cma",
		ContentfulLocale:      "es-AR",
	})

	assert.Equal(t, "es-AR", c.Locale())
	assert.Equal(t, defaultCDABaseURL, c.cdaBaseURL)
	assert.Equal(t, defaultCMABaseURL, c.cmaBaseURL)
	assert.Equal(t, defaultUploadBaseURL, c.uploadBaseURL)
	assert.NotNil(t, c.httpClient)
}
