package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"hitalent-test/internal/handlers"
	"hitalent-test/internal/models"
	"hitalent-test/internal/repositories"
	"hitalent-test/internal/server"
	"hitalent-test/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.Department{}, &models.Employee{})
	require.NoError(t, err)
	return db
}

func TestCreateDepartmentAndEmployee(t *testing.T) {
	db := setupTestDB(t)

	deptRepo := repositories.NewDepartmentRepository(db)
	empRepo := repositories.NewEmployeeRepository(db)

	deptService := services.NewDepartmentService(deptRepo)
	empService := services.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handlers.NewDepartmentHandler(deptService)
	empHandler := handlers.NewEmployeeHandler(empService)

	router := server.NewRouter(deptHandler, empHandler)
	ts := httptest.NewServer(router)
	defer ts.Close()

	createDeptBody := []byte(`{"name":"HR"}`)
	resp, err := http.Post(ts.URL+"/departments/", "application/json", bytes.NewReader(createDeptBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var dept models.Department
	err = json.NewDecoder(resp.Body).Decode(&dept)
	require.NoError(t, err)
	assert.Equal(t, "HR", dept.Name)

	createEmpBody := []byte(`{"full_name":"Alice Smith","position":"Recruiter"}`)
	empResp, err := http.Post(ts.URL+"/departments/"+strconv.Itoa(int(dept.ID))+"/employees",
		"application/json", bytes.NewReader(createEmpBody))
	require.NoError(t, err)
	defer empResp.Body.Close()

	bodyBytes, _ := io.ReadAll(empResp.Body)
	bodyStr := string(bodyBytes)

	assert.Equal(t, http.StatusCreated, empResp.StatusCode, "response body: %s", bodyStr)

	var emp models.Employee
	err = json.Unmarshal(bodyBytes, &emp)
	require.NoError(t, err)
	assert.Equal(t, "Alice Smith", emp.FullName)
	assert.Equal(t, "Recruiter", emp.Position)
	assert.Equal(t, dept.ID, emp.DepartmentID)

	getResp, err := http.Get(ts.URL + "/departments/" + strconv.Itoa(int(dept.ID)) + "?include_employees=true")
	require.NoError(t, err)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var result struct {
		models.Department
		Employees []models.Employee `json:"employees"`
	}
	err = json.NewDecoder(getResp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Len(t, result.Employees, 1)
	assert.Equal(t, emp.ID, result.Employees[0].ID)
}
