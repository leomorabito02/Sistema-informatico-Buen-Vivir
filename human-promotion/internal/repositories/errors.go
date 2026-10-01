package repositories

import "errors"

// ErrNotFound señala que el documento pedido no existe.
var ErrNotFound = errors.New("recurso no encontrado")

// ErrAlreadyExists señala que el participante ya tiene una ficha
// (la relación ficha-participante es 1:1, HU-29).
var ErrAlreadyExists = errors.New("el participante ya tiene una ficha")
