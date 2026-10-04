package dto

import (
	"errors"
	"html"
	"strings"

	"buenvivir/rxresume/internal/models"
)

type GenerateResumeRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone"`
	Location string `json:"location"`
	Headline string `json:"headline"`
}

func (p *GenerateResumeRequest) Sanitize() {
	p.Name = html.EscapeString(strings.TrimSpace(p.Name))
	p.Email = html.EscapeString(strings.TrimSpace(p.Email))
	p.Phone = html.EscapeString(strings.TrimSpace(p.Phone))
	p.Location = html.EscapeString(strings.TrimSpace(p.Location))
	p.Headline = html.EscapeString(strings.TrimSpace(p.Headline))
}

func (p *GenerateResumeRequest) Validate() error {
	if len(p.Name) > 200 {
		return errors.New("el nombre excede la longitud máxima")
	}
	return nil
}

type UpdateResumeRequest struct {
	models.ResumeData
}

func (r *UpdateResumeRequest) Sanitize() {
	r.Basics.Name = html.EscapeString(strings.TrimSpace(r.Basics.Name))
	r.Basics.Email = html.EscapeString(strings.TrimSpace(r.Basics.Email))
	r.Basics.Phone = html.EscapeString(strings.TrimSpace(r.Basics.Phone))
	r.Basics.Headline = html.EscapeString(strings.TrimSpace(r.Basics.Headline))
	r.Basics.Location = html.EscapeString(strings.TrimSpace(r.Basics.Location))
	r.Basics.Summary = html.EscapeString(strings.TrimSpace(r.Basics.Summary))
	for i := range r.Work {
		r.Work[i].Company = html.EscapeString(strings.TrimSpace(r.Work[i].Company))
		r.Work[i].Position = html.EscapeString(strings.TrimSpace(r.Work[i].Position))
		r.Work[i].Description = html.EscapeString(strings.TrimSpace(r.Work[i].Description))
	}
	for i := range r.Education {
		r.Education[i].Institution = html.EscapeString(strings.TrimSpace(r.Education[i].Institution))
		r.Education[i].Degree = html.EscapeString(strings.TrimSpace(r.Education[i].Degree))
	}
}

func (r *UpdateResumeRequest) Validate() error {
	if len(r.Basics.Name) > 200 {
		return errors.New("el nombre excede la longitud máxima")
	}
	if len(r.Basics.Email) > 200 {
		return errors.New("el correo excede la longitud máxima")
	}
	return nil
}
