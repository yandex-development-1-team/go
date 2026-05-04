package postgres

import (
	"context"
	"testing"
)

func insertApplicationWithManager(t *testing.T, customerName, contactInfo, formAnswerID string, managerID int64) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(`
		INSERT INTO applications (customer_name, contact_info, form_answer_id, status, description, manager_id)
		VALUES ($1, $2, $3, 'pending', '', $4)
		RETURNING id`,
		customerName, contactInfo, formAnswerID, managerID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert application with manager: %v", err)
	}
	return id
}

func insertUser(t *testing.T, telegramID int64, username string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(`
			INSERT INTO users (telegram_id, username)
			VALUES ($1, $2)
			RETURNING id`,
		telegramID, username,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return telegramID
}

func insertService(t *testing.T, name, slug string, price int) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(`
			INSERT INTO services (name, slug, price)
			VALUES ($1, $2, $3)
			RETURNING id`,
		name, slug, price,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert service: %v", err)
	}
	return id
}

func insertBooking(t *testing.T, userTelegramID, serviceID, managerID int64, status string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(`
			INSERT INTO bookings (user_id, service_id, booking_date, guest_name, status, manager_id)
			VALUES ($1, $2, CURRENT_DATE, 'Test Guest', $3, $4)
			RETURNING id`,
		userTelegramID, serviceID, status, managerID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert booking: %v", err)
	}
	return id
}

func clearBookings(t *testing.T) {
	t.Helper()
	_, err := db.Exec("DELETE FROM bookings")
	if err != nil {
		t.Fatalf("clear bookings: %v", err)
	}
}

func clearUsers(t *testing.T) {
	t.Helper()
	_, err := db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("clear users: %v", err)
	}
}

func clearServices(t *testing.T) {
	t.Helper()
	_, err := db.Exec("DELETE FROM services")
	if err != nil {
		t.Fatalf("clear services: %v", err)
	}
}

