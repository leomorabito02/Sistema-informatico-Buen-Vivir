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

const documentStoragePrefix = "documents/"

func extractStorageObjectName(url string) string {
	idx := strings.Index(url, documentStoragePrefix)
	if idx == -1 {
		return ""
	}
	objectName := url[idx:]
	if qIdx := strings.Index(objectName, "?"); qIdx != -1 {
		objectName = objectName[:qIdx]
	}
	return objectName
}

func (s *participantService) scheduleStorageCleanup(ctx context.Context, objectName string) {
	if objectName == "" || s.cleanupRepo == nil {
		return
	}
	task := &models.StorageCleanupTask{
		ID:          "cln-" + utils.GenerateRandomString(12),
		ObjectName:  objectName,
		Status:      models.CleanupStatusPending,
		Attempts:    0,
		MaxAttempts: 5,
		NextRunAt:   time.Now(),
	}
	if err := s.cleanupRepo.Enqueue(ctx, task); err != nil {
		fmt.Printf("Warning: failed to enqueue storage cleanup task for %s: %v\n", objectName, err)
	}
}

func (s *participantService) rollbackUploadedFile(objectName string) {
	if objectName == "" || s.storage == nil {
		return
	}
	if delErr := s.storage.DeleteFile(context.Background(), objectName); delErr != nil {
		fmt.Printf("Warning: failed to rollback file %s from GCS: %v\n", objectName, delErr)
		s.scheduleStorageCleanup(context.Background(), objectName)
	}
}

func (s *participantService) deleteStorageFileIfExists(ctx context.Context, docURL string) {
	objectName := extractStorageObjectName(docURL)
	if objectName == "" || s.storage == nil {
		return
	}
	if err := s.storage.DeleteFile(ctx, objectName); err != nil {
		fmt.Printf("Warning: failed to delete file %s from GCS: %v\n", objectName, err)
		s.scheduleStorageCleanup(ctx, objectName)
	}
}

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
	objectName := fmt.Sprintf("%s%s/%d_%s", documentStoragePrefix, participantID, time.Now().Unix(), cleanName)
	
	docURL, err := s.storage.UploadFile(ctx, objectName, file, contentType)
	if err != nil {
		return nil, fmt.Errorf("error uploading document: %w", err)
	}

	// 3. Crear registro en BD
	doc := &models.ParticipantDocument{
		ID:            "pdoc-" + utils.GenerateRandomString(10),
		ParticipantID: participantID,
		Name:          docName,
		DocURL:        docURL,
		FileType:      contentType,
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		s.rollbackUploadedFile(objectName)
		return nil, fmt.Errorf("failed to save document in database: %w", err)
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

	var newObjectName string
	var oldDocURL string

	if file != nil {
		cleanName := filepath.Base(originalName)
		newObjectName = fmt.Sprintf("%s%s/%d_%s", documentStoragePrefix, participantID, time.Now().Unix(), cleanName)
		
		docURL, err := s.storage.UploadFile(ctx, newObjectName, file, contentType)
		if err != nil {
			return nil, fmt.Errorf("error uploading document: %w", err)
		}
		
		oldDocURL = doc.DocURL
		doc.DocURL = docURL
		doc.FileType = contentType
	}

	if err := s.docRepo.Update(ctx, doc); err != nil {
		s.rollbackUploadedFile(newObjectName)
		return nil, fmt.Errorf("failed to update document in database: %w", err)
	}

	// Éxito confirmado en BD: purgar archivo previo
	if oldDocURL != "" {
		s.deleteStorageFileIfExists(ctx, oldDocURL)
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

	// Eliminar primero de la base de datos
	if err := s.docRepo.Delete(ctx, docID); err != nil {
		return fmt.Errorf("failed to delete document from database: %w", err)
	}

	// Tras éxito en BD, purgar archivo en GCS
	s.deleteStorageFileIfExists(ctx, doc.DocURL)

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

	objectName := extractStorageObjectName(doc.DocURL)
	if objectName == "" {
		return nil, "", "", fmt.Errorf("invalid document url")
	}

	reader, err := s.storage.DownloadFile(ctx, objectName)
	if err != nil {
		return nil, "", "", err
	}

	return reader, doc.FileType, doc.Name, nil
}
