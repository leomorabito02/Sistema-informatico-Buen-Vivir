package client

import (
	"buenvivir/rxresume/internal/models"
	"fmt"
)

// MapToRxResumeDTO takes existing RxResume JSON data and merges our internal Resume data into it.
func MapToRxResumeDTO(existingData map[string]interface{}, resume *models.Resume) map[string]interface{} {
	mergeBasics(existingData, resume)
	mergeSummary(existingData, resume)
	mergeSections(existingData, resume)
	mergeMetadata(existingData, resume)
	return existingData
}

func mapWork(work []models.WorkExperience) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(work))
	for _, w := range work {
		items = append(items, map[string]interface{}{
			"id":          w.ID,
			"hidden":      false,
			"company":     w.Company,
			"position":    w.Position,
			"location":    "",
			"period":      fmt.Sprintf("%s - %s", w.StartDate, w.EndDate),
			"website":     map[string]interface{}{"url": "", "label": "", "inlineLink": false},
			"description": w.Description,
			"roles":       []interface{}{},
		})
	}
	return items
}

func mapEducation(edu []models.EducationItem) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(edu))
	for _, e := range edu {
		items = append(items, map[string]interface{}{
			"id":          e.ID,
			"hidden":      false,
			"school":      e.Institution,
			"degree":      e.Degree,
			"area":        e.Field,
			"grade":       e.Grade,
			"location":    "",
			"period":      fmt.Sprintf("%s - %s", e.StartDate, e.EndDate),
			"website":     map[string]interface{}{"url": "", "label": "", "inlineLink": false},
			"description": e.Status,
		})
	}
	return items
}

func mapSkills(skills []models.Skill) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(skills))
	for _, s := range skills {
		items = append(items, map[string]interface{}{
			"id":          s.ID,
			"hidden":      false,
			"name":        s.Name,
			"proficiency": s.Level,
			"level":       0,
			"icon":        "",
			"iconColor":   "",
			"keywords":    []string{},
		})
	}
	return items
}

func mapLanguages(langs []models.LanguageItem) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(langs))
	for _, l := range langs {
		items = append(items, map[string]interface{}{
			"id":       l.ID,
			"hidden":   false,
			"language": l.Name,
			"fluency":  l.Level,
			"level":    0,
		})
	}
	return items
}

func mapCertifications(certs []models.Certification) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(certs))
	for _, c := range certs {
		items = append(items, map[string]interface{}{
			"id":          c.ID,
			"hidden":      false,
			"title":       c.Name,
			"issuer":      c.Issuer,
			"date":        c.Date,
			"website":     map[string]interface{}{"url": "", "label": "", "inlineLink": false},
			"description": "",
		})
	}
	return items
}

func mapAwards(awards []models.Award) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(awards))
	for _, a := range awards {
		items = append(items, map[string]interface{}{
			"id":          a.ID,
			"hidden":      false,
			"title":       a.Title,
			"awarder":     a.Awarder,
			"date":        a.Date,
			"website":     map[string]interface{}{"url": "", "label": "", "inlineLink": false},
			"description": a.Summary,
		})
	}
	return items
}

func mapVolunteer(volunteers []models.VolunteerExperience) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(volunteers))
	for _, v := range volunteers {
		items = append(items, map[string]interface{}{
			"id":           v.ID,
			"hidden":       false,
			"organization": v.Organization,
			"location":     "",
			"period":       fmt.Sprintf("%s - %s", v.StartDate, v.EndDate),
			"website":      map[string]interface{}{"url": "", "label": "", "inlineLink": false},
			"description":  v.Summary,
		})
	}
	return items
}

func mapProjects(projects []models.Project) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(projects))
	for _, p := range projects {
		items = append(items, map[string]interface{}{
			"id":          p.ID,
			"hidden":      false,
			"name":        p.Name,
			"period":      fmt.Sprintf("%s - %s", p.StartDate, p.EndDate),
			"website":     map[string]interface{}{"url": "", "label": "", "inlineLink": false},
			"description": p.Description,
			"keywords":    p.Keywords,
		})
	}
	return items
}

