package controllers

import (
	"encoding/json"
	"net/http"

	"auth/internal/dto"
	"auth/internal/services"
	"auth/internal/utils"

	"github.com/go-chi/chi/v5"
)

type TeacherController struct {
	teacherService services.TeacherService
}

func NewTeacherController(teacherService services.TeacherService) *TeacherController {
	return &TeacherController{teacherService: teacherService}
}

func (c *TeacherController) CreateTeacher(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTeacherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Formato de solicitud inválido")
		return
	}

	teacher, err := c.teacherService.CreateTeacher(r.Context(), &req)
	if err != nil {
		if validationErrors := utils.FormatValidationError(err); len(validationErrors) > 0 {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, validationErrors)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusCreated, true, teacher, nil)
}

func (c *TeacherController) GetTeacher(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	teacher, err := c.teacherService.GetTeacherByID(r.Context(), id)
	if err != nil {
		utils.JSONResponse(w, http.StatusNotFound, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, teacher, nil)
}

func (c *TeacherController) GetAllTeachers(w http.ResponseWriter, r *http.Request) {
	teachers, err := c.teacherService.GetAllTeachers(r.Context())
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, teachers, nil)
}

func (c *TeacherController) UpdateTeacher(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	var req dto.UpdateTeacherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Formato de solicitud inválido")
		return
	}

	teacher, err := c.teacherService.UpdateTeacher(r.Context(), id, &req)
	if err != nil {
		if validationErrors := utils.FormatValidationError(err); len(validationErrors) > 0 {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, validationErrors)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, teacher, nil)
}

func (c *TeacherController) DeleteTeacher(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	if err := c.teacherService.DeleteTeacher(r.Context(), id); err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, map[string]string{"message": "Profesor eliminado correctamente"}, nil)
}
