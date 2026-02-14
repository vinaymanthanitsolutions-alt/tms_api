package controllers

import (
	"math"
	"net/http"
	"strconv"
	"time"

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

	managerID := c.Query("emp_id")
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
	if status == "" {
		utils.Failed(c, http.StatusBadRequest, "status is required")
		return
	}

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

	where := "WHERE manager_id = ?"
	args := []interface{}{managerID}

	if status != "ALL" {
		where += " AND status = ?"
		args = append(args, status)
	}

	if search != "" {
		where += " AND (emp_id LIKE ? OR emp_name LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	countQuery := "SELECT COUNT(*) FROM employee " + where
	var total int
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to count employees")
		return
	}

	query := `
		SELECT emp_id,emp_name,email,phone,department,role,status
		FROM employee
		` + where + `
		LIMIT ? OFFSET ?`

	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
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

func UpdatePassword(c *gin.Context) {

	var input struct {
		Email       string `json:"email"`
		OTP         string `json:"otp"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	var dbOTP string
	var expiry time.Time

	err := config.DB.QueryRow(
		`SELECT otp, expire_at FROM employee WHERE email = ?`,
		input.Email,
	).Scan(&dbOTP, &expiry)

	if err != nil {
		utils.Failed(c, http.StatusNotFound, "Email not registered")
		return
	}

	if dbOTP != input.OTP {
		utils.Failed(c, http.StatusUnauthorized, "Invalid OTP")
		return
	}

	if time.Now().After(expiry) {
		utils.Failed(c, http.StatusUnauthorized, "OTP expired")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(input.NewPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	_, err = config.DB.Exec(
		`UPDATE employee 
		 SET emp_password = ?, otp = NULL, expire_at = NULL 
		 WHERE email = ?`,
		string(hashedPassword),
		input.Email,
	)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to update password")
		return
	}

	utils.Success(c, gin.H{
		"message": "Password updated successfully",
	})
}
