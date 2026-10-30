package models

import "time"

// Indicadores es lo que Promoción Humana le da al Dashboard (HU-52 a
// HU-54) para calcular "Fichas activas" y "Participantes con
// seguimiento". Lleva solo ids, estados y fechas: nunca la situación, los
// avances ni la descripción de los hitos (datos confidenciales).
type Indicadores struct {
	Fichas []IndicadorFicha `json:"fichas"`
	Hitos  []IndicadorHito  `json:"hitos"`
}

// Estados de una ficha para los indicadores.
const (
	IndicadorFichaActiva = "activa"
	IndicadorFichaBaja   = "baja"
)

type IndicadorFicha struct {
	ID             string     `json:"id"`
	ParticipanteID string     `json:"participante_id"`
	Estado         string     `json:"estado"` // "activa" | "baja"
	CreadaEn       time.Time  `json:"creada_en"`
	FechaBaja      *time.Time `json:"fecha_baja"`
}

type IndicadorHito struct {
	ParticipanteID string `json:"participante_id"`
	Fecha          string `json:"fecha"` // AAAA-MM-DD
}
