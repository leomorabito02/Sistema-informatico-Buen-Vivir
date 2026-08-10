package controllers

import (
	"encoding/json"
	"net/http"

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
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
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
		utils.JSONResponse(w, http.StatusNotFound, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, user, nil)
}

func (c *UserController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.userService.GetAllUsers(r.Context())
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, users, nil)
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
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
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
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, map[string]string{"message": "Usuario eliminado correctamente"}, nil)
}
