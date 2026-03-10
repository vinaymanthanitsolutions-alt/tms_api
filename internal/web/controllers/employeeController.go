package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/utils"
	"backend/internal/web/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// all done
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
		INSERT INTO employee_master (
			employee_id,
			employee_name,
			employee_email,
			employee_phone,
			employee_password,
			employee_department,
			employee_role,
			manager_employee_id
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
		c.Error(err)
		return
	}

	utils.Success(c, "Employee created successfully")
}

func ShowEmployees(c *gin.Context) {

	// GET /emp?emp_id=SA001&role=SUPER_ADMIN&status=ALL&filter_role=ADMIN&page=1&limit=10

	// managerID := c.Get("emp_id")
	managerID := strings.TrimSpace(c.Query("emp_id"))
	// currentUserRole := c.Get("role")
	currentUserRole := strings.TrimSpace(c.Query("role"))
	filterRole := strings.TrimSpace(c.Query("filter_role"))

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
		conditions = append(conditions, "e.manager_employee_id = ?")
		args = append(args, managerID)
	}

	if status != "" && status != "ALL" {
		conditions = append(conditions, "e.employee_status = ?")
		args = append(args, status)
	}

	if filterRole != "" && filterRole != "ALL" {

		if currentUserRole != "SUPER_ADMIN" && filterRole == "SUPER_ADMIN" {
			utils.Failed(c, http.StatusForbidden, "Not allowed to view SUPER_ADMIN")
			return
		}
		conditions = append(conditions, "e.employee_role = ?")
		args = append(args, filterRole)
	}

	if search != "" {
		conditions = append(conditions, "(e.employee_id LIKE ? OR e.employee_name LIKE ?)")
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM employee_master e" + where

	var total int
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.Error(err)
		return
	}

	query := `
		SELECT 
			e.employee_id,
			e.employee_name,
			e.employee_email,
			e.employee_phone,
			e.employee_department,
			e.employee_role,
			e.employee_status,
			e.manager_employee_id,
			m.employee_name AS manager_name
		FROM employee_master e
		LEFT JOIN employee_master m 
			ON e.manager_employee_id = m.employee_id
	` + where + `
		LIMIT ? OFFSET ?`

	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := config.DB.Query(query, queryArgs...)
	if err != nil {
		c.Error(err)
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
			c.Error(err)
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

	query := `UPDATE employee_master SET deleted_at = NOW(), employee_status = 'SUSPENDED' WHERE employee_id=?`
	result, err := config.DB.Exec(query, empId)
	if err != nil {
		c.Error(err)
		return
	}
	rowAffected, err := result.RowsAffected()
	if err != nil {
		c.Error(err)
		return
	} else if rowAffected == 0 {
		utils.Failed(c, http.StatusNotFound, "Nothing Updated")
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
	UPDATE employee_master 
	SET deleted_at = NULL,
	    employee_status = 'ACTIVE'
	WHERE employee_id = ?
	  AND deleted_at IS NOT NULL
	`

	result, err := config.DB.Exec(query, empId)
	if err != nil {
		c.Error(err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.Error(err)
		return
	} else if rowsAffected == 0 {
		utils.Failed(c, http.StatusNotFound, "User already active ")
		return
	}

	utils.Success(c, "User restored successfully")
}

func UpdateProfile(c *gin.Context) {

	empID := c.Param("emp_id")
	if empID == "" {
		utils.Failed(c, http.StatusBadRequest, "emp_id is required")
		return
	}

	var data models.UserUpdate
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	var exists string
	err := config.DB.QueryRow(
		"SELECT employee_id FROM employee_master WHERE employee_id = ?",
		empID,
	).Scan(&exists)

	if err != nil {
		utils.Failed(c, http.StatusNotFound, "Employee not found")
		return
	}

	query := "UPDATE employee_master SET "
	var updates []string
	var args []interface{}

	updates = append(updates,
		"employee_name = ?",
		"employee_email = ?",
		"employee_phone = ?",
		"employee_department = ?",
		"employee_role = ?",
		"manager_employee_id = ?",
		"employee_status = ?",
	)

	args = append(args,
		data.EmpName,
		data.Email,
		data.Phone,
		data.Department,
		data.Role,
		data.ManagerID,
		data.Status,
	)

	if data.Status == "SUSPENDED" {
		updates = append(updates, "deleted_at = ?")
		args = append(args, time.Now())
	} else {
		updates = append(updates, "deleted_at = NULL")
	}

	if data.Password != "" {
		hash, err := bcrypt.GenerateFromPassword(
			[]byte(data.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			utils.Failed(c, http.StatusInternalServerError, "Error hashing password")
			return
		}

		updates = append(updates, "employee_password = ?")
		args = append(args, string(hash))
	}

	query += strings.Join(updates, ", ") + " WHERE employee_id = ?"
	args = append(args, empID)

	_, err = config.DB.Exec(query, args...)
	if err != nil {
		utils.LogError(err)
		c.Error(err)
		return
	}

	utils.Success(c, "Profile updated successfully")
}

func GetEmployeesUnderSameManager(c *gin.Context) {

	empID := c.Query("emp_id")
	// empID, _ := c.Get("emp_id")
	userRole := c.Query("role")
	// userRole,_ := c.Get("role")
	filterRole := c.Query("filter_role")

	if empID == "" || userRole == "" {
		utils.Failed(c, http.StatusBadRequest, "emp_id and role are required")
		return
	}

	baseQuery := `
		SELECT 
			e.employee_id,
			e.employee_name,
			e.employee_role
		FROM employee_master e
		WHERE 1=1
	`

	var args []interface{}

	switch userRole {

	case "SUPER_ADMIN":

	case "ADMIN":
		baseQuery += " AND e.manager_employee_id = ?"
		args = append(args, empID)

	case "PM":
		baseQuery += " AND e.pm_employee_id = ?"
		args = append(args, empID)

	default:
		utils.Failed(c, http.StatusForbidden, "invalid role")
		return
	}

	if filterRole != "" && filterRole != "ALL" {
		baseQuery += " AND e.employee_role = ?"
		args = append(args, filterRole)
	}

	rows, err := config.DB.Query(baseQuery, args...)
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

	if err = rows.Err(); err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, employees)
}

func GetEmployeesUnderSameManagert(c *gin.Context) {

	empID := c.Query("emp_id")
	// empID,  _ := c.Get("emp_id")
	filterRole := c.Query("filter_role")

	if empID == "" {
		utils.Failed(c, http.StatusBadRequest, "emp_id is required")
		return
	}

	baseQuery := `
		SELECT 
			e.employee_id,
			e.employee_name,
			e.employee_role
		FROM employee_master e
		WHERE e.pm_employee_id = ?
	`

	var rows *sql.Rows
	var err error

	if filterRole != "" && filterRole != "ALL" {
		query := baseQuery + " AND e.employee_role = ?"
		rows, err = config.DB.Query(query, empID, filterRole)
	} else {
		rows, err = config.DB.Query(baseQuery, empID)
	}

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

	utils.Success(c, employees)
}


func GetLatestEmployee(c *gin.Context) {

	role := c.Query("role")
	if role == "" {
		utils.Failed(c, http.StatusBadRequest, "role is required")
		return
	}

	var condition string

	switch role {
	case "SUPER_ADMIN":
		condition = "e.employee_role = ?"
	case "ADMIN":
		condition = "e.employee_role NOT IN (?,?)"
	default:
		utils.Failed(c, http.StatusBadRequest, "invalid role")
		return
	}

	query := `
	SELECT
		e.employee_id,
		e.employee_name,
		e.employee_role,
		e.employee_department,
		e.created_at,
		e.employee_status
	FROM employee_master e
	WHERE ` + condition + `
	ORDER BY e.created_at DESC
	LIMIT 3
	`

	var rows *sql.Rows
	var err error

	if role == "SUPER_ADMIN" {
		rows, err = config.DB.Query(query, "ADMIN")
	} else {
		rows, err = config.DB.Query(query, "ADMIN", "SUPER_ADMIN")
	}

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	var employees []models.LatestEmployee

	for rows.Next() {

		var emp models.LatestEmployee

		err := rows.Scan(
			&emp.EmpID,
			&emp.EmpName,
			&emp.EmpRole,
			&emp.EmpDepartment,
			&emp.CreatedAt,
			&emp.EmpStatus,
		)

		if err != nil {
			c.Error(err)
			continue
		}

		employees = append(employees, emp)
	}

	if err = rows.Err(); err != nil {
		utils.Failed(c, http.StatusInternalServerError, "row iteration error")
		return
	}

	utils.Success(c, employees)
}