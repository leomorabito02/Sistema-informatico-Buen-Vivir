#!/bin/bash
sed -i '/type CourseRepository interface {/a \\tListByResponsibleTeachers(ctx context.Context, teacherIDs []string) ([]models.Course, error)' education/internal/repositories/course_repository.go