func mapReferences(references []models.Reference) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(references))
	for _, r := range references {
		items = append(items, map[string]interface{}{
			"id":          r.ID,
			"hidden":      false,
			"name":        r.Name,
			"position":    "",
			"website":     map[string]interface{}{"url": "", "label": "", "inlineLink": false},
			"phone":       "",
			"description": r.Description,
		})
	}
	return items
}

func mapInterests(interests []models.Interest) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(interests))
	for _, i := range interests {
		items = append(items, map[string]interface{}{
			"id":        i.ID,
			"hidden":    false,
			"name":      i.Name,
			"icon":      "",
			"iconColor": "",
			"keywords":  i.Keywords,
		})
	}
	return items
}

func mergeBasics(data map[string]interface{}, resume *models.Resume) {
	basics, ok := data["basics"].(map[string]interface{})
	if !ok {
		return
	}
	basics["name"] = resume.Data.Basics.Name
	basics["email"] = resume.Data.Basics.Email
	basics["phone"] = resume.Data.Basics.Phone
	basics["headline"] = resume.Data.Basics.Headline
	basics["location"] = resume.Data.Basics.Location
	if resume.Data.Basics.Picture != "" {
		if picture, ok := data["picture"].(map[string]interface{}); ok {
			picture["url"] = resume.Data.Basics.Picture
			picture["hidden"] = false
		} else {
			data["picture"] = map[string]interface{}{
				"url":          resume.Data.Basics.Picture,
				"hidden":       false,
				"size":         120,
				"aspectRatio":  1,
				"borderRadius": 0,
				"borderWidth":  0,
				"rotation":     0,
				"fit":          "cover",
				"borderColor":  "#000000",
				"shadowColor":  "#000000",
				"shadowWidth":  0,
			}
		}
	}

	if len(resume.Data.Basics.Profiles) > 0 {
		profiles := make([]map[string]interface{}, 0, len(resume.Data.Basics.Profiles))
		for _, p := range resume.Data.Basics.Profiles {
			profiles = append(profiles, map[string]interface{}{
				"id":   p.ID,
				"icon": p.Network,
				"text": p.Username,
				"link": p.URL,
			})
		}
		basics["customFields"] = profiles
	}
}

func mergeSummary(data map[string]interface{}, resume *models.Resume) {
	if summary, ok := data["summary"].(map[string]interface{}); ok {
		summary["content"] = resume.Data.Basics.Summary
	}
}

func mergeMetadata(data map[string]interface{}, resume *models.Resume) {
	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok || resume.Data.Metadata == nil || resume.Data.Metadata["template"] == nil {
		return
	}
	if tmplStr, isStr := resume.Data.Metadata["template"].(string); isStr && tmplStr != "" {
		validTemplates := map[string]bool{
			"azurill": true, "bronzor": true, "chikorita": true, "ditgar": true,
			"ditto": true, "gengar": true, "glalie": true, "kakuna": true,
			"lapras": true, "leafish": true, "meowth": true, "onyx": true,
			"pikachu": true, "rhyhorn": true, "scizor": true,
		}
		if validTemplates[tmplStr] {
			metadata["template"] = tmplStr
		} else {
			metadata["template"] = "azurill"
		}
	}
}

func mergeSections(data map[string]interface{}, resume *models.Resume) {
	sections, ok := data["sections"].(map[string]interface{})
	if !ok {
		return
	}

	setSectionItems(sections, "experience", mapWork(resume.Data.Work))
	setSectionItems(sections, "education", mapEducation(resume.Data.Education))
	setSectionItems(sections, "skills", mapSkills(resume.Data.Skills))
	setSectionItems(sections, "languages", mapLanguages(resume.Data.Languages))
	setSectionItems(sections, "certifications", mapCertifications(resume.Data.Certifications))
	setSectionItems(sections, "awards", mapAwards(resume.Data.Awards))
	setSectionItems(sections, "volunteer", mapVolunteer(resume.Data.Volunteer))
	setSectionItems(sections, "projects", mapProjects(resume.Data.Projects))
	setSectionItems(sections, "references", mapReferences(resume.Data.References))
	setSectionItems(sections, "interests", mapInterests(resume.Data.Interests))
}

func setSectionItems(sections map[string]interface{}, name string, items []map[string]interface{}) {
	if sec, ok := sections[name].(map[string]interface{}); ok {
		sec["items"] = items
	}
}
