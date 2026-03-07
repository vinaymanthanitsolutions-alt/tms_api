package models

var RoleHierarchy = map[string]int{
	"SUPER_ADMIN":     1,
	"ADMIN":           2,
	"PROJECT_MANAGER": 3,
	"TEAM_LEADER":     4,
	"DEVELOPER":       5,
	"TESTER":          6,
}