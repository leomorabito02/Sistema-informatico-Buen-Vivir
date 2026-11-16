package controllers

import (
	"errors"
	"encoding/json"
	"net/http"
	"strings"

	"auth/internal/dto"
	"auth/internal/services"
	"auth/internal/utils"

	"github.com/go-chi/chi/v5"
)

type UserController struct {
	userService services.UserService
}

func NewUserController(userService services.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Formato de solicitud inválido")
		return
	}

	user, err := c.userService.CreateUser(r.Context(), &req)
	if err != nil {
		if validationErrors := utils.FormatValidationError(err); len(validationErrors) > 0 {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, validationErrors)
			return
		}
		utils.JSONResponse(w, errorStatus(err, http.StatusInternalServerError), false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusCreated, true, user, nil)
}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	user, err := c.userService.GetUserByID(r.Context(), id)
	if err != nil {
		utils.JSONResponse(w, errorStatus(err, http.StatusNotFound), false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, user, nil)
}

func (c *UserController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// ?q= busca por nombre, apellido, email o teléfono y ?role= filtra por rol
	// (por ejemplo role=role_teacher para el selector de docentes).
	q := r.URL.Query().Get("q")
	if len([]rune(q)) > services.MaxBusquedaUsuarios {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "La búsqueda es demasiado larga")
		return
	}

	users, err := c.userService.GetAllUsers(r.Context())
	if err != nil {
		utils.JSONResponse(w, errorStatus(err, http.StatusInternalServerError), false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, services.FiltrarUsuarios(users, q, r.URL.Query().Get("role")), nil)
}

func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Formato de solicitud inválido")
		return
	}

	user, err := c.userService.UpdateUser(r.Context(), id, &req)
	if err != nil {
		if validationErrors := utils.FormatValidationError(err); len(validationErrors) > 0 {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, validationErrors)
			return
		}
		utils.JSONResponse(w, errorStatus(err, http.StatusInternalServerError), false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, user, nil)
}

func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	if err := c.userService.DeleteUser(r.Context(), id); err != nil {
		utils.JSONResponse(w, errorStatus(err, http.StatusInternalServerError), false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, map[string]string{"message": "Usuario eliminado correctamente"}, nil)
}

// GetParticipantsHistory responde GET /api/v1/users/participants/history
// (solo admin): todos los participantes, también los dados de baja, con
// sus fechas. Lo consume el servicio de Dashboard.
func (c *UserController) GetParticipantsHistory(w http.ResponseWriter, r *http.Request) {
	rows, err := c.userService.GetParticipantsHistory(r.Context())
	if err != nil {
		utils.JSONResponse(w, errorStatus(err, http.StatusInternalServerError), false, nil, err.Error())
		return
	}
	utils.JSONResponse(w, http.StatusOK, true, rows, nil)
}

func (c *UserController) ApproveUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	// PATCH /users/{id}/approval acepta {"approval_status":"approved"|"rejected"};
	// POST /users/{id}/approve sin cuerpo aprueba (compatibilidad).
	var req dto.ApproveUserRequest
	if r.Body != nil {
		_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req)
	}
	var user *dto.UserResponse
	var err error
	switch req.ApprovalStatus {
	case "", "approved":
		user, err = c.userService.ApproveUser(r.Context(), id)
	case "rejected":
		user, err = c.userService.RejectUser(r.Context(), id)
	default:
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "approval_status debe ser approved o rejected")
		return
	}
	if err != nil {
		utils.JSONResponse(w, errorStatus(err, http.StatusInternalServerError), false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, user, nil)
}

// errorStatus devuelve 403 para los errores de permisos del servicio
// (empiezan con "forbidden") y fallback para el resto.
func errorStatus(err error, fallback int) int {
	if strings.HasPrefix(err.Error(), "forbidden") {
		return http.StatusForbidden
	}
	if errors.Is(err, services.ErrValidacion) {
		return http.StatusBadRequest
	}
	return fallback
}
