package repositories

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/stretchr/testify/require"
)

const testFirestoreProject = "education-test"

// fakeCredentialsJSON es un JSON de credenciales mínimo y obviamente
// falso (sin private_key ni client_email — no es una service account
// key). Alcanza para ejercitar las ramas de NewFirestoreClient contra
// el emulador, que ignora las credenciales.
const fakeCredentialsJSON = `{"type":"authorized_user","project_id":"education-test"}`

// newTestFirestore devuelve un cliente apuntado al emulador de
// Firestore (variable FIRESTORE_EMULATOR_HOST). Si la variable no está
// seteada o el emulador no responde, salta el test: estos tests corren
// en CI con el emulador levantado como service container (ver
// .github/workflows/pipeline.yml, job ci-education) y localmente solo
// si el dev lo levanta a mano — igual que auth asume su Postgres.
func newTestFirestore(t *testing.T) *firestore.Client {
	t.Helper()

	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST no seteado: se omiten los tests contra el emulador de Firestore")
	}
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		t.Skipf("emulador de Firestore no accesible en %s: %v", host, err)
	}
	_ = conn.Close()

	client, err := firestore.NewClient(context.Background(), testFirestoreProject)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	return client
}

// wipeCollection deja la colección vacía antes de cada test: el
// emulador es un proceso compartido entre todos los tests del paquete,
// así que hay que aislar el estado a mano.
func wipeCollection(t *testing.T, client *firestore.Client, name string) {
	t.Helper()

	ctx := context.Background()
	docs, err := client.Collection(name).Documents(ctx).GetAll()
	require.NoError(t, err)
	for _, d := range docs {
		_, err := d.Ref.Delete(ctx)
		require.NoError(t, err)
	}
}
