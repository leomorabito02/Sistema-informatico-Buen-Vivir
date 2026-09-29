package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"html"
	"strings"
	"time"
)

type SocialProfile struct {
	ID       string `json:"id"`
	Network  string `json:"network"`
	Username string `json:"username"`
	URL      string `json:"url"`
}

type Basics struct {
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	Headline string          `json:"headline"`
	Location string          `json:"location"`
	Website  string          `json:"website"`
	Summary  string          `json:"summary"`
	Picture  string          `json:"picture"`
	Profiles []SocialProfile `json:"profiles"`
}

type WorkExperience struct {
	ID          string `json:"id"`
	Company     string `json:"company"`
	Position    string `json:"position"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Description string `json:"description"`
}

type EducationItem struct {
	ID          string `json:"id"`
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Field       string `json:"field"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Grade       string `json:"grade"`
	Status      string `json:"status"`
}

type Skill struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level string `json:"level"`
}

type Certification struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
}

type LanguageItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level string `json:"level"`
}

type Award struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Awarder string `json:"awarder"`
	Date    string `json:"date"`
	Summary string `json:"summary"`
}

type VolunteerExperience struct {
	ID           string `json:"id"`
	Organization string `json:"organization"`
	Position     string `json:"position"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	Summary      string `json:"summary"`
}

type Project struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	URL         string   `json:"url"`
	Keywords    []string `json:"keywords"`
}

type Reference struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Interest struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
}

type ResumeData struct {
	Basics         Basics                 `json:"basics"`
	Work           []WorkExperience       `json:"work"`
	Education      []EducationItem        `json:"education"`
	Skills         []Skill                `json:"skills"`
	Certifications []Certification        `json:"certifications"`
	Languages      []LanguageItem         `json:"languages"`
	Awards         []Award                `json:"awards"`
	Volunteer      []VolunteerExperience  `json:"volunteer"`
	Projects       []Project              `json:"projects"`
	References     []Reference            `json:"references"`
	Interests      []Interest             `json:"interests"`
	Metadata       map[string]interface{} `json:"metadata"`
}

func (r *ResumeData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, r)
}

func (r ResumeData) Value() (driver.Value, error) {
	return json.Marshal(r)
}

type Resume struct {
	ID         string     `gorm:"primaryKey;column:id" json:"id"`
	UserID     string     `gorm:"uniqueIndex;not null;column:user_id" json:"user_id"`
	Title      string     `gorm:"not null;column:title" json:"title"`
	PDFURL     string     `gorm:"column:pdf_url" json:"pdf_url"`
	PublicURL  string     `gorm:"column:public_url" json:"public_url"`
	RxResumeID string     `gorm:"column:rxresume_id" json:"rxresume_id"`
	Data       ResumeData `gorm:"type:jsonb;column:data" json:"data"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

type ProfileInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Location string `json:"location"`
	Headline string `json:"headline"`
}

func (r *ResumeData) Sanitize() {
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

func (r *ResumeData) Validate() error {
	if len(r.Basics.Name) > 200 {
		return errors.New("el nombre excede la longitud máxima")
	}
	if len(r.Basics.Email) > 200 {
		return errors.New("el correo excede la longitud máxima")
	}
	return nil
}

func (p *ProfileInput) Sanitize() {
	p.Name = html.EscapeString(strings.TrimSpace(p.Name))
	p.Email = html.EscapeString(strings.TrimSpace(p.Email))
	p.Phone = html.EscapeString(strings.TrimSpace(p.Phone))
	p.Location = html.EscapeString(strings.TrimSpace(p.Location))
	p.Headline = html.EscapeString(strings.TrimSpace(p.Headline))
}

func (p *ProfileInput) Validate() error {
	if len(p.Name) > 200 {
		return errors.New("el nombre excede la longitud máxima")
	}
	return nil
}
