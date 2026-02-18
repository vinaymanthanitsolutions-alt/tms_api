package controllers

import (
	"log"
	"math"
	"net/http"
	"strconv"

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

	// GET /emp?emp_id=SA001&status=ACTIVE&page=2&limit=5&search=ayu example api call

	managerID := c.Query("emp_id")
	role:= c.Query("role")
	if managerID == "" {

		utils.Failed(c, http.StatusBadRequest, "emp_id is required")
		return
	}

	// managerID, exists := c.Get("emp_id")
	// if !exists {
	// 	utils.Failed(c, http.StatusUnauthorized, "Unauthorized")
	// 	return
	// }

	status := c.Query("status")
	// if status == "" {
	// 	utils.Failed(c, http.StatusBadRequest, "status is required")
	// 	return
	// }

	search := c.Query("search")

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	where := "WHERE e.manager_id = ?"
	args := []interface{}{managerID}

	if status != "ALL" && status!= "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	if role != ""{
		where+=" AND role = ? "
		args=append(args,role)
	}

	if search != "" {
		where += " AND (emp_id LIKE ? OR emp_name LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	countQuery := "SELECT COUNT(*) FROM employee e " + where
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

	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		log.Println("Failed to fetch employees",err)
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
			utils.Failed(c, http.StatusInternalServerError, "Error scanning employees")
			return
		}
		employees = append(employees, emp)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    employees,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(limit))),
		},
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
		utils.Success(c,  []models.EmployeeUnderManager{})
		return
	}

	utils.Success(c,  employees)
}
