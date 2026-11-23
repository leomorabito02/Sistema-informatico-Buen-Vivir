package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// defaultUploadBaseURL es el host de la Upload API de Contentful (los tests lo reemplazan).
const defaultUploadBaseURL = "https://upload.contentful.com"

// UploadAsset sube un binario a Contentful y devuelve el id del Asset
// ya procesado y publicado, listo para usarse como material.file
// (fileAssetId). Encapsula el flujo de 5 pasos de la CMA:
//  1. subir el binario -> uploadId
//  2. crear el Asset linkeando ese upload
//  3. disparar el procesamiento del archivo
//  4. esperar a que el procesamiento termine (poll)
//  5. publicar el Asset
func (c *ContentfulClient) UploadAsset(ctx context.Context, title, fileName, contentType string, data []byte) (string, error) {
	uploadID, err := c.uploadBinary(ctx, data)
	if err != nil {
		return "", fmt.Errorf("no se pudo subir el archivo a Contentful: %w", err)
	}

	assetBody, _ := json.Marshal(map[string]any{
		"fields": map[string]any{
			"title": wrapLocale(c.locale, title),
			"file": wrapLocale(c.locale, map[string]any{
				"contentType": contentType,
				"fileName":    fileName,
				"uploadFrom": map[string]any{
					"sys": map[string]any{"type": "Link", "linkType": "Upload", "id": uploadID},
				},
			}),
		},
	})
	body, _, err := c.doCMA(ctx, http.MethodPost, "/assets", nil, assetBody)
	if err != nil {
		return "", err
	}
	var created struct {
		Sys struct {
			ID      string `json:"id"`
			Version int    `json:"version"`
		} `json:"sys"`
	}
	if err := json.Unmarshal(body, &created); err != nil || created.Sys.ID == "" {
		return "", fmt.Errorf("respuesta inesperada al crear el asset: %s", string(body))
	}

	if _, _, err := c.doCMA(ctx, http.MethodPut,
		fmt.Sprintf("/assets/%s/files/%s/process", created.Sys.ID, c.locale),
		map[string]string{"X-Contentful-Version": strconv.Itoa(created.Sys.Version)}, nil); err != nil {
		return "", fmt.Errorf("no se pudo iniciar el procesamiento del asset: %w", err)
	}

	// 4. poll hasta que fields.file[locale].url exista
	version := created.Sys.Version + 1
	processed := false
	for i := 0; i < 25; i++ {
		time.Sleep(800 * time.Millisecond)
		b, _, err := c.doCMA(ctx, http.MethodGet, "/assets/"+created.Sys.ID, nil, nil)
		if err != nil {
			return "", err
		}
		var a struct {
			Sys struct {
				Version int `json:"version"`
			} `json:"sys"`
			Fields struct {
				File map[string]struct {
					URL string `json:"url"`
				} `json:"file"`
			} `json:"fields"`
		}
		if err := json.Unmarshal(b, &a); err != nil {
			return "", err
		}
		if f, ok := a.Fields.File[c.locale]; ok && f.URL != "" {
			version = a.Sys.Version
			processed = true
			break
		}
	}
	if !processed {
		return "", fmt.Errorf("el archivo no terminó de procesarse en Contentful (reintentá)")
	}

	if _, _, err := c.doCMA(ctx, http.MethodPut, "/assets/"+created.Sys.ID+"/published",
		map[string]string{"X-Contentful-Version": strconv.Itoa(version)}, nil); err != nil {
		return "", fmt.Errorf("el asset se subió pero no se pudo publicar: %w", err)
	}

	return created.Sys.ID, nil
}

func (c *ContentfulClient) uploadBinary(ctx context.Context, data []byte) (string, error) {
	reqURL := fmt.Sprintf("%s/spaces/%s/uploads", c.uploadBaseURL, c.spaceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.cmaToken)
	req.Header.Set("Content-Type", "application/octet-stream")

	body, _, err := c.do(req)
	if err != nil {
		return "", err
	}
	var up struct {
		Sys struct {
			ID string `json:"id"`
		} `json:"sys"`
	}
	if err := json.Unmarshal(body, &up); err != nil || up.Sys.ID == "" {
		return "", fmt.Errorf("contentful no devolvió un upload id: %s", string(body))
	}
	return up.Sys.ID, nil
}
