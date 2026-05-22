package integration

import (
	"encoding/json"
	"fmt"
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

type DepartmentResponse struct {
	ID       uint                 `json:"id"`
	Name     string               `json:"name"`
	Children []DepartmentResponse `json:"children"`
}

func setupTreeTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.Department{}, &models.Employee{})
	require.NoError(t, err)
	return db
}

func buildTree(t *testing.T, db *gorm.DB, depth int) uint {
	if depth < 1 {
		return 0
	}
	var rootID uint
	root := models.Department{Name: "Level1"}
	err := db.Create(&root).Error
	require.NoError(t, err)
	rootID = root.ID

	currentParentID := rootID
	for level := 2; level <= depth; level++ {
		child := models.Department{
			Name:     "Level" + strconv.Itoa(level),
			ParentID: &currentParentID,
		}
		err := db.Create(&child).Error
		require.NoError(t, err)
		currentParentID = child.ID
	}
	return rootID
}

func TestGetDepartmentTreeDepth(t *testing.T) {
	db := setupTreeTestDB(t)

	rootID := buildTree(t, db, 5)

	deptRepo := repositories.NewDepartmentRepository(db)
	empRepo := repositories.NewEmployeeRepository(db)

	deptService := services.NewDepartmentService(deptRepo)
	empService := services.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handlers.NewDepartmentHandler(deptService)
	empHandler := handlers.NewEmployeeHandler(empService)

	router := server.NewRouter(deptHandler, empHandler)
	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("depth=2", func(t *testing.T) {
		url := ts.URL + "/departments/" + strconv.Itoa(int(rootID)) + "?depth=2&include_employees=false"
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var root DepartmentResponse
		err = json.NewDecoder(resp.Body).Decode(&root)
		require.NoError(t, err)

		t.Logf("Depth=2 response:\n%s", prettyPrint(root))

		assert.Equal(t, "Level1", root.Name)
		require.Len(t, root.Children, 1, "should have one child at depth 2")
		assert.Equal(t, "Level2", root.Children[0].Name)
		assert.Empty(t, root.Children[0].Children, "children of level2 should be empty when depth=2")
	})

	t.Run("depth=5", func(t *testing.T) {
		url := ts.URL + "/departments/" + strconv.Itoa(int(rootID)) + "?depth=5&include_employees=false"
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var root DepartmentResponse
		err = json.NewDecoder(resp.Body).Decode(&root)
		require.NoError(t, err)

		t.Logf("Depth=5 response:\n%s", prettyPrint(root))

		assert.Equal(t, "Level1", root.Name)
		require.Len(t, root.Children, 1)
		assert.Equal(t, "Level2", root.Children[0].Name)
		require.Len(t, root.Children[0].Children, 1)
		assert.Equal(t, "Level3", root.Children[0].Children[0].Name)
		require.Len(t, root.Children[0].Children[0].Children, 1)
		assert.Equal(t, "Level4", root.Children[0].Children[0].Children[0].Name)
		require.Len(t, root.Children[0].Children[0].Children[0].Children, 1)
		assert.Equal(t, "Level5", root.Children[0].Children[0].Children[0].Children[0].Name)
		assert.Empty(t, root.Children[0].Children[0].Children[0].Children[0].Children)
	})

	t.Run("depth=10", func(t *testing.T) {
		url := ts.URL + "/departments/" + strconv.Itoa(int(rootID)) + "?depth=10&include_employees=false"
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("default depth=1", func(t *testing.T) {
		url := ts.URL + "/departments/" + strconv.Itoa(int(rootID)) + "?include_employees=false"
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var root DepartmentResponse
		err = json.NewDecoder(resp.Body).Decode(&root)
		require.NoError(t, err)

		t.Logf("Default depth response:\n%s", prettyPrint(root))

		assert.Equal(t, "Level1", root.Name)
		assert.Empty(t, root.Children, "children should be empty when depth not specified (default 1)")
	})
}

func TestGetDepartmentTwoChildren(t *testing.T) {
	db := setupTreeTestDB(t)

	parent := models.Department{Name: "Parent"}
	err := db.Create(&parent).Error
	require.NoError(t, err)

	child1 := models.Department{Name: "Child1", ParentID: &parent.ID}
	child2 := models.Department{Name: "Child2", ParentID: &parent.ID}
	err = db.Create(&child1).Error
	require.NoError(t, err)
	err = db.Create(&child2).Error
	require.NoError(t, err)

	deptRepo := repositories.NewDepartmentRepository(db)
	empRepo := repositories.NewEmployeeRepository(db)

	deptService := services.NewDepartmentService(deptRepo)
	empService := services.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handlers.NewDepartmentHandler(deptService)
	empHandler := handlers.NewEmployeeHandler(empService)

	router := server.NewRouter(deptHandler, empHandler)
	ts := httptest.NewServer(router)
	defer ts.Close()

	url := ts.URL + "/departments/" + strconv.Itoa(int(parent.ID)) + "?depth=2&include_employees=false"
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	type DepartmentResponse struct {
		ID       uint                 `json:"id"`
		Name     string               `json:"name"`
		Children []DepartmentResponse `json:"children"`
	}
	var root DepartmentResponse
	err = json.NewDecoder(resp.Body).Decode(&root)
	require.NoError(t, err)

	assert.Equal(t, "Parent", root.Name)
	require.Len(t, root.Children, 2, "parent should have exactly 2 children")

	childNames := make(map[string]bool)
	for _, child := range root.Children {
		childNames[child.Name] = true
		assert.Empty(t, child.Children, "children should have no children of their own at depth=2")
	}
	assert.True(t, childNames["Child1"], "Child1 not found")
	assert.True(t, childNames["Child2"], "Child2 not found")

	prettyJSON, _ := json.MarshalIndent(root, "", "  ")
	t.Logf("Response structure (depth=2):\n%s", prettyJSON)
}

func TestGetDepartmentTwoChildrenWithGrandchildren(t *testing.T) {
	db := setupTreeTestDB(t)

	parent := models.Department{Name: "Parent"}
	err := db.Create(&parent).Error
	require.NoError(t, err)

	child1 := models.Department{Name: "Child1", ParentID: &parent.ID}
	child2 := models.Department{Name: "Child2", ParentID: &parent.ID}
	err = db.Create(&child1).Error
	require.NoError(t, err)
	err = db.Create(&child2).Error
	require.NoError(t, err)

	grandchild1_1 := models.Department{Name: "Grandchild1_1", ParentID: &child1.ID}
	grandchild1_2 := models.Department{Name: "Grandchild1_2", ParentID: &child1.ID}
	err = db.Create(&grandchild1_1).Error
	require.NoError(t, err)
	err = db.Create(&grandchild1_2).Error
	require.NoError(t, err)

	grandchild2_1 := models.Department{Name: "Grandchild2_1", ParentID: &child2.ID}
	grandchild2_2 := models.Department{Name: "Grandchild2_2", ParentID: &child2.ID}
	err = db.Create(&grandchild2_1).Error
	require.NoError(t, err)
	err = db.Create(&grandchild2_2).Error
	require.NoError(t, err)

	deptRepo := repositories.NewDepartmentRepository(db)
	empRepo := repositories.NewEmployeeRepository(db)

	deptService := services.NewDepartmentService(deptRepo)
	empService := services.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handlers.NewDepartmentHandler(deptService)
	empHandler := handlers.NewEmployeeHandler(empService)

	router := server.NewRouter(deptHandler, empHandler)
	ts := httptest.NewServer(router)
	defer ts.Close()

	url := ts.URL + "/departments/" + strconv.Itoa(int(parent.ID)) + "?depth=3&include_employees=false"
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var root DepartmentResponse
	err = json.NewDecoder(resp.Body).Decode(&root)
	require.NoError(t, err)

	assert.Equal(t, "Parent", root.Name)
	require.Len(t, root.Children, 2, "parent should have exactly 2 children")

	for _, child := range root.Children {
		assert.Contains(t, []string{"Child1", "Child2"}, child.Name)
		require.Len(t, child.Children, 2, "child %s should have 2 grandchildren", child.Name)
		for _, grand := range child.Children {
			assert.Contains(t, grand.Name, "Grandchild")
			assert.Empty(t, grand.Children, "grandchildren should have no children at depth=3")
		}
	}

	prettyJSON, _ := json.MarshalIndent(root, "", "  ")
	t.Logf("Response structure (depth=3):\n%s", prettyJSON)
}

func prettyPrint(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}
