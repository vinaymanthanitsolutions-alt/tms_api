package controllers

import (
	"backend/internal/config"
	"backend/internal/services"
	"backend/internal/utils"
	"backend/web/models"
	"database/sql"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateSubTask(c *gin.Context) {
	var input models.SubTask

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Println("CreateSubTask bind error:", err)
		utils.Failed(c, 400, "Invalid input")
		return
	}

	if input.Title == "" {
		utils.Failed(c, 400, "title are required")
		return
	}

	_, err := config.DB.Exec(`
		INSERT INTO sub_tasks 
		(title,task_id, description, assigned_to, assigned_by, priority)
		VALUES (?, ?, ?, ?, ?, ?)`,
		input.Title,
		input.TaskID,
		input.Description,
		input.AssignedTo,
		input.AssignedBy,
		input.Priority,
	)

	if err != nil {
		log.Println("CreateSubTask DB error:", err)
		utils.Failed(c, 500, "Database error")
		return
	}

	utils.Success(c, "SubTask created successfully")
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
		`UPDATE sub_tasks SET status = ? WHERE id = ? AND deleted_at IS NULL`,
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
		`UPDATE sub_tasks SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`,
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
		SELECT id, task_id, title, description, status, 
		       assigned_to, assigned_by, priority, 
		       created_at, updated_at
		FROM sub_tasks
		WHERE task_id = ? AND deleted_at IS NULL`,
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

func GetTeamMembersWithSubTasks(c *gin.Context) {
	teamID := c.Query("team_id")
	roleFilter := c.Query("role")
	if teamID == "" {
		utils.Failed(c, 400, "team_id is required")
		return
	}

	query := `
		SELECT 
			e.emp_id,
			e.emp_name,
			e.role,
			e.department,
			st.id,
			st.title,
			st.status,
			st.priority,
			st.task_id
		FROM team_members tm
		JOIN employee e 
			ON tm.employee_id = e.emp_id
		LEFT JOIN sub_tasks st 
			ON st.assigned_to = e.emp_id 
			AND st.deleted_at IS NULL
		WHERE tm.team_id = ?
	`

	args := []interface{}{teamID}

	if roleFilter != "" && roleFilter != "ALL" {
		query += " AND e.role = ?"
		args = append(args, roleFilter)
	}

	query += " ORDER BY e.emp_name"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		log.Println("GetTeamMembersWithSubTasks DB error:", err)
		utils.Failed(c, 500, "Database error")
		return
	}
	defer rows.Close()

	memberMap := make(map[string]*models.Member)

	for rows.Next() {
		var empID, empName, role, department string
		var subID sql.NullInt64
		var title, status, priority sql.NullString
		var taskID sql.NullInt64

		err := rows.Scan(
			&empID,
			&empName,
			&role,
			&department,
			&subID,
			&title,
			&status,
			&priority,
			&taskID,
		)
		if err != nil {
			log.Println("Scan error:", err)
			utils.Failed(c, 500, "Data processing error")
			return
		}

		if _, exists := memberMap[empID]; !exists {
			memberMap[empID] = &models.Member{
				EmpID:      empID,
				EmpName:    empName,
				Role:       role,
				Department: department,
				SubTasks:   []models.GetSubTask{},
			}
		}

		if subID.Valid {
			memberMap[empID].SubTasks = append(memberMap[empID].SubTasks, models.GetSubTask{
				ID:       int(subID.Int64),
				Title:    title.String,
				Status:   status.String,
				Priority: priority.String,
				TaskID:   int(taskID.Int64),
			})
		}
	}

	var members []models.Member
	for _, m := range memberMap {
		members = append(members, *m)
	}

	utils.Success(c, gin.H{
		"team_id": teamID,
		"members": members,
	})
}
