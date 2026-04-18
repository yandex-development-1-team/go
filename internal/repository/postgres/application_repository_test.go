package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/yandex-development-1-team/go/internal/models"
)

func insertManager(t *testing.T, email, firstName, lastName string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(`
		INSERT INTO staff (email, first_name, last_name, role, status)
		VALUES ($1, $2, $3, 'manager_1', 'active')
		RETURNING id`,
		email, firstName, lastName,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert manager: %v", err)
	}
	return id
}

func clearStaff(t *testing.T) {
	t.Helper()
	_, err := db.Exec("DELETE FROM staff")
	if err != nil {
		t.Fatalf("clear staff: %v", err)
	}
}
func clearApplications(t *testing.T) {
	t.Helper()
	_, err := db.Exec("DELETE FROM applications")
	if err != nil {
		t.Fatalf("clear applications: %v", err)
	}
}

func insertApplication(t *testing.T, customerName, contactInfo, formAnswerID string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(`
			INSERT INTO applications (customer_name, contact_info, form_answer_id, status, description)
			VALUES ($1, $2, $3, 'pending', '')
			RETURNING id`,
		customerName, contactInfo, formAnswerID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert application: %v", err)
	}
	return id
}

// -- GetApplicationByID --

func TestGetApplicationByID_Found(t *testing.T) {
	clearApplications(t)
	id := insertApplication(t, "Иван Иванов", "@ivan", "form-001")

	repo := NewApplicationRepository(db)
	app, err := repo.GetApplicationByID(context.Background(), id)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID != id {
		t.Errorf("expected id %d, got %d", id, app.ID)
	}
	if app.CustomerName != "Иван Иванов" {
		t.Errorf("expected customer_name %q, got %q", "Иван Иванов", app.CustomerName)
	}
	if app.Status != "pending" {
		t.Errorf("expected status pending, got %q", app.Status)
	}
}

func TestGetApplicationByID_NotFound(t *testing.T) {
	clearApplications(t)

	repo := NewApplicationRepository(db)
	_, err := repo.GetApplicationByID(context.Background(), 99999)

	if !errors.Is(err, models.ErrApplicationNotFound) {
		t.Errorf("expected ErrApplicationNotFound, got %v", err)
	}
}

// -- UpdateApplicationStatus --

func TestUpdateApplicationStatus_Success(t *testing.T) {
	clearApplications(t)
	id := insertApplication(t, "Мария", "@maria", "form-002")

	repo := NewApplicationRepository(db)
	err := repo.UpdateApplicationStatus(context.Background(), id, "confirmed")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	app, err := repo.GetApplicationByID(context.Background(), id)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if app.Status != "confirmed" {
		t.Errorf("expected confirmed, got %q", app.Status)
	}
}

func TestUpdateApplicationStatus_NotFound(t *testing.T) {
	clearApplications(t)

	repo := NewApplicationRepository(db)
	err := repo.UpdateApplicationStatus(context.Background(), 99999, "confirmed")

	if !errors.Is(err, models.ErrApplicationNotFound) {
		t.Errorf("expected ErrApplicationNotFound, got %v", err)
	}
}

// -- DeleteApplication --

func TestDeleteApplication_Success(t *testing.T) {
	clearApplications(t)
	id := insertApplication(t, "Петр", "@petr", "form-003")

	repo := NewApplicationRepository(db)
	err := repo.DeleteApplication(context.Background(), id)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetApplicationByID(context.Background(), id)
	if !errors.Is(err, models.ErrApplicationNotFound) {
		t.Errorf("expected ErrApplicationNotFound after delete, got %v", err)
	}
}

func TestDeleteApplication_NotFound(t *testing.T) {
	clearApplications(t)

	repo := NewApplicationRepository(db)
	err := repo.DeleteApplication(context.Background(), 99999)

	if !errors.Is(err, models.ErrApplicationNotFound) {
		t.Errorf("expected ErrApplicationNotFound, got %v", err)
	}
}

// -- ApplicationsList --

