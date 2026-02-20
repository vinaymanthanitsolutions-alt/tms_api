package services

import (
	"backend/internal/config"
	"database/sql"
	"fmt"
)

func UpdateProgressFromSubTask(subTaskID int) error {
	taskID, err := GetTaskIDBySubTaskID(subTaskID)
	if err != nil {
		return err
	}

	projectID, err := RefreshTaskProgress(taskID)
	if err != nil {
		return err
	}

	return RefreshProjectProgress(projectID)
}

func GetTaskIDBySubTaskID(subTaskID int) (int, error) {
	var taskID int

	err := config.DB.QueryRow(
		`SELECT parent_task_id 
		 FROM sub_task_master 
		 WHERE sub_task_id = ?`,
		subTaskID,
	).Scan(&taskID)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("subtask not found")
		}
		return 0, err
	}

	return taskID, nil
}

func RefreshTaskProgress(taskID int) (string, error) {
	var projectID string

	err := config.DB.QueryRow(
		`SELECT project_id 
		 FROM task_master 
		 WHERE task_id = ?`,
		taskID,
	).Scan(&projectID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("task not found")
		}
		return "", err
	}

	var total, completed int

	err = config.DB.QueryRow(`
		SELECT 
			COUNT(*),
			IFNULL(SUM(sub_task_status = 'COMPLETED'),0)
		FROM sub_task_master
		WHERE parent_task_id = ?
	`, taskID).Scan(&total, &completed)

	if err != nil {
		return "", err
	}

	var progress int
	var taskStatus string

	if total == 0 {
		progress = 0
		taskStatus = "TODO"
	} else {
		progress = (completed * 100) / total

		if completed == 0 {
			taskStatus = "TODO"
		} else if completed < total {
			taskStatus = "IN_PROGRESS"
		} else {
			taskStatus = "COMPLETED"
		}
	}

	_, err = config.DB.Exec(`
		UPDATE task_master
		SET task_progress = ?, task_status = ?
		WHERE task_id = ?
	`, progress, taskStatus, taskID)

	if err != nil {
		return "", err
	}

	return projectID, nil
}

func RefreshProjectProgress(projectID string) error {
	var progress int

	err := config.DB.QueryRow(`
		SELECT IFNULL(CAST(AVG(task_progress) AS UNSIGNED),0)
		FROM task_master
		WHERE project_id = ?
	`, projectID).Scan(&progress)

	if err != nil {
		return err
	}

	var status string

	if progress == 0 {
		status = "NOT_STARTED"
	} else if progress < 100 {
		status = "ACTIVE"
	} else {
		status = "COMPLETED"
	}

	_, err = config.DB.Exec(`
		UPDATE project_master
		SET project_progress = ?, project_status = ?
		WHERE project_id = ?
	`, progress, status, projectID)

	return err
}

// EXAMPLE TO USE WHENEVER SUBTASK STATUS CHANGED
// err := UpdateProgressFromSubTask(subTaskID)
// if err != nil {
// 	log.Println("Progress update failed:", err)
// }
