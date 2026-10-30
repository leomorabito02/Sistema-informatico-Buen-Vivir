package crypto

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// LoadKey resuelve la clave AES-256. En Cloud Run se lee de Secret
// Manager (secretName = "projects/<id>/secrets/<nombre>/versions/<v>");
// en local, sin Secret Manager, se acepta la clave en base64 directo
// desde el entorno. En ambos casos el valor guardado es la clave de 32
// bytes codificada en base64, nunca la clave cruda en el código.
func LoadKey(ctx context.Context, secretName, base64Key string) ([]byte, error) {
	encoded := base64Key
	if secretName != "" {
		value, err := accessSecret(ctx, secretName)
		if err != nil {
			return nil, err
		}
		encoded = value
	}
	if encoded == "" {
		return nil, errors.New("falta la clave de cifrado: definí ENCRYPTION_KEY_SECRET (Secret Manager) o ENCRYPTION_KEY (base64, solo desarrollo)")
	}

	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, fmt.Errorf("la clave de cifrado no es base64 válido: %w", err)
	}
	if len(key) != KeySize {
		return nil, fmt.Errorf("la clave de cifrado debe tener %d bytes, tiene %d", KeySize, len(key))
	}
	return key, nil
}

func accessSecret(ctx context.Context, secretName string) (string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("no se pudo crear el cliente de Secret Manager: %w", err)
	}
	defer func() { _ = client.Close() }()

	resp, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: secretName})
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el secreto de la clave de cifrado: %w", err)
	}
	return string(resp.GetPayload().GetData()), nil
}
