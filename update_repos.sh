#!/bin/bash
# Adding methods to interfaces
sed -i '/type EnrollmentRepository interface {/a \\tListByParticipants(ctx context.Context, participantIDs []string) ([]models.Enrollment, error)' education/internal/repositories/enrollment_repository.go
