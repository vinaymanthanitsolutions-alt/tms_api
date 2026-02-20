package controllers

import (
	"backend/internal/config"
	"backend/internal/services"
	"backend/internal/utils"
	"backend/web/models"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)
//all check
func CreateSubTask(c *gin.Context) {
	var input models.SubTask

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, 400, "Invalid input")
		return
	}
	result, err := config.DB.Exec(`
		INSERT INTO sub_task_master
		(parent_task_id, sub_task_title, sub_task_description, assigned_to_employee_id, assigned_by_employee_id, sub_task_priority)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		input.TaskID,
		input.Title,
		input.Description,
		input.AssignedTo,
		input.AssignedBy,
		input.Priority,
	)

	if err != nil {
		utils.Failed(c, 500, err.Error())
		return
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		utils.Failed(c, 500, err.Error())
		return
	}

	subTaskCode := fmt.Sprintf("ST%d%d", input.TaskID, lastID)

	_, err = config.DB.Exec(`
		UPDATE sub_task_master
		SET sub_task_code = ?
		WHERE sub_task_id = ?
	`, subTaskCode, lastID)

	if err != nil {
		utils.Failed(c, 500, err.Error())
		return
	}

	utils.Success(c, "Sub task created successfully")
}

func UpdateSubTaskStatus(c *gin.Context) {
	idParam := c.Param("id")
	subTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		log.Println("UpdateSubTaskStatus invalid id:", err)
		utils.Failed(c, 400, "Invalid subtask ID")
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Println("UpdateSubTaskStatus bind error:", err)
		utils.Failed(c, 400, "Invalid status input")
		return
	}

	_, err = config.DB.Exec(
		`UPDATE sub_task_master 
	 SET sub_task_status = ? 
	 WHERE sub_task_id = ? AND deleted_at IS NULL`,
		input.Status,
		subTaskID,
	)
	if err != nil {
		log.Println("UpdateSubTaskStatus DB error:", err)
		utils.Failed(c, 500, "Database error")
		return
	}

	// rows, _ := res.RowsAffected()
	// if rows == 0 {
	// 	log.Println("SubTask not found :",)
	// 	utils.Failed(c, 404, "SubTask not found")
	// 	return
	// }

	if err := services.UpdateProgressFromSubTask(subTaskID); err != nil {
		log.Println("Progress update error:", err)
		utils.Failed(c, 500, "Progress recalculation failed")
		return
	}

	utils.Success(c, "SubTask updated successfully")
}

func DeleteSubTask(c *gin.Context) {
	idParam := c.Param("id")
	subTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		log.Println("DeleteSubTask invalid id:", err)
		utils.Failed(c, 400, "Invalid subtask ID")
		return
	}

	res, err := config.DB.Exec(
		`UPDATE sub_task_master 
	 SET deleted_at = NOW() 
	 WHERE sub_task_id = ? AND deleted_at IS NULL`,
		subTaskID,
	)

	if err != nil {
		log.Println("DeleteSubTask DB error:", err)
		utils.Failed(c, 500, "Database error")
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		utils.Failed(c, 404, "SubTask already deleted or not found")
		return
	}

	if err := services.UpdateProgressFromSubTask(subTaskID); err != nil {
		log.Println("DeleteSubTask progress error:", err)
		utils.Failed(c, 500, "Progress recalculation failed")
		return
	}

	utils.Success(c, "SubTask deleted successfully")
}

func GetSubTasksByTask(c *gin.Context) {
	taskIDParam := c.Param("task_id")
	taskID, err := strconv.Atoi(taskIDParam)
	if err != nil {
		log.Println("GetSubTasksByTask invalid id:", err)
		utils.Failed(c, 400, "Invalid task ID")
		return
	}

	rows, err := config.DB.Query(`
	SELECT 
		sub_task_id,
		parent_task_id,
		sub_task_title,
		sub_task_description,
		sub_task_status,
		assigned_to_employee_id,
		assigned_by_employee_id,
		sub_task_priority,
		created_at,
		updated_at
	FROM sub_task_master
	WHERE parent_task_id = ? AND deleted_at IS NULL`,
		taskID,
	)

	if err != nil {
		log.Println("GetSubTasksByTask DB error:", err)
		utils.Failed(c, 500, "Database error")
		return
	}
	defer rows.Close()

	var subTasks []models.SubTask

	for rows.Next() {
		var sub models.SubTask

		err := rows.Scan(
			&sub.ID,
			&sub.TaskID,
			&sub.Title,
			&sub.Description,
			&sub.Status,
			&sub.AssignedTo,
			&sub.AssignedBy,
			&sub.Priority,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)

		if err != nil {
			log.Println("GetSubTasksByTask scan error:", err)
			utils.Failed(c, 500, "Data processing error")
			return
		}

		subTasks = append(subTasks, sub)
	}

	utils.Success(c, subTasks)
}
