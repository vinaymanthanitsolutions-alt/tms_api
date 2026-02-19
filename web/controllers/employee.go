package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/config"
	"backend/internal/utils"
	"backend/web/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Add(c *gin.Context) {
	var data models.UserGet

	if err := c.ShouldBindJSON(&data); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(data.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Error generating password")
		return
	}

	query := `
		INSERT INTO employee (
			emp_id,
			emp_name,
			email,
			phone,
			emp_password,
			department,
			role,
			manager_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = config.DB.Exec(
		query,
		data.EmpID,
		data.EmpName,
		data.Email,
		data.Phone,
		string(hashedPassword),
		data.Department,
		data.Role,
		data.ManagerID,
	)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Employee created successfully",
	})
}

func ShowEmployees(c *gin.Context) {

	// GET /emp?emp_id=SA001&role=SUPER_ADMIN&status=ALL&filter_role=ADMIN&page=1&limit=10

	managerID := strings.TrimSpace(c.Query("emp_id"))   
	currentUserRole := strings.TrimSpace(c.Query("currentUserRole")) 
	filterRole := strings.TrimSpace(c.Query("role")) 

	if currentUserRole != "SUPER_ADMIN" && managerID == "" {
		utils.Failed(c, http.StatusBadRequest, "emp_id is required")
		return
	}

	status := strings.TrimSpace(c.Query("status"))
	search := strings.TrimSpace(c.Query("search"))

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	conditions := []string{}
	args := []interface{}{}

	
	if currentUserRole != "SUPER_ADMIN" {
		conditions = append(conditions, "e.manager_id = ?")
		args = append(args, managerID)
	}

	if status != "" && status != "ALL" {
		conditions = append(conditions, "e.status = ?")
		args = append(args, status)
	}

	if filterRole != "" && filterRole != "ALL" {
	
		if currentUserRole != "SUPER_ADMIN" && filterRole == "SUPER_ADMIN" {
			utils.Failed(c, http.StatusForbidden, "Not allowed to view SUPER_ADMIN")
			return
		}
		conditions = append(conditions, "e.role = ?")
		args = append(args, filterRole)
	}


	if search != "" {
		conditions = append(conditions, "(e.emp_id LIKE ? OR e.emp_name LIKE ?)")
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}


	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	
	countQuery := "SELECT COUNT(*) FROM employee e" + where

	var total int
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to count employees")
		return
	}

	query := `
		SELECT 
			e.emp_id,
			e.emp_name,
			e.email,
			e.phone,
			e.department,
			e.role,
			e.status,
			e.manager_id,
			m.emp_name AS manager_name
		FROM employee e
		LEFT JOIN employee m 
			ON e.manager_id = m.emp_id
	` + where + `
		LIMIT ? OFFSET ?`

	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := config.DB.Query(query, queryArgs...)
	if err != nil {
		log.Println("Failed to fetch employees:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch employees")
		return
	}
	defer rows.Close()

	var employees []models.UserShow

	for rows.Next() {
		var emp models.UserShow

		if err := rows.Scan(
			&emp.EmpID,
			&emp.EmpName,
			&emp.Email,
			&emp.Phone,
			&emp.Department,
			&emp.Role,
			&emp.Status,
			&emp.ManagerID,
			&emp.ManagerName,
		); err != nil {
			log.Println("Scan error:", err)
			utils.Failed(c, http.StatusInternalServerError, "Error scanning employees")
			return
		}

		
		if emp.Role == "SUPER_ADMIN" {
			continue
		}

		employees = append(employees, emp)
	}

	utils.Success(c, gin.H{
		"page":  page,
		"limit": limit,
		"total": total,
		"data":  employees,
	})
}

func DeleteUser(c *gin.Context) {
	empId := c.Param("emp_id")
	if empId == "" {
		utils.Failed(c, http.StatusBadRequest, "emp_id is required")
		return
	}

	query := `UPDATE employee SET deleted_at = NOW(),status = 'SUSPENDED' WHERE emp_id=?`
	result, err := config.DB.Exec(query, empId)
	if err != nil {
		utils.Failed(c, 501, "Database executing query error ")
		return
	}
	rowAffected, err := result.RowsAffected()
	if err != nil {
		utils.Failed(c, http.StatusNotFound, "Unable to verify deletion")
		return
	} else if rowAffected == 0 {
		utils.Failed(c, http.StatusNotFound, "Userid is not in database")
		return
	}
	utils.Success(c, "User data deleted successfully")

}

func RestoreUser(c *gin.Context) {
	empId := c.Param("emp_id")
	if empId == "" {
		utils.Failed(c, http.StatusBadRequest, "emp_id is required")
		return
	}

	query := `
	UPDATE employee 
	SET deleted_at = NULL,
	    status = 'ACTIVE'
	WHERE emp_id = ?
	  AND deleted_at IS NOT NULL
	`

	result, err := config.DB.Exec(query, empId)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Database error")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.Failed(c, http.StatusNotFound, "User not found or already active")
		return
	}

	utils.Success(c, "User restored successfully")
}

func UpdateProfile(c *gin.Context) {

	empID := c.Param("emp_id")

	var data models.UserUpdate

	if err := c.ShouldBindJSON(&data); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	var exists string
	err := config.DB.QueryRow(
		"SELECT emp_id FROM employee WHERE emp_id = ?",
		empID,
	).Scan(&exists)

	if err != nil {
		utils.Failed(c, http.StatusNotFound, "Employee not found")
		return
	}

	var hashedPassword string
	if data.Password != "" {
		hash, err := bcrypt.GenerateFromPassword(
			[]byte(data.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			utils.Failed(c, http.StatusInternalServerError, "Error hashing password")
			return
		}
		hashedPassword = string(hash)
	}

	if data.Password != "" {

		// Update including password
		_, err = config.DB.Exec(`
			UPDATE employee SET
				emp_name = ?,
				email = ?,
				phone = ?,
				department = ?,
				role = ?,
				manager_id = ?,
				emp_password = ?,
				status =?
			WHERE emp_id = ?
		`,
			data.EmpName,
			data.Email,
			data.Phone,
			data.Department,
			data.Role,
			data.ManagerID,
			hashedPassword,
			data.Status,
			empID,
		)

	} else {

		// Update without password
		_, err = config.DB.Exec(`
			UPDATE employee SET
				emp_name = ?,
				email = ?,
				phone = ?,
				department = ?,
				role = ?,
				manager_id = ?,
				status =?
			WHERE emp_id = ?
		`,
			data.EmpName,
			data.Email,
			data.Phone,
			data.Department,
			data.Role,
			data.ManagerID,
			data.Status,
			empID,
		)
	}

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to update employee")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

func GetEmployeesUnderSameManager(c *gin.Context) {

	empID := c.Query("emp_id")
	if empID == "" {
		utils.Failed(c, http.StatusBadRequest, "emp_id is required")
		return
	}

	query := `
		SELECT 
			e.emp_id,
			e.emp_name,
			e.role
		FROM employee e
		WHERE e.manager_id = (
			SELECT manager_id 
			FROM employee 
			WHERE emp_id = ?
		)
	`

	rows, err := config.DB.Query(query, empID)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var employees []models.EmployeeUnderManager

	for rows.Next() {
		var emp models.EmployeeUnderManager
		if err := rows.Scan(
			&emp.EmpID,
			&emp.EmpName,
			&emp.Role,
		); err != nil {
			utils.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		employees = append(employees, emp)
	}

	if len(employees) == 0 {
		utils.Success(c, []models.EmployeeUnderManager{})
		return
	}

	utils.Success(c, employees)
}
