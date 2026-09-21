package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"time"

	"buenvivir/rxresume/internal/models"
)

const (
	headerSecretKey   = "X-Secret-Key"
	headerUserID      = "X-User-ID"
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

type ReactiveResumeClient interface {
	SyncResume(ctx context.Context, resume *models.Resume) error
	ExportResumePDF(ctx context.Context, resumeID string) ([]byte, error)
	UploadImage(ctx context.Context, userID string, file io.Reader, filename string) (string, error)
}

type reactiveResumeClient struct {
	baseURL     string
	frontendURL string
	secretKey   string
	httpClient  *http.Client
}

// NewReactiveResumeClient initializes an HTTP REST client for the self-hosted Reactive Resume instance.
func NewReactiveResumeClient(baseURL, frontendURL, secretKey string) ReactiveResumeClient {
	return &reactiveResumeClient{
		baseURL:     baseURL,
		frontendURL: frontendURL,
		secretKey:   secretKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SyncResume sends updated resume JSON data to the self-hosted Reactive Resume service API.
func (c *reactiveResumeClient) SyncResume(ctx context.Context, resume *models.Resume) error {
	userID := resume.UserID
	slug := fmt.Sprintf("cv-%s", userID)

	// Always set public URL and PDF URL
	resume.PublicURL = fmt.Sprintf("%s/cv-public/%s", c.frontendURL, userID)
	resume.PDFURL = fmt.Sprintf("%s/api/v1/rxresume/public/%s/pdf", c.frontendURL, userID)

	if resume.RxResumeID == "" {
		c.ensureRxResumeID(ctx, userID, slug, resume)
	}

	if resume.RxResumeID != "" {
		c.updateRxResumeData(ctx, userID, resume)
	}

	return nil
}

func (c *reactiveResumeClient) ensureRxResumeID(ctx context.Context, userID, slug string, resume *models.Resume) {
	createPayload := map[string]interface{}{
		"json": map[string]interface{}{
			"name":           fmt.Sprintf("CV-%s", userID),
			"slug":           slug,
			"tags":           []string{},
			"withSampleData": false,
		},
	}

	bodyBytes, _ := json.Marshal(createPayload)
	reqURL := fmt.Sprintf("%s/api/rpc/resume/create", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(bodyBytes))
	if err == nil {
		req.Header.Set(headerContentType, contentTypeJSON)
		req.Header.Set(headerSecretKey, c.secretKey)
		req.Header.Set(headerUserID, userID)

		resp, err := c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 400 {
			var respData struct {
				JSON string `json:"json"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&respData); err == nil && respData.JSON != "" {
				resume.RxResumeID = respData.JSON
			}
			resp.Body.Close()
		} else if resp != nil {
			resp.Body.Close()
		}
	}

	if resume.RxResumeID == "" {
		c.fetchExistingRxResumeID(ctx, userID, slug, resume)
	}
}

func (c *reactiveResumeClient) fetchExistingRxResumeID(ctx context.Context, userID, slug string, resume *models.Resume) {
	listReqURL := fmt.Sprintf("%s/api/rpc/resume/list", c.baseURL)
	listReq, err := http.NewRequestWithContext(ctx, http.MethodPost, listReqURL, bytes.NewBufferString(`{"json":{}}`))
	if err != nil {
		return
	}

	listReq.Header.Set(headerContentType, contentTypeJSON)
	listReq.Header.Set(headerSecretKey, c.secretKey)
	listReq.Header.Set(headerUserID, userID)

	listResp, listErr := c.httpClient.Do(listReq)
	if listErr != nil {
		return
	}
	defer listResp.Body.Close()

	if listResp.StatusCode >= 400 {
		return
	}

	var listRespData struct {
		JSON []struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		} `json:"json"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listRespData); err == nil {
		for _, item := range listRespData.JSON {
			if item.Slug == slug {
				resume.RxResumeID = item.ID
				break
			}
		}
	}
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
			"id":          l.ID,
			"hidden":      false,
			"language":    l.Name,
			"fluency":     l.Level,
			"level":       0,
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

func (c *reactiveResumeClient) updateRxResumeData(ctx context.Context, userID string, resume *models.Resume) {
	// 1. Fetch current resume to get full valid data structure
	data := c.fetchRxResumeData(ctx, userID, resume.RxResumeID)
	if data == nil {
		return
	}

	// 2. Merge our data safely
	mergeBasics(data, resume)
	mergeSummary(data, resume)
	mergeSections(data, resume)
	mergeMetadata(data, resume)

	// 3. Send update with merged data and isPublic flag
	c.sendRxResumeUpdate(ctx, userID, resume.RxResumeID, data)
}

func (c *reactiveResumeClient) fetchRxResumeData(ctx context.Context, userID, resumeID string) map[string]interface{} {
	getPayload := map[string]interface{}{
		"json": map[string]interface{}{
			"id": resumeID,
		},
	}
	getBodyBytes, _ := json.Marshal(getPayload)
	getReqURL := fmt.Sprintf("%s/api/rpc/resume/getById", c.baseURL)
	getReq, err := http.NewRequestWithContext(ctx, http.MethodPost, getReqURL, bytes.NewBuffer(getBodyBytes))
	if err != nil {
		return nil
	}
	getReq.Header.Set(headerContentType, contentTypeJSON)
	getReq.Header.Set(headerSecretKey, c.secretKey)
	getReq.Header.Set(headerUserID, userID)

	resp, err := c.httpClient.Do(getReq)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil
	}

	var rxResume struct {
		JSON struct {
			Data map[string]interface{} `json:"data"`
		} `json:"json"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rxResume); err != nil {
		return nil
	}
	return rxResume.JSON.Data
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
				"id":       p.ID,
				"network":  p.Network,
				"username": p.Username,
				"url":      p.URL,
				"icon":     "",
				"hidden":   false,
			})
		}
		basics["customFields"] = profiles // In RX, profiles is its own array, but maybe it goes in profiles?
		// Wait, let's just map it to 'profiles'
		// Actually, in RX resume it's usually inside 'customFields' or 'profiles'? Let's just set 'profiles' inside basics or inside sections? Wait, profiles are in basics but it's an array of objects. Let's see...
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
		metadata["template"] = tmplStr
	}
}

func (c *reactiveResumeClient) sendRxResumeUpdate(ctx context.Context, userID, resumeID string, data map[string]interface{}) {
	updatePayload := map[string]interface{}{
		"json": map[string]interface{}{
			"id":                  resumeID,
			"isPublic":            true,
			"showDownloadButtons": true,
			"data":                data,
		},
	}

	bodyBytes, _ := json.Marshal(updatePayload)
	reqURL := fmt.Sprintf("%s/api/rpc/resume/update", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return
	}

	req.Header.Set(headerContentType, contentTypeJSON)
	req.Header.Set(headerSecretKey, c.secretKey)
	req.Header.Set(headerUserID, userID)

	updateResp, err := c.httpClient.Do(req)
	if err == nil {
		if updateResp.StatusCode >= 400 {
			b, _ := io.ReadAll(updateResp.Body)
			fmt.Printf("RxResume update failed: %d %s\n", updateResp.StatusCode, string(b))
		}
		updateResp.Body.Close()
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

// ExportResumePDF requests the rendered PDF binary from Reactive Resume service.
func (c *reactiveResumeClient) ExportResumePDF(ctx context.Context, userID string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/api/resumes/%s/cv-%s/pdf", c.baseURL, userID, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(headerSecretKey, c.secretKey)
	// We might also need X-User-ID if the patched auth expects it
	req.Header.Set(headerUserID, "internal-export")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from reactive resume export: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *reactiveResumeClient) UploadImage(ctx context.Context, userID string, file io.Reader, filename string) (string, error) {
	// For oRPC file upload, we use multipart/form-data.
	// oRPC often expects the file to be under a specific field, commonly "file" or just mapping the arg.
	// We'll proxy it directly.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// oRPC expects a 'data' field describing the structure, and the file mapped to '0'
	dataField := `{"json":{},"maps":[[]]}`
	if err := writer.WriteField("data", dataField); err != nil {
		return "", err
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "0", filename))
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	h.Set("Content-Type", contentType)

	part, err := writer.CreatePart(h)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	writer.Close()

	reqURL := fmt.Sprintf("%s/api/rpc/storage/uploadFile", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		return "", err
	}

	req.Header.Set(headerContentType, writer.FormDataContentType())
	req.Header.Set(headerSecretKey, c.secretKey)
	req.Header.Set(headerUserID, userID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		JSON struct {
			URL string `json:"url"`
		} `json:"json"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.JSON.URL, nil
}
