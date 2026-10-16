package producers

import (
	"context"
	"testing"

	"buenvivir/human-promotion/internal/events"
)

// Con la mensajería apagada (productor nil o sin writer) nada falla ni
// se publica. Los eventos con writer real se prueban en
// repositories/events_test.go contra el emulador.
func TestProductorApagadoNoPublica(t *testing.T) {
	for _, p := range []*PromocionHumana{nil, NewPromocionHumana(nil, "t")} {
		if p.Enabled() {
			t.Fatal("sin writer no publica")
		}
		p.Notify()
		ctx := context.Background()
		if err := p.FichaCreada(ctx, nil, "f", "p"); err != nil {
			t.Fatal(err)
		}
		if err := p.FichaActualizada(ctx, nil, "f", "p"); err != nil {
			t.Fatal(err)
		}
		for _, nuevo := range []bool{true, false} {
			if err := p.Hito(ctx, nil, nuevo, events.HitoPayload{HitoID: "h", ParticipanteID: "p"}); err != nil {
				t.Fatal(err)
			}
		}
	}
}