func TestGetDashboard_Overview(t *testing.T) {
	clearApplications(t)
	clearBookings(t)
	clearUsers(t)
	clearServices(t)
	clearStaff(t)

	managerID := insertManager(t, "m1@test.com", "Ivan", "Ivanov")
	userID := insertUser(t, 100001, "testuser")
	serviceID := insertService(t, "Экскурсия", "tour", 1000)

	// 2 pending — из разных таблиц
	insertApplication(t, "Клиент 1", "@client1", "form-101")
	insertBooking(t, userID, serviceID, managerID, "pending")

	// 1 confirmed
	insertBooking(t, userID, serviceID, managerID, "confirmed")

	repo := NewStaffRepo(db)
	result, err := repo.GetDashboard(context.Background(), managerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Overview.NewApplications != 2 {
		t.Errorf("expected 2 new_applications, got %d", result.Overview.NewApplications)
	}
	if result.Overview.InProgressApplications != 1 {
		t.Errorf("expected 1 in_progress, got %d", result.Overview.InProgressApplications)
	}
}

func TestGetDashboard_ManagerStats(t *testing.T) {
	clearApplications(t)
	clearBookings(t)
	clearUsers(t)
	clearServices(t)
	clearStaff(t)

	managerID := insertManager(t, "m2@test.com", "Petr", "Petrov")
	otherManagerID := insertManager(t, "m3@test.com", "Serg", "Sergov")
	userID := insertUser(t, 100002, "testuser2")
	serviceID := insertService(t, "Лекция", "lecture", 500)

	// confirmed для нашего менеджера
	insertBooking(t, userID, serviceID, managerID, "confirmed")
	insertBooking(t, userID, serviceID, managerID, "confirmed")
	// cancelled для нашего менеджера
	insertBooking(t, userID, serviceID, managerID, "cancelled")
	// confirmed для другого — не должен попасть в счётчик
	insertBooking(t, userID, serviceID, otherManagerID, "confirmed")

	repo := NewStaffRepo(db)
	result, err := repo.GetDashboard(context.Background(), managerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ManagerStats.InProgress != 2 {
		t.Errorf("expected 2 in_progress, got %d", result.ManagerStats.InProgress)
	}
	if result.ManagerStats.Processed != 1 {
		t.Errorf("expected 1 processed, got %d", result.ManagerStats.Processed)
	}
}

func TestGetDashboard_ApplicationsList(t *testing.T) {
	clearApplications(t)
	clearBookings(t)
	clearUsers(t)
	clearServices(t)
	clearStaff(t)

	managerID := insertManager(t, "m4@test.com", "Anna", "Annova")
	userID := insertUser(t, 100003, "testuser3")
	serviceID := insertService(t, "Мастеркласс", "masterclass", 2000)

	insertBooking(t, userID, serviceID, managerID, "pending")
	insertBooking(t, userID, serviceID, managerID, "confirmed")
	// cancelled — не должен попасть в список
	insertBooking(t, userID, serviceID, managerID, "cancelled")

	repo := NewStaffRepo(db)
	result, err := repo.GetDashboard(context.Background(), managerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Applications) != 2 {
		t.Errorf("expected 2 applications, got %d", len(result.Applications))
	}
}

func TestGetDashboard_Empty(t *testing.T) {
	clearApplications(t)
	clearBookings(t)
	clearUsers(t)
	clearServices(t)
	clearStaff(t)

	managerID := insertManager(t, "m5@test.com", "Empty", "Manager")

	repo := NewStaffRepo(db)
	result, err := repo.GetDashboard(context.Background(), managerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Overview.NewApplications != 0 {
		t.Errorf("expected 0, got %d", result.Overview.NewApplications)
	}
	if len(result.Applications) != 0 {
		t.Errorf("expected empty slice, got %d", len(result.Applications))
	}
}

func TestGetDashboard_MixedApplicationsList(t *testing.T) {
	clearApplications(t)
	clearBookings(t)
	clearUsers(t)
	clearServices(t)
	clearStaff(t)

	managerID := insertManager(t, "m6@test.com", "Mix", "Manager")
	userID := insertUser(t, 100004, "mixuser")
	serviceID := insertService(t, "Тур", "tour2", 300)

	insertApplicationWithManager(t, "Клиент из applications", "@appclient", "form-mix-1", managerID)
	insertBooking(t, userID, serviceID, managerID, "pending")

	repo := NewStaffRepo(db)
	result, err := repo.GetDashboard(context.Background(), managerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Applications) != 2 {
		t.Errorf("expected 2 applications (1 app + 1 booking), got %d", len(result.Applications))
	}
}

func TestGetDashboard_OtherManagerNotInList(t *testing.T) {
	clearApplications(t)
	clearBookings(t)
	clearUsers(t)
	clearServices(t)
	clearStaff(t)

	managerID := insertManager(t, "m7@test.com", "My", "Manager")
	otherManagerID := insertManager(t, "m8@test.com", "Other", "Manager")
	userID := insertUser(t, 100005, "user5")
	serviceID := insertService(t, "Лекция2", "lecture2", 700)

	insertBooking(t, userID, serviceID, managerID, "pending")
	insertBooking(t, userID, serviceID, otherManagerID, "pending")

	repo := NewStaffRepo(db)
	result, err := repo.GetDashboard(context.Background(), managerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Applications) != 1 {
		t.Errorf("expected 1 application, got %d — other manager leaked", len(result.Applications))
	}
}

func TestGetDashboard_TelegramNickFormat(t *testing.T) {
	clearApplications(t)
	clearBookings(t)
	clearUsers(t)
	clearServices(t)
	clearStaff(t)

	managerID := insertManager(t, "m9@test.com", "Nick", "Checker")
	userID := insertUser(t, 100006, "nicktestuser")
	serviceID := insertService(t, "Ивент", "event1", 100)

	insertBooking(t, userID, serviceID, managerID, "pending")

	repo := NewStaffRepo(db)
	result, err := repo.GetDashboard(context.Background(), managerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Applications) == 0 {
		t.Fatal("expected at least 1 application")
	}
	if result.Applications[0].TelegramNick != "@nicktestuser" {
		t.Errorf("expected @nicktestuser, got %q", result.Applications[0].TelegramNick)
	}
}

func TestGetDashboard_ServiceType_Booking(t *testing.T) {
	clearApplications(t)
	clearBookings(t)
	clearUsers(t)
	clearServices(t)
	clearStaff(t)

	managerID := insertManager(t, "m10@test.com", "Type", "Checker")
	userID := insertUser(t, 100007, "typeuser")
	serviceID := insertService(t, "Моя коробка", "box1", 500)

	insertBooking(t, userID, serviceID, managerID, "pending")

	repo := NewStaffRepo(db)
	result, err := repo.GetDashboard(context.Background(), managerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Applications) == 0 {
		t.Fatal("expected at least 1 application")
	}
	if result.Applications[0].ServiceType != "box_solution" {
		t.Errorf("expected box_solution, got %q", result.Applications[0].ServiceType)
	}
	if result.Applications[0].ServiceName != "Моя коробка" {
		t.Errorf("expected Моя коробка, got %q", result.Applications[0].ServiceName)
	}
}

func TestGetDashboard_ServiceType_Application(t *testing.T) {
	clearApplications(t)
	clearBookings(t)
	clearUsers(t)
	clearServices(t)
	clearStaff(t)

	managerID := insertManager(t, "m11@test.com", "App", "Checker")

	insertApplicationWithManager(t, "Клиент спец", "@speclient", "form-spec-1", managerID)

	repo := NewStaffRepo(db)
	result, err := repo.GetDashboard(context.Background(), managerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Applications) == 0 {
		t.Fatal("expected at least 1 application")
	}
	if result.Applications[0].ServiceType != "spec_project" {
		t.Errorf("expected spec_project, got %q", result.Applications[0].ServiceType)
	}
	if result.Applications[0].ServiceName != "spec_project" {
		t.Errorf("expected spec_project, got %q", result.Applications[0].ServiceName)
	}
}
