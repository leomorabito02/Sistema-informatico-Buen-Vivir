package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"buenvivir/education/internal/models"
)

func TestTeacherNames_CompletaYCachea(t *testing.T) {
	llamadas := 0
	tn := NewTeacherNames(func(_ context.Context, ids []string) (map[string]string, error) {
		llamadas++
		return map[string]string{"usr-doc": "Teresa Docente"}, nil
	})
	cursos := []models.Course{
		{ID: "c1", ResponsibleTeacher: "usr-doc"},
		{ID: "c2", ResponsibleTeacher: "usr-doc"},
		{ID: "c3"},
	}
	tn.Fill(context.Background(), cursos)
	if cursos[0].ResponsibleTeacherName != "Teresa Docente" || cursos[1].ResponsibleTeacherName != "Teresa Docente" {
		t.Fatalf("no completó el nombre: %+v", cursos)
	}
	if cursos[2].ResponsibleTeacherName != "" {
		t.Error("un curso sin docente no debe tener nombre")
	}
	if cursos[0].ResponsibleTeacher != "usr-doc" {
		t.Error("el id del docente tiene que seguir en la respuesta")
	}
	uno := &models.Course{ResponsibleTeacher: "usr-doc"}
	tn.FillOne(context.Background(), uno)
	if uno.ResponsibleTeacherName != "Teresa Docente" || llamadas != 1 {
		t.Errorf("debía salir del caché: nombre=%q llamadas=%d", uno.ResponsibleTeacherName, llamadas)
	}
}

func TestTeacherNames_VenceElCache(t *testing.T) {
	llamadas := 0
	tn := NewTeacherNames(func(context.Context, []string) (map[string]string, error) {
		llamadas++
		return map[string]string{"u": "Nombre"}, nil
	})
	ahora := time.Now()
	tn.now = func() time.Time { return ahora }
	tn.Fill(context.Background(), []models.Course{{ResponsibleTeacher: "u"}})
	ahora = ahora.Add(11 * time.Minute)
	tn.Fill(context.Background(), []models.Course{{ResponsibleTeacher: "u"}})
	if llamadas != 2 {
		t.Errorf("después de 10 minutos debía volver a consultar, llamadas=%d", llamadas)
	}
}

func TestTeacherNames_SiAuthFallaNoRompe(t *testing.T) {
	llamadas := 0
	tn := NewTeacherNames(func(context.Context, []string) (map[string]string, error) {
		llamadas++
		return nil, errors.New("auth caído")
	})
	cursos := []models.Course{{ResponsibleTeacher: "u"}}
	tn.Fill(context.Background(), cursos)
	if cursos[0].ResponsibleTeacherName != "" {
		t.Error("sin auth el nombre queda vacío")
	}
	tn.Fill(context.Background(), cursos)
	if llamadas != 2 {
		t.Error("un fallo no se cachea: debe reintentar en el próximo pedido")
	}
}

func TestTeacherNames_NilNoHaceNada(t *testing.T) {
	var tn *TeacherNames
	tn.Fill(context.Background(), []models.Course{{ResponsibleTeacher: "u"}})
	tn.FillOne(context.Background(), nil)
	NewTeacherNames(nil).Fill(context.Background(), []models.Course{{ResponsibleTeacher: "u"}})
}
