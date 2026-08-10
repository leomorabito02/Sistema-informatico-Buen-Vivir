package main

import (
	"log/slog"
	"os"

	"auth/config"
	"auth/internal/models"
)

func main() {
	config.InitLogger()
	slog.Info("Iniciando migración de base de datos...")

	db, err := config.InitDatabase()
	if err != nil {
		slog.Error("Error al conectar a la base de datos", "error", err)
		os.Exit(1)
	}

	err = db.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.User{},
		&models.Company{},
		&models.Participant{},
		&models.Teacher{},
		&models.Volunteer{},
		&models.PasswordRecovery{},
		&models.RefreshToken{},
	)
	
	if err != nil {
		slog.Error("Error durante la auto-migración", "error", err)
		os.Exit(1)
	}

	slog.Info("¡Migración completada con éxito!")
}
