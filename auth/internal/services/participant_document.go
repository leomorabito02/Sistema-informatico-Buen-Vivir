package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"auth/internal/middlewares"
	"auth/internal/models"
	"auth/internal/utils"
)

func (s *participantService) UploadDocument(ctx context.Context, participantID string, docName string, file io.Reader, contentType string, originalName string) (*models.ParticipantDocument, error) {
	// 1. Validar acceso
	participant, err := s.participantRepo.FindByID(ctx, participantID)
	if err != nil {
		return nil, err
	}
	if err := checkParticipantUpdateAccess(ctx, participant); err != nil {
		return nil, err
	}

	// 2. Subir a GCS
	cleanName := filepath.Base(originalName)
	objectName := fmt.Sprintf("documents/%s/%d_%s", participantID, time.Now().Unix(), cleanName)
	
	docURL, err := s.storage.UploadFile(ctx, objectName, file, contentType)
	if err != nil {
		return nil, fmt.Errorf("error uploading document: %w", err)
	}

	// 3. Crear registro
	doc := &models.ParticipantDocument{
		ID:            "pdoc-" + utils.GenerateRandomString(10),
		ParticipantID: participantID,
		Name:          docName,
		DocURL:        docURL,
		FileType:      contentType,
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *participantService) GetDocuments(ctx context.Context, participantID string) ([]models.ParticipantDocument, error) {
	participant, err := s.participantRepo.FindByID(ctx, participantID)
	if err != nil {
		return nil, err
	}
	
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_participant" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if participant.UserID != reqUserID {
			return nil, fmt.Errorf("forbidden: access denied to other participant documents")
		}
	}

	return s.docRepo.GetByParticipantID(ctx, participantID)
}

func (s *participantService) UpdateDocument(ctx context.Context, participantID string, docID string, newName string, file io.Reader, contentType string, originalName string) (*models.ParticipantDocument, error) {
	participant, err := s.participantRepo.FindByID(ctx, participantID)
	if err != nil {
		return nil, err
	}
	if err := checkParticipantUpdateAccess(ctx, participant); err != nil {
		return nil, err
	}

	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc.ParticipantID != participantID {
		return nil, fmt.Errorf("document does not belong to participant")
	}

	if newName != "" {
		doc.Name = newName
	}

	if file != nil {
		cleanName := filepath.Base(originalName)
		objectName := fmt.Sprintf("documents/%s/%d_%s", participantID, time.Now().Unix(), cleanName)
		
		docURL, err := s.storage.UploadFile(ctx, objectName, file, contentType)
		if err != nil {
			return nil, fmt.Errorf("error uploading document: %w", err)
		}
		
		// Attempt to delete old file, but don't fail if it doesn't work perfectly right now
		// (e.g. if URL parsing is needed to get old objectName)
		
		doc.DocURL = docURL
		doc.FileType = contentType
	}

	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *participantService) DeleteDocument(ctx context.Context, participantID string, docID string) error {
	participant, err := s.participantRepo.FindByID(ctx, participantID)
	if err != nil {
		return err
	}
	if err := checkParticipantUpdateAccess(ctx, participant); err != nil {
		return err
	}

	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.ParticipantID != participantID {
		return fmt.Errorf("document does not belong to participant")
	}

	// Try to delete from GCS
	idx := strings.Index(doc.DocURL, "documents/")
	if idx != -1 {
		objectName := doc.DocURL[idx:]
		err = s.storage.DeleteFile(ctx, objectName)
		if err != nil {
			fmt.Printf("Warning: failed to delete file %s from GCS: %v\n", objectName, err)
		}
	}

	// Delete from DB
	if err := s.docRepo.Delete(ctx, docID); err != nil {
		return err
	}

	return nil
}

func (s *participantService) DownloadDocument(ctx context.Context, participantID string, docID string) (io.ReadCloser, string, string, error) {
	participant, err := s.participantRepo.FindByID(ctx, participantID)
	if err != nil {
		return nil, "", "", err
	}
	if err := checkParticipantUpdateAccess(ctx, participant); err != nil {
		return nil, "", "", err
	}

	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, "", "", err
	}
	if doc.ParticipantID != participantID {
		return nil, "", "", fmt.Errorf("document does not belong to participant")
	}

	idx := strings.Index(doc.DocURL, "documents/")
	if idx == -1 {
		return nil, "", "", fmt.Errorf("invalid document url")
	}
	objectName := doc.DocURL[idx:]

	reader, err := s.storage.DownloadFile(ctx, objectName)
	if err != nil {
		return nil, "", "", err
	}

	return reader, doc.FileType, doc.Name, nil
}
