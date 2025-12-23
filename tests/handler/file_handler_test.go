package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/http/handler"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"github.com/straye-as/relation-api/internal/storage"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const testMaxUploadMB int64 = 50

func setupFileHandlerTestDB(t *testing.T) *gorm.DB {
	return testutil.SetupCleanTestDB(t)
}

func setupTestStorage(t *testing.T) storage.Storage {
	// Create a temp directory for test files
	tempDir, err := os.MkdirTemp("", "file_handler_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	localStorage, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	return localStorage
}

func createFileHandler(t *testing.T, db *gorm.DB, store storage.Storage) *handler.FileHandler {
	logger := zap.NewNop()
	fileRepo := repository.NewFileRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	fileService := service.NewFileService(
		fileRepo,
		offerRepo,
		customerRepo,
		projectRepo,
		supplierRepo,
		activityRepo,
		store,
		logger,
	)

	return handler.NewFileHandler(fileService, testMaxUploadMB, logger)
}

func createFileTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

// withFileChiContext adds Chi route context with the given URL parameters
func withFileChiContext(ctx context.Context, params map[string]string) context.Context {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}

// createTestProject creates a test project for file tests
func createTestProjectForFiles(t *testing.T, db *gorm.DB, customer *domain.Customer, name string) *domain.Project {
	startDate := time.Now()
	customerID := customer.ID
	project := &domain.Project{
		Name:         name,
		CustomerID:   &customerID,
		CustomerName: customer.Name,
		Phase:        domain.ProjectPhaseWorking,
		StartDate:    startDate,
	}
	err := db.Create(project).Error
	require.NoError(t, err)
	return project
}

// createTestOfferForFiles creates a test offer for file tests
func createTestOfferForFiles(t *testing.T, db *gorm.DB, customer *domain.Customer) *domain.Offer {
	offer := &domain.Offer{
		Title:        "Test Offer for Files",
		CustomerID:   &customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Phase:        domain.OfferPhaseDraft,
		Status:       domain.OfferStatusActive,
		Probability:  50,
		Value:        100000,
	}
	err := db.Create(offer).Error
	require.NoError(t, err)
	return offer
}

// createMultipartFormFile creates a multipart form request with a file
func createMultipartFormFile(t *testing.T, fieldName, filename, content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filename)
	require.NoError(t, err)

	_, err = part.Write([]byte(content))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	return body, writer.FormDataContentType()
}

// createMultipartFormFileWithCompany creates a multipart form request with a file and company_id
func createMultipartFormFileWithCompany(t *testing.T, fieldName, filename, content, companyID string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add company_id field first
	if companyID != "" {
		err := writer.WriteField("company_id", companyID)
		require.NoError(t, err)
	}

	part, err := writer.CreateFormFile(fieldName, filename)
	require.NoError(t, err)

	_, err = part.Write([]byte(content))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	return body, writer.FormDataContentType()
}

// ============================================================================
// Upload Tests
// ============================================================================

