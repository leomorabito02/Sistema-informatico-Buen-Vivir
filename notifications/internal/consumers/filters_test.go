package consumers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/notifications/internal/consumers"
	"buenvivir/notifications/internal/events"
)

// TestFiltrosDeInfraCoincidenConLosConsumidores verifies that the subscription
// filters in infra/pubsub/setup.sh exactly match the event types registered by
// each consumer in notifications.
func TestFiltrosDeInfraCoincidenConLosConsumidores(t *testing.T) {
	filters, err := events.InfraFilters("../../../infra/pubsub/setup.sh")
	require.NoError(t, err)

	eduConsumer := consumers.NewEducacionConsumer(events.NoVerifier{}, nil, nil, nil, nil, nil, nil)
	assert.ElementsMatch(t, eduConsumer.EventTypes(), filters["notificaciones-educacion"])

	phConsumer := consumers.NewPromocionHumanaConsumer(events.NoVerifier{}, nil, nil, nil, nil, nil, nil)
	assert.ElementsMatch(t, phConsumer.EventTypes(), filters["notificaciones-promocion-humana"])

	usrConsumer := consumers.NewUsuariosConsumer(events.NoVerifier{}, nil, nil, nil, nil, nil, nil)
	assert.ElementsMatch(t, usrConsumer.EventTypes(), filters["notificaciones-usuarios"])

	laborConsumer := consumers.NewInsercionLaboralConsumer(events.NoVerifier{}, nil, nil, nil, nil, nil, nil)
	assert.ElementsMatch(t, laborConsumer.EventTypes(), filters["notificaciones-insercion-laboral"])

	campConsumer := consumers.NewCampanasConsumer(events.NoVerifier{}, nil, nil, nil, nil, nil, nil)
	assert.ElementsMatch(t, campConsumer.EventTypes(), filters["notificaciones-campanas"])
}

