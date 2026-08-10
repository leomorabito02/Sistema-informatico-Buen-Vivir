package main

import (
	"log/slog"
	"os"

	"auth/config"
	"auth/internal/models"
)

func main() {
	config.InitLogger()
	slog.Info("Iniciando seeder de roles y permisos...")

	db, err := config.InitDatabase()
	if err != nil {
		slog.Error("Error al conectar a la base de datos", "error", err)
		os.Exit(1)
	}

	roles := []models.Role{
		{RoleID: "role_admin", Name: "admin", Description: "Administrador total del sistema"},
		{RoleID: "role_dev", Name: "dev", Description: "Desarrollador y soporte técnico"},
		{RoleID: "role_company", Name: "company", Description: "Empresa madrina/aliada"},
		{RoleID: "role_volunteer", Name: "volunteer", Description: "Voluntario de la fundación"},
		{RoleID: "role_participant", Name: "participant", Description: "Participante o educando"},
	}

	for _, role := range roles {
		if err := db.Where(&models.Role{Name: role.Name}).FirstOrCreate(&role).Error; err != nil {
			slog.Error("Error al crear el rol", "role", role.Name, "error", err)
		}
	}

	// Ejemplo de permisos básicos (puedes extender esto según tus necesidades)
	permissions := []models.Permission{
		{PermissionID: "perm_users_read", Resource: "users", Action: "read", Description: "Ver usuarios"},
		{PermissionID: "perm_users_write", Resource: "users", Action: "write", Description: "Crear/Editar usuarios"},
	}

	for _, perm := range permissions {
		if err := db.Where(models.Permission{Resource: perm.Resource, Action: perm.Action}).FirstOrCreate(&perm).Error; err != nil {
			slog.Error("Error al crear el permiso", "permission", perm.Action+" "+perm.Resource, "error", err)
		}
	}

	// Ejemplo: Asignar todos los permisos al rol 'admin'
	var adminRole models.Role
	var allPerms []models.Permission
	db.Where("name = ?", "admin").First(&adminRole)
	db.Find(&allPerms)

	for _, perm := range allPerms {
		rp := models.RolePermission{RoleID: adminRole.RoleID, PermissionID: perm.PermissionID}
		db.FirstOrCreate(&rp, rp)
	}

	slog.Info("Seeding completado con éxito.")
}
