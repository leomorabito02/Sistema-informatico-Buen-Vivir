package repositories

import "errors"

// ErrNotFound señala que la campaña pedida no existe.
var ErrNotFound = errors.New("recurso no encontrado")
