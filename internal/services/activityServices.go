package services

import (
	"backend/internal/config"
	"encoding/json"
)

func LogActivity(
	employeeID string,
	actionType string,
	entityType string,
	entityID string,
	oldData interface{},
	newData interface{},
	description string,
) {

	var oldJSON []byte
	var newJSON []byte

	if oldData != nil {
		oldJSON, _ = json.Marshal(oldData)
	}

	if newData != nil {
		newJSON, _ = json.Marshal(newData)
	}

	query := `
	INSERT INTO activity_log
	(employee_id, action_type, entity_type, entity_id, old_data, new_data, description)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	config.DB.Exec(
		query,
		employeeID,
		actionType,
		entityType,
		entityID,
		oldJSON,
		newJSON,
		description,
	)
}