func TestFileHandler_UploadToCustomer(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")

	t.Run("upload file to customer successfully (inherits gruppen as default)", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "test-document.txt", "Hello, this is test content!")

		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		if rr.Code != http.StatusCreated {
			t.Logf("Response body: %s", rr.Body.String())
		}
		require.Equal(t, http.StatusCreated, rr.Code)

		var result domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "test-document.txt", result.Filename)
		require.NotNil(t, result.CustomerID)
		assert.Equal(t, customer.ID, *result.CustomerID)
		assert.Greater(t, result.Size, int64(0))
		// Company should default to gruppen for customers without a specific company
		assert.Equal(t, domain.CompanyGruppen, result.CompanyID)
	})

	t.Run("upload file to customer with explicit company_id", func(t *testing.T) {
		body, contentType := createMultipartFormFileWithCompany(t, "file", "test-stalbygg.txt", "Stalbygg content", "stalbygg")

		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, domain.CompanyStalbygg, result.CompanyID)
	})

	t.Run("upload with invalid company_id returns 400", func(t *testing.T) {
		body, contentType := createMultipartFormFileWithCompany(t, "file", "test.txt", "content", "invalid_company")

		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("upload to non-existent customer returns 404", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "test.txt", "content")
		fakeID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, "/customers/"+fakeID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("upload with invalid customer ID returns 400", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "test.txt", "content")

		req := httptest.NewRequest(http.MethodPost, "/customers/invalid-id/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": "invalid-id"}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("upload without file field returns 400", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("other_field", "some value")
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestFileHandler_UploadToProject(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	project := createTestProjectForFiles(t, db, customer, "Test Project")

	t.Run("upload file to project successfully with company_id", func(t *testing.T) {
		// Projects require company_id since they are cross-company
		body, contentType := createMultipartFormFileWithCompany(t, "file", "project-spec.pdf", "PDF content here", "stalbygg")

		req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": project.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToProject(rr, req)

		if rr.Code != http.StatusCreated {
			t.Logf("Response body: %s", rr.Body.String())
		}
		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "project-spec.pdf", result.Filename)
		assert.NotNil(t, result.ProjectID)
		assert.Equal(t, project.ID, *result.ProjectID)
		assert.Equal(t, domain.CompanyStalbygg, result.CompanyID)
	})

	t.Run("upload without company_id returns 400 for projects", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "test.txt", "content")

		req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": project.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToProject(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("upload to non-existent project returns 404", func(t *testing.T) {
		body, contentType := createMultipartFormFileWithCompany(t, "file", "test.txt", "content", "stalbygg")
		fakeID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, "/projects/"+fakeID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToProject(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("upload with invalid project ID returns 400", func(t *testing.T) {
		body, contentType := createMultipartFormFileWithCompany(t, "file", "test.txt", "content", "stalbygg")

		req := httptest.NewRequest(http.MethodPost, "/projects/not-a-uuid/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": "not-a-uuid"}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToProject(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestFileHandler_UploadToOffer(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	offer := createTestOfferForFiles(t, db, customer)

	t.Run("upload file to offer successfully (inherits company from offer)", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "offer-attachment.docx", "Word doc content")

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToOffer(rr, req)

		if rr.Code != http.StatusCreated {
			t.Logf("Response body: %s", rr.Body.String())
		}
		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "offer-attachment.docx", result.Filename)
		assert.NotNil(t, result.OfferID)
		assert.Equal(t, offer.ID, *result.OfferID)
		// File should inherit company from offer (stalbygg in this test setup)
		assert.Equal(t, domain.CompanyStalbygg, result.CompanyID)
	})

	t.Run("upload to non-existent offer returns 404", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "test.txt", "content")
		fakeID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, "/offers/"+fakeID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToOffer(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("upload with invalid offer ID returns 400", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "test.txt", "content")

		req := httptest.NewRequest(http.MethodPost, "/offers/bad-id/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": "bad-id"}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToOffer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestFileHandler_UploadToSupplier(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	supplier := testutil.CreateTestSupplier(t, db, "Test Supplier")

	t.Run("upload file to supplier successfully", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "supplier-contract.pdf", "Contract content")

		req := httptest.NewRequest(http.MethodPost, "/suppliers/"+supplier.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": supplier.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToSupplier(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "supplier-contract.pdf", result.Filename)
		assert.NotNil(t, result.SupplierID)
		assert.Equal(t, supplier.ID, *result.SupplierID)
	})

	t.Run("upload to non-existent supplier returns 404", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "test.txt", "content")
		fakeID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, "/suppliers/"+fakeID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToSupplier(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// ============================================================================
// List Files Tests
// ============================================================================

func TestFileHandler_ListCustomerFiles(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")

	// Upload a couple of files first
	for i := 0; i < 3; i++ {
		body, contentType := createMultipartFormFile(t, "file", "customer-file-"+string(rune('a'+i))+".txt", "content")
		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		req.Header.Set("Content-Type", contentType)
		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)
	}

	t.Run("list customer files successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/customers/"+customer.ID.String()+"/files", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))

		rr := httptest.NewRecorder()
		h.ListCustomerFiles(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var files []domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &files)
		require.NoError(t, err)
		assert.Len(t, files, 3)
	})

	t.Run("list files for non-existent customer returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/customers/"+fakeID.String()+"/files", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))

		rr := httptest.NewRecorder()
		h.ListCustomerFiles(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("list files with invalid customer ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/customers/invalid/files", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": "invalid"}))

		rr := httptest.NewRecorder()
		h.ListCustomerFiles(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestFileHandler_ListProjectFiles(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	project := createTestProjectForFiles(t, db, customer, "Test Project")

	// Upload files to the project
	for i := 0; i < 2; i++ {
		body, contentType := createMultipartFormFile(t, "file", "project-file-"+string(rune('1'+i))+".txt", "content")
		req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": project.ID.String()}))
		req.Header.Set("Content-Type", contentType)
		rr := httptest.NewRecorder()
		h.UploadToProject(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)
	}

	t.Run("list project files successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ID.String()+"/files", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": project.ID.String()}))

		rr := httptest.NewRecorder()
		h.ListProjectFiles(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var files []domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &files)
		require.NoError(t, err)
		assert.Len(t, files, 2)
	})

	t.Run("list files for non-existent project returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/projects/"+fakeID.String()+"/files", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))

		rr := httptest.NewRecorder()
		h.ListProjectFiles(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestFileHandler_ListOfferFiles(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	offer := createTestOfferForFiles(t, db, customer)

	// Upload files to the offer
	body, contentType := createMultipartFormFile(t, "file", "offer-doc.pdf", "pdf content")
	req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/files", body)
	req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": offer.ID.String()}))
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()
	h.UploadToOffer(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	t.Run("list offer files successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+offer.ID.String()+"/files", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.ListOfferFiles(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var files []domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &files)
		require.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, "offer-doc.pdf", files[0].Filename)
	})

	t.Run("list files for non-existent offer returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/offers/"+fakeID.String()+"/files", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))

		rr := httptest.NewRecorder()
		h.ListOfferFiles(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestFileHandler_ListSupplierFiles(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	supplier := testutil.CreateTestSupplier(t, db, "Test Supplier")

	// Upload files to the supplier
	body, contentType := createMultipartFormFile(t, "file", "supplier-doc.pdf", "pdf content")
	req := httptest.NewRequest(http.MethodPost, "/suppliers/"+supplier.ID.String()+"/files", body)
	req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": supplier.ID.String()}))
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()
	h.UploadToSupplier(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	t.Run("list supplier files successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers/"+supplier.ID.String()+"/files", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": supplier.ID.String()}))

		rr := httptest.NewRecorder()
		h.ListSupplierFiles(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var files []domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &files)
		require.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, "supplier-doc.pdf", files[0].Filename)
	})

	t.Run("list files for non-existent supplier returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/suppliers/"+fakeID.String()+"/files", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))

		rr := httptest.NewRecorder()
		h.ListSupplierFiles(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// ============================================================================
// GetByID Tests
// ============================================================================

func TestFileHandler_GetByID(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")

	// Upload a file first
	body, contentType := createMultipartFormFile(t, "file", "get-by-id-test.txt", "test content for GetByID")
	uploadReq := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
	uploadReq = uploadReq.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
	uploadReq.Header.Set("Content-Type", contentType)
	uploadRr := httptest.NewRecorder()
	h.UploadToCustomer(uploadRr, uploadReq)
	require.Equal(t, http.StatusCreated, uploadRr.Code)

	var uploadedFile domain.FileDTO
	err := json.Unmarshal(uploadRr.Body.Bytes(), &uploadedFile)
	require.NoError(t, err)

	t.Run("get file by ID successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/"+uploadedFile.ID.String(), nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": uploadedFile.ID.String()}))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, uploadedFile.ID, result.ID)
		assert.Equal(t, "get-by-id-test.txt", result.Filename)
	})

	t.Run("get non-existent file returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/files/"+fakeID.String(), nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("get with invalid ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/invalid-uuid", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": "invalid-uuid"}))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// ============================================================================
// Download Tests
// ============================================================================

func TestFileHandler_Download(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")

	// Upload a file first
	fileContent := "This is the content to download!"
	body, contentType := createMultipartFormFile(t, "file", "download-test.txt", fileContent)
	uploadReq := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
	uploadReq = uploadReq.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
	uploadReq.Header.Set("Content-Type", contentType)
	uploadRr := httptest.NewRecorder()
	h.UploadToCustomer(uploadRr, uploadReq)
	require.Equal(t, http.StatusCreated, uploadRr.Code)

	var uploadedFile domain.FileDTO
	err := json.Unmarshal(uploadRr.Body.Bytes(), &uploadedFile)
	require.NoError(t, err)

	t.Run("download file successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/"+uploadedFile.ID.String()+"/download", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": uploadedFile.ID.String()}))

		rr := httptest.NewRecorder()
		h.Download(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Header().Get("Content-Disposition"), "download-test.txt")

		downloadedContent, err := io.ReadAll(rr.Body)
		require.NoError(t, err)
		assert.Equal(t, fileContent, string(downloadedContent))
	})

	t.Run("download non-existent file returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/files/"+fakeID.String()+"/download", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))

		rr := httptest.NewRecorder()
		h.Download(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("download with invalid ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/not-valid/download", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": "not-valid"}))

		rr := httptest.NewRecorder()
		h.Download(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// ============================================================================
// Delete Tests
// ============================================================================

func TestFileHandler_Delete(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")

	t.Run("delete file successfully", func(t *testing.T) {
		// Upload a file first
		body, contentType := createMultipartFormFile(t, "file", "to-delete.txt", "content to delete")
		uploadReq := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		uploadReq = uploadReq.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		uploadReq.Header.Set("Content-Type", contentType)
		uploadRr := httptest.NewRecorder()
		h.UploadToCustomer(uploadRr, uploadReq)
		require.Equal(t, http.StatusCreated, uploadRr.Code)

		var uploadedFile domain.FileDTO
		err := json.Unmarshal(uploadRr.Body.Bytes(), &uploadedFile)
		require.NoError(t, err)

		// Delete the file
		deleteReq := httptest.NewRequest(http.MethodDelete, "/files/"+uploadedFile.ID.String(), nil)
		deleteReq = deleteReq.WithContext(withFileChiContext(ctx, map[string]string{"id": uploadedFile.ID.String()}))

		deleteRr := httptest.NewRecorder()
		h.Delete(deleteRr, deleteReq)

		assert.Equal(t, http.StatusNoContent, deleteRr.Code)

		// Verify file no longer exists
		getReq := httptest.NewRequest(http.MethodGet, "/files/"+uploadedFile.ID.String(), nil)
		getReq = getReq.WithContext(withFileChiContext(ctx, map[string]string{"id": uploadedFile.ID.String()}))
		getRr := httptest.NewRecorder()
		h.GetByID(getRr, getReq)

		assert.Equal(t, http.StatusNotFound, getRr.Code)
	})

	t.Run("delete non-existent file returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/files/"+fakeID.String(), nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": fakeID.String()}))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with invalid ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/files/invalid", nil)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": "invalid"}))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// ============================================================================
// Full Lifecycle Tests
// ============================================================================

func TestFileHandler_FullLifecycle(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Lifecycle Customer")

	t.Run("complete file lifecycle: upload -> get -> download -> delete", func(t *testing.T) {
		fileContent := "Lifecycle test content - upload, get metadata, download, then delete"

		// Step 1: Upload
		body, contentType := createMultipartFormFile(t, "file", "lifecycle-test.txt", fileContent)
		uploadReq := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		uploadReq = uploadReq.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		uploadReq.Header.Set("Content-Type", contentType)
		uploadRr := httptest.NewRecorder()
		h.UploadToCustomer(uploadRr, uploadReq)

		require.Equal(t, http.StatusCreated, uploadRr.Code)

		var uploadedFile domain.FileDTO
		err := json.Unmarshal(uploadRr.Body.Bytes(), &uploadedFile)
		require.NoError(t, err)
		assert.Equal(t, "lifecycle-test.txt", uploadedFile.Filename)
		assert.Equal(t, int64(len(fileContent)), uploadedFile.Size)

		// Step 2: Get metadata
		getReq := httptest.NewRequest(http.MethodGet, "/files/"+uploadedFile.ID.String(), nil)
		getReq = getReq.WithContext(withFileChiContext(ctx, map[string]string{"id": uploadedFile.ID.String()}))
		getRr := httptest.NewRecorder()
		h.GetByID(getRr, getReq)

		require.Equal(t, http.StatusOK, getRr.Code)

		var retrievedFile domain.FileDTO
		err = json.Unmarshal(getRr.Body.Bytes(), &retrievedFile)
		require.NoError(t, err)
		assert.Equal(t, uploadedFile.ID, retrievedFile.ID)

		// Step 3: Download
		downloadReq := httptest.NewRequest(http.MethodGet, "/files/"+uploadedFile.ID.String()+"/download", nil)
		downloadReq = downloadReq.WithContext(withFileChiContext(ctx, map[string]string{"id": uploadedFile.ID.String()}))
		downloadRr := httptest.NewRecorder()
		h.Download(downloadRr, downloadReq)

		require.Equal(t, http.StatusOK, downloadRr.Code)
		downloadedContent, _ := io.ReadAll(downloadRr.Body)
		assert.Equal(t, fileContent, string(downloadedContent))

		// Step 4: Delete
		deleteReq := httptest.NewRequest(http.MethodDelete, "/files/"+uploadedFile.ID.String(), nil)
		deleteReq = deleteReq.WithContext(withFileChiContext(ctx, map[string]string{"id": uploadedFile.ID.String()}))
		deleteRr := httptest.NewRecorder()
		h.Delete(deleteRr, deleteReq)

		require.Equal(t, http.StatusNoContent, deleteRr.Code)

		// Verify deletion
		verifyReq := httptest.NewRequest(http.MethodGet, "/files/"+uploadedFile.ID.String(), nil)
		verifyReq = verifyReq.WithContext(withFileChiContext(ctx, map[string]string{"id": uploadedFile.ID.String()}))
		verifyRr := httptest.NewRecorder()
		h.GetByID(verifyRr, verifyReq)

		assert.Equal(t, http.StatusNotFound, verifyRr.Code)
	})
}

// ============================================================================
// File Isolation Tests
// ============================================================================

func TestFileHandler_FileIsolation(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer1 := testutil.CreateTestCustomer(t, db, "Customer 1")
	customer2 := testutil.CreateTestCustomer(t, db, "Customer 2")

	t.Run("files are isolated between entities", func(t *testing.T) {
		// Upload file to customer 1
		body1, ct1 := createMultipartFormFile(t, "file", "customer1-file.txt", "customer 1 content")
		req1 := httptest.NewRequest(http.MethodPost, "/customers/"+customer1.ID.String()+"/files", body1)
		req1 = req1.WithContext(withFileChiContext(ctx, map[string]string{"id": customer1.ID.String()}))
		req1.Header.Set("Content-Type", ct1)
		rr1 := httptest.NewRecorder()
		h.UploadToCustomer(rr1, req1)
		require.Equal(t, http.StatusCreated, rr1.Code)

		// Upload file to customer 2
		body2, ct2 := createMultipartFormFile(t, "file", "customer2-file.txt", "customer 2 content")
		req2 := httptest.NewRequest(http.MethodPost, "/customers/"+customer2.ID.String()+"/files", body2)
		req2 = req2.WithContext(withFileChiContext(ctx, map[string]string{"id": customer2.ID.String()}))
		req2.Header.Set("Content-Type", ct2)
		rr2 := httptest.NewRecorder()
		h.UploadToCustomer(rr2, req2)
		require.Equal(t, http.StatusCreated, rr2.Code)

		// List files for customer 1
		listReq1 := httptest.NewRequest(http.MethodGet, "/customers/"+customer1.ID.String()+"/files", nil)
		listReq1 = listReq1.WithContext(withFileChiContext(ctx, map[string]string{"id": customer1.ID.String()}))
		listRr1 := httptest.NewRecorder()
		h.ListCustomerFiles(listRr1, listReq1)

		var files1 []domain.FileDTO
		err := json.Unmarshal(listRr1.Body.Bytes(), &files1)
		require.NoError(t, err)
		assert.Len(t, files1, 1)
		assert.Equal(t, "customer1-file.txt", files1[0].Filename)

		// List files for customer 2
		listReq2 := httptest.NewRequest(http.MethodGet, "/customers/"+customer2.ID.String()+"/files", nil)
		listReq2 = listReq2.WithContext(withFileChiContext(ctx, map[string]string{"id": customer2.ID.String()}))
		listRr2 := httptest.NewRecorder()
		h.ListCustomerFiles(listRr2, listReq2)

		var files2 []domain.FileDTO
		err = json.Unmarshal(listRr2.Body.Bytes(), &files2)
		require.NoError(t, err)
		assert.Len(t, files2, 1)
		assert.Equal(t, "customer2-file.txt", files2[0].Filename)
	})
}

// ============================================================================
// Edge Cases Tests
// ============================================================================

func TestFileHandler_EdgeCases(t *testing.T) {
	db := setupFileHandlerTestDB(t)
	store := setupTestStorage(t)
	h := createFileHandler(t, db, store)
	ctx := createFileTestContext()

	customer := testutil.CreateTestCustomer(t, db, "Edge Case Customer")

	t.Run("upload file with special characters in name", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "file with spaces & special (chars).txt", "content")
		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "file with spaces & special (chars).txt", result.Filename)
	})

	t.Run("upload file with unicode characters in name", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "dokument-oversikt.txt", "content with unicode")
		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "dokument-oversikt.txt", result.Filename)
	})

	t.Run("upload empty file name handled", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "")
		require.NoError(t, err)
		_, err = part.Write([]byte("content"))
		require.NoError(t, err)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		// Should still work - creates file with empty name
		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("upload preserves file extension", func(t *testing.T) {
		body, contentType := createMultipartFormFile(t, "file", "spreadsheet.xlsx", "excel content")
		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/files", body)
		req = req.WithContext(withFileChiContext(ctx, map[string]string{"id": customer.ID.String()}))
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		h.UploadToCustomer(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)

		var result domain.FileDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.True(t, filepath.Ext(result.Filename) == ".xlsx")
	})
}