func TestApplicationsList_Empty(t *testing.T) {
	clearApplications(t)

	repo := NewApplicationRepository(db)
	result, err := repo.ApplicationsList(context.Background(), &models.ApplicationFilter{
		Limit:  20,
		Offset: 0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestApplicationsList_FilterByStatus(t *testing.T) {
	clearApplications(t)
	insertApplication(t, "Анна", "@anna", "form-004")
	insertApplication(t, "Борис", "@boris", "form-005")

	// меняем статус одной
	_, err := db.Exec(`UPDATE applications SET status = 'confirmed' WHERE form_answer_id = 'form-005'`)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	repo := NewApplicationRepository(db)
	result, err := repo.ApplicationsList(context.Background(), &models.ApplicationFilter{
		Status: "confirmed",
		Limit:  20,
		Offset: 0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].CustomerName != "Борис" {
		t.Errorf("expected Борис, got %q", result.Items[0].CustomerName)
	}
}

func TestApplicationsList_Pagination(t *testing.T) {
	clearApplications(t)
	for i := 0; i < 5; i++ {
		insertApplication(t,
			fmt.Sprintf("Customer %d", i),
			fmt.Sprintf("@contact%d", i),
			fmt.Sprintf("form-pag-%d", i),
		)
	}

	repo := NewApplicationRepository(db)
	result, err := repo.ApplicationsList(context.Background(), &models.ApplicationFilter{
		Limit:  2,
		Offset: 0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Items))
	}
	if result.Total != 5 {
		t.Errorf("expected total 5, got %d", result.Total)
	}
}

// -- DeleteApplication --

func TestDeleteApplication_AlreadyDeleted(t *testing.T) {
	clearApplications(t)
	id := insertApplication(t, "Петр", "@petr", "form-deleted-002")

	repo := NewApplicationRepository(db)

	if err := repo.DeleteApplication(context.Background(), id); err != nil {
		t.Fatalf("first delete: %v", err)
	}

	err := repo.DeleteApplication(context.Background(), id)
	if !errors.Is(err, models.ErrApplicationNotFound) {
		t.Errorf("expected ErrApplicationNotFound on second delete, got %v", err)
	}
}

// -- ApplicationsList --

func TestApplicationsList_FilterByManagerID(t *testing.T) {
	clearApplications(t)
	clearStaff(t)

	managerID := insertManager(t, "manager@test.com", "Иван", "Менеджеров")
	defer clearStaff(t)

	_, err := db.Exec(`
    INSERT INTO applications (customer_name, contact_info, form_answer_id, status, manager_id, description)
    VALUES ('С менеджером', '@with', 'form-mgr-001', 'pending', $1, ' ')`, managerID)
	if err != nil {
		t.Fatalf("insert with manager: %v", err)
	}
	insertApplication(t, "Без менеджера", "@without", "form-mgr-002")

	repo := NewApplicationRepository(db)
	result, err := repo.ApplicationsList(context.Background(), &models.ApplicationFilter{
		ManagerID: managerID,
		Limit:     20,
		Offset:    0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].ManagerID != managerID {
		t.Errorf("expected manager_id %d, got %d", managerID, result.Items[0].ManagerID)
	}
}

func TestApplicationsList_FilterByCustomerName(t *testing.T) {
	clearApplications(t)
	insertApplication(t, "Александр Петров", "@alex", "form-name-001")
	insertApplication(t, "Мария Сидорова", "@maria", "form-name-002")

	repo := NewApplicationRepository(db)
	result, err := repo.ApplicationsList(context.Background(), &models.ApplicationFilter{
		CustomerName: "алекс", // ILIKE — регистр не важен
		Limit:        20,
		Offset:       0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].CustomerName != "Александр Петров" {
		t.Errorf("expected Александр Петров, got %q", result.Items[0].CustomerName)
	}
}

func TestApplicationsList_MultipleFilters(t *testing.T) {
	clearApplications(t)
	insertApplication(t, "Тест Фильтров", "@test", "form-multi-001")
	insertApplication(t, "Тест Фильтров", "@test2", "form-multi-002")

	_, err := db.Exec(`
		UPDATE applications SET status = 'confirmed' 
		WHERE form_answer_id = 'form-multi-001'`)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	repo := NewApplicationRepository(db)
	result, err := repo.ApplicationsList(context.Background(), &models.ApplicationFilter{
		CustomerName: "Тест",
		Status:       "confirmed",
		Limit:        20,
		Offset:       0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Status != "confirmed" {
		t.Errorf("expected confirmed, got %q", result.Items[0].Status)
	}
}

func TestApplicationsList_Offset(t *testing.T) {
	clearApplications(t)
	for i := 0; i < 5; i++ {
		insertApplication(t,
			fmt.Sprintf("Customer %d", i),
			fmt.Sprintf("@contact%d", i),
			fmt.Sprintf("form-offset-%d", i),
		)
	}

	repo := NewApplicationRepository(db)

	page1, err := repo.ApplicationsList(context.Background(), &models.ApplicationFilter{
		Limit:  2,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("page1: %v", err)
	}

	page2, err := repo.ApplicationsList(context.Background(), &models.ApplicationFilter{
		Limit:  2,
		Offset: 2,
	})
	if err != nil {
		t.Fatalf("page2: %v", err)
	}

	if len(page2.Items) != 2 {
		t.Errorf("expected 2 items on page2, got %d", len(page2.Items))
	}
	if page1.Items[0].ID == page2.Items[0].ID {
		t.Error("pages overlap — same item on both pages")
	}
	if page1.Total != 5 || page2.Total != 5 {
		t.Errorf("expected total 5 on both pages, got %d and %d", page1.Total, page2.Total)
	}
}

// -- CreateApplication --

func TestCreateApplication_AssignsManager(t *testing.T) {
	clearApplications(t)
	clearStaff(t)

	insertManager(t, "assign@test.com", "Мария", "Менеджерова")
	defer clearStaff(t)

	repo := NewApplicationRepository(db)
	err := repo.CreateApplication(context.Background(), &models.Application{
		CustomerName: "Новый клиент",
		ContactInfo:  "@newclient",
		Description:  "тестовое описание",
		FormAnswerId: "form-create-001",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var managerID int64
	err = db.QueryRow(`
		SELECT manager_id FROM applications 
		WHERE form_answer_id = 'form-create-001'`,
	).Scan(&managerID)
	if err != nil {
		t.Fatalf("get manager_id: %v", err)
	}
	if managerID == 0 {
		t.Error("expected manager to be assigned, got 0")
	}
}

func TestCreateApplication_DuplicateFormAnswerID(t *testing.T) {
	clearApplications(t)

	repo := NewApplicationRepository(db)
	app := &models.Application{
		CustomerName: "Клиент",
		ContactInfo:  "@client",
		Description:  "описание",
		FormAnswerId: "form-dup-001",
	}

	if err := repo.CreateApplication(context.Background(), app); err != nil {
		t.Fatalf("first create: %v", err)
	}

	err := repo.CreateApplication(context.Background(), app)
	if err == nil {
		t.Error("expected error on duplicate form_answer_id, got nil")
	}
}

func TestCreateApplication_NoManagers(t *testing.T) {
	clearApplications(t)
	clearStaff(t)

	repo := NewApplicationRepository(db)
	err := repo.CreateApplication(context.Background(), &models.Application{
		CustomerName: "Клиент",
		ContactInfo:  "@client",
		Description:  "описание",
		FormAnswerId: "form-nomanager-001",
	})

	// подзапрос вернёт NULL для manager_id — зависит от схемы
	// если manager_id NOT NULL — ожидаем ошибку
	// если nullable — заявка создастся без менеджера
	if err != nil {
		t.Logf("got expected error when no managers: %v", err)
	} else {
		var managerID *int64
		_ = db.QueryRow(`
			SELECT manager_id FROM applications 
			WHERE form_answer_id = 'form-nomanager-001'`,
		).Scan(&managerID)
		if managerID != nil {
			t.Errorf("expected nil manager_id, got %v", *managerID)
		}
	}
}
