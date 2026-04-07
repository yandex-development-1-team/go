package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/models"
)

func TestCreateBox(t *testing.T) {
	// ctx := context.Background()

	// helper для очистки таблиц
	truncateTables := func() {
		truncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := db.ExecContext(truncCtx, "TRUNCATE services, service_available_slots RESTART IDENTITY CASCADE")
		require.NoError(t, err)
	}

	t.Run("success - create box without slots", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		name := "Тестовая коробка"
		slug := "test-box"
		description := "Описание"
		rules := "Правила"
		location := "Москва"
		price := 1000
		status := "active"
		organizer := "Организатор"

		box := &models.BoxCreate{
			Name:        &name,
			Slug:        &slug,
			Description: &description,
			Rules:       &rules,
			Location:    &location,
			Price:       &price,
			Image:       nil,
			Status:      &status,
			Organizer:   &organizer,
			Slots:       []models.BoxAvailableSlot{},
		}

		result, err := boxRepo.CreateBox(ctx, box)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.NotZero(t, result.ID)
		assert.Equal(t, name, result.Name)
		assert.Equal(t, slug, result.Slug)
		assert.Equal(t, description, result.Description)
		assert.Equal(t, rules, result.Rules)
		assert.Equal(t, location, result.Location)
		assert.Equal(t, price, result.Price)
		assert.Equal(t, status, result.Status)
		assert.Equal(t, organizer, result.Organizer)
		assert.Empty(t, result.BoxAvailableSlots)
		assert.NotZero(t, result.CreatedAt)
		assert.NotZero(t, result.UpdatedAt)

		// Проверяем, что данные сохранились в БД
		var dbCount int
		err = db.GetContext(ctx, &dbCount, "SELECT COUNT(*) FROM services WHERE id = $1 AND deleted_at IS NULL", result.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, dbCount)
	})

	t.Run("success - create box with slots", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		name := "Квест с временными слотами"
		slug := "quest-with-slots"
		price := 1500
		status := "active"

		slots := []models.BoxAvailableSlot{
			{Date: "2024-05-01", StartTime: "10:00", EndTime: "12:00"},
			{Date: "2024-05-01", StartTime: "14:00", EndTime: "16:00"},
			{Date: "2024-05-02", StartTime: "11:00", EndTime: "13:00"},
		}

		box := &models.BoxCreate{
			Name:   &name,
			Slug:   &slug,
			Price:  &price,
			Status: &status,
			Slots:  slots,
		}

		result, err := boxRepo.CreateBox(ctx, box)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.NotZero(t, result.ID)
		assert.Equal(t, name, result.Name)
		assert.Equal(t, price, result.Price)
		assert.Len(t, result.BoxAvailableSlots, 3)

		// Проверяем слоты
		assert.Equal(t, "2024-05-01", result.BoxAvailableSlots[0].Date)
		assert.Equal(t, "10:00", result.BoxAvailableSlots[0].StartTime)
		assert.Equal(t, "12:00", result.BoxAvailableSlots[0].EndTime)

		// Проверяем, что слоты сохранились в БД
		var slotsCount int
		err = db.GetContext(ctx, &slotsCount,
			"SELECT COUNT(*) FROM service_available_slots WHERE service_id = $1", result.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, slotsCount)
	})

	t.Run("success - create box with full day slot (no time)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		name := "Весь день"
		slug := "full-day"
		price := 2000
		status := "active"

		slots := []models.BoxAvailableSlot{
			{Date: "2024-05-01", StartTime: "", EndTime: ""}, // весь день
		}

		box := &models.BoxCreate{
			Name:   &name,
			Slug:   &slug,
			Price:  &price,
			Status: &status,
			Slots:  slots,
		}

		result, err := boxRepo.CreateBox(ctx, box)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.BoxAvailableSlots, 1)
		assert.Equal(t, "2024-05-01", result.BoxAvailableSlots[0].Date)
		assert.Empty(t, result.BoxAvailableSlots[0].StartTime)
		assert.Empty(t, result.BoxAvailableSlots[0].EndTime)

		// Проверяем, что слот сохранился с NULL временем
		var startTime sql.NullString
		var endTime sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT start_time, end_time FROM service_available_slots WHERE service_id = $1", result.ID).
			Scan(&startTime, &endTime)
		require.NoError(t, err)
		assert.False(t, startTime.Valid)
		assert.False(t, endTime.Valid)
	})

	t.Run("error - nil box", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		result, err := boxRepo.CreateBox(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "service cannot be nil")
	})

	t.Run("error - empty name", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		name := ""
		slug := "test"
		price := 1000
		status := "active"

		box := &models.BoxCreate{
			Name:   &name,
			Slug:   &slug,
			Price:  &price,
			Status: &status,
			Slots:  []models.BoxAvailableSlot{},
		}

		result, err := boxRepo.CreateBox(ctx, box)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("error - invalid date format", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		name := "Test"
		slug := "test"
		price := 1000
		status := "active"

		slots := []models.BoxAvailableSlot{
			{Date: "invalid-date", StartTime: "10:00", EndTime: "12:00"},
		}

		box := &models.BoxCreate{
			Name:   &name,
			Slug:   &slug,
			Price:  &price,
			Status: &status,
			Slots:  slots,
		}

		result, err := boxRepo.CreateBox(ctx, box)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid date format")
	})

	t.Run("error - duplicate slug", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		// Создаем первую коробку
		name1 := "First Box"
		slug1 := "same-slug"
		price1 := 1000
		status1 := "active"

		box1 := &models.BoxCreate{
			Name:   &name1,
			Slug:   &slug1,
			Price:  &price1,
			Status: &status1,
			Slots:  []models.BoxAvailableSlot{},
		}

		_, err := boxRepo.CreateBox(ctx, box1)
		require.NoError(t, err)

		// Пытаемся создать вторую коробку с таким же slug
		name2 := "Second Box"
		slug2 := "same-slug"
		price2 := 2000
		status2 := "active"

		box2 := &models.BoxCreate{
			Name:   &name2,
			Slug:   &slug2,
			Price:  &price2,
			Status: &status2,
			Slots:  []models.BoxAvailableSlot{},
		}

		result, err := boxRepo.CreateBox(ctx, box2)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		name := "Test"
		slug := "test"
		price := 1000
		status := "active"

		box := &models.BoxCreate{
			Name:   &name,
			Slug:   &slug,
			Price:  &price,
			Status: &status,
			Slots:  []models.BoxAvailableSlot{},
		}

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		result, err := boxRepo.CreateBox(cancelledCtx, box)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("success - create box with multiple slots on same date", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		name := "Много слотов"
		slug := "many-slots"
		price := 3000
		status := "active"

		slots := []models.BoxAvailableSlot{
			{Date: "2024-06-01", StartTime: "09:00", EndTime: "11:00"},
			{Date: "2024-06-01", StartTime: "11:00", EndTime: "13:00"},
			{Date: "2024-06-01", StartTime: "14:00", EndTime: "16:00"},
			{Date: "2024-06-02", StartTime: "10:00", EndTime: "12:00"},
		}

		box := &models.BoxCreate{
			Name:   &name,
			Slug:   &slug,
			Price:  &price,
			Status: &status,
			Slots:  slots,
		}

		result, err := boxRepo.CreateBox(ctx, box)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.BoxAvailableSlots, 4)

		// Проверяем порядок слотов
		assert.Equal(t, "2024-06-01", result.BoxAvailableSlots[0].Date)
		assert.Equal(t, "09:00", result.BoxAvailableSlots[0].StartTime)
		assert.Equal(t, "2024-06-01", result.BoxAvailableSlots[1].Date)
		assert.Equal(t, "11:00", result.BoxAvailableSlots[1].StartTime)
		assert.Equal(t, "2024-06-01", result.BoxAvailableSlots[2].Date)
		assert.Equal(t, "14:00", result.BoxAvailableSlots[2].StartTime)
		assert.Equal(t, "2024-06-02", result.BoxAvailableSlots[3].Date)
		assert.Equal(t, "10:00", result.BoxAvailableSlots[3].StartTime)
	})

	t.Run("success - create box with all optional fields", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		name := "Полная коробка"
		slug := "full-box"
		description := "Полное описание"
		rules := "Полные правила"
		location := "Санкт-Петербург"
		price := 5000
		image := "https://example.com/image.jpg"
		status := "active"
		organizer := "ООО Организатор"

		box := &models.BoxCreate{
			Name:        &name,
			Slug:        &slug,
			Description: &description,
			Rules:       &rules,
			Location:    &location,
			Price:       &price,
			Image:       &image,
			Status:      &status,
			Organizer:   &organizer,
			Slots:       []models.BoxAvailableSlot{},
		}

		result, err := boxRepo.CreateBox(ctx, box)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, name, result.Name)
		assert.Equal(t, description, result.Description)
		assert.Equal(t, rules, result.Rules)
		assert.Equal(t, location, result.Location)
		assert.Equal(t, price, result.Price)
		assert.Equal(t, image, *result.Image)
		assert.Equal(t, organizer, result.Organizer)
	})

	t.Run("error - empty slots date", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		truncateTables()

		name := "Test"
		slug := "test"
		price := 1000
		status := "active"

		slots := []models.BoxAvailableSlot{
			{Date: "", StartTime: "10:00", EndTime: "12:00"},
		}

		box := &models.BoxCreate{
			Name:   &name,
			Slug:   &slug,
			Price:  &price,
			Status: &status,
			Slots:  slots,
		}

		result, err := boxRepo.CreateBox(ctx, box)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "slot date cannot be empty")
	})
}

// helper функция
func extractIDs(services []models.Service) []int64 {
	ids := make([]int64, len(services))
	for i, s := range services {
		ids[i] = s.ID
	}
	return ids
}

func TestList(t *testing.T) {
	ctx := context.Background()

	// helper для очистки таблиц
	truncateTables := func() {
		_, err := db.ExecContext(ctx, "TRUNCATE services, service_available_slots RESTART IDENTITY CASCADE")
		require.NoError(t, err)
	}

	// helper для вставки сервиса
	insertService := func(t *testing.T, name string, status string, price int) int64 {
		t.Helper()
		suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
		// Генерируем уникальный slug
		slug := name + "-" + suffix
		var id int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO services (name, slug, description, rules, location, price, image, status, organizer)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`,
			name, slug, "Описание "+name, "Правила "+name, "Москва", price, nil, status, "Организатор",
		).Scan(&id)
		require.NoError(t, err, "Ошибка вставки сервиса: %s", name)
		return id
	}

	// helper для вставки сервиса со слотами
	insertServiceWithSlots := func(t *testing.T, name string, status string, price int, slots []struct {
		date, start, end string
	}) int64 {
		t.Helper()
		id := insertService(t, name, status, price)

		for _, slot := range slots {
			_, err := db.ExecContext(ctx, `
				INSERT INTO service_available_slots (service_id, slot_date, start_time, end_time)
				VALUES ($1, $2, $3, $4)`,
				id, slot.date, slot.start, slot.end,
			)
			require.NoError(t, err)
		}
		return id
	}

	t.Run("returns all services with default pagination", func(t *testing.T) {
		truncateTables()

		// Вставляем 25 сервисов
		for i := 0; i < 25; i++ {
			insertService(t, "Service"+strconv.Itoa(i), "active", 1000)
		}

		// Проверяем, что данные вставились
		var count int
		err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM services WHERE deleted_at IS NULL")
		require.NoError(t, err)
		assert.Equal(t, 25, count, "Должно быть 25 сервисов в БД")

		query := models.BoxList{
			Limit:  0,
			Offset: 0,
			Sort:   "",
			Order:  "",
		}

		t.Logf("Query: %+v", query)

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		t.Logf("Result items count: %d, Total: %d", len(result.Items), result.Total)

		assert.Len(t, result.Items, 20, "Должно вернуть 20 элементов (default limit)")
		assert.Equal(t, 25, result.Total, "Total должно быть 25")
		assert.Equal(t, 20, result.Limit, "Limit должен быть 20")
		assert.Equal(t, 0, result.Offset, "Offset должен быть 0")
	})

	t.Run("pagination - full test", func(t *testing.T) {
		truncateTables()

		// Вставляем 25 сервисов
		for i := 0; i < 25; i++ {
			insertService(t, "Service"+strconv.Itoa(i), "active", 1000)
		}

		// Страница 1: первые 10
		page1, err := boxRepo.List(ctx, models.BoxList{Limit: 10, Offset: 0, Sort: "id", Order: "asc"})
		require.NoError(t, err)
		assert.Len(t, page1.Items, 10)

		// Страница 2: следующие 10
		page2, err := boxRepo.List(ctx, models.BoxList{Limit: 10, Offset: 10, Sort: "id", Order: "asc"})
		require.NoError(t, err)
		assert.Len(t, page2.Items, 10)

		// Страница 3: последние 5
		page3, err := boxRepo.List(ctx, models.BoxList{Limit: 10, Offset: 20, Sort: "id", Order: "asc"})
		require.NoError(t, err)
		assert.Len(t, page3.Items, 5)

		// Проверяем, что ID не пересекаются
		ids1 := extractIDs(page1.Items)
		ids2 := extractIDs(page2.Items)
		ids3 := extractIDs(page3.Items)

		for _, id := range ids1 {
			assert.NotContains(t, ids2, id)
			assert.NotContains(t, ids3, id)
		}
		for _, id := range ids2 {
			assert.NotContains(t, ids3, id)
		}

		// Проверяем, что total правильный
		assert.Equal(t, 25, page1.Total)
		assert.Equal(t, 25, page2.Total)
		assert.Equal(t, 25, page3.Total)
	})

	t.Run("pagination - second page", func(t *testing.T) {
		truncateTables()

		// Вставляем 25 сервисов
		for i := 0; i < 25; i++ {
			insertService(t, "Service"+strconv.Itoa(i), "active", 1000)
		}

		// Первая страница: первые 10
		queryPage1 := models.BoxList{
			Limit:  10,
			Offset: 0,
			Sort:   "id",
			Order:  "asc",
		}
		page1, err := boxRepo.List(ctx, queryPage1)
		require.NoError(t, err)

		// Вторая страница: следующие 10
		queryPage2 := models.BoxList{
			Limit:  10,
			Offset: 10,
			Sort:   "id",
			Order:  "asc",
		}
		page2, err := boxRepo.List(ctx, queryPage2)
		require.NoError(t, err)

		// Проверяем, что страницы не пересекаются
		assert.Len(t, page1.Items, 10)
		assert.Len(t, page2.Items, 10)

		// ID на страницах разные
		page1IDs := make([]int64, len(page1.Items))
		page2IDs := make([]int64, len(page2.Items))
		for i, item := range page1.Items {
			page1IDs[i] = item.ID
		}
		for i, item := range page2.Items {
			page2IDs[i] = item.ID
		}

		for _, id := range page1IDs {
			assert.NotContains(t, page2IDs, id)
		}
	})

	t.Run("pagination - last page with fewer items", func(t *testing.T) {
		truncateTables()

		// Вставляем 25 сервисов
		for i := 0; i < 25; i++ {
			insertService(t, "Service"+strconv.Itoa(i), "active", 1000)
		}

		// Последняя страница: limit=10, offset=20 (должно быть 5 элементов)
		query := models.BoxList{
			Limit:  10,
			Offset: 20,
			Sort:   "id",
			Order:  "asc",
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)

		assert.Len(t, result.Items, 5) // осталось 5 записей
		assert.Equal(t, 25, result.Total)
		assert.Equal(t, 10, result.Limit)
		assert.Equal(t, 20, result.Offset)
	})

	t.Run("pagination - offset beyond total", func(t *testing.T) {
		truncateTables()

		// Вставляем 10 сервисов
		for i := 0; i < 10; i++ {
			insertService(t, "Service"+strconv.Itoa(i), "active", 1000)
		}

		// Offset больше чем всего записей
		query := models.BoxList{
			Limit:  10,
			Offset: 20, // больше 10
			Sort:   "id",
			Order:  "asc",
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)

		assert.Empty(t, result.Items)     // пустой результат
		assert.Equal(t, 10, result.Total) // но total = 10
	})

	t.Run("respects limit and offset parameters", func(t *testing.T) {
		truncateTables()

		// Вставляем 10 сервисов
		for i := 0; i < 10; i++ {
			insertService(t, "Service"+strconv.Itoa(i), "active", 1000)
		}

		query := models.BoxList{
			Limit:  3,
			Offset: 2,
			Sort:   "id",
			Order:  "asc",
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Items, 3)
		assert.Equal(t, 10, result.Total)
		assert.Equal(t, 3, result.Limit)
		assert.Equal(t, 2, result.Offset)
	})

	t.Run("filters by status active", func(t *testing.T) {
		truncateTables()

		insertService(t, "Active Box 1", "active", 1000)
		insertService(t, "Active Box 2", "active", 2000)
		insertService(t, "Inactive Box", "inactive", 1500)

		status := "active"
		query := models.BoxList{
			Status: &status,
			Limit:  10,
			Offset: 0,
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Items, 2)
		assert.Equal(t, 2, result.Total)

		for _, item := range result.Items {
			assert.Equal(t, "active", item.Status)
		}
	})

	t.Run("filters by status inactive", func(t *testing.T) {
		truncateTables()

		insertService(t, "Active Box", "active", 1000)
		insertService(t, "Inactive Box 1", "inactive", 1500)
		insertService(t, "Inactive Box 2", "inactive", 2000)

		status := "inactive"
		query := models.BoxList{
			Status: &status,
			Limit:  10,
			Offset: 0,
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Items, 2)
		assert.Equal(t, 2, result.Total)

		for _, item := range result.Items {
			assert.Equal(t, "inactive", item.Status)
		}
	})

	t.Run("filters by search", func(t *testing.T) {
		truncateTables()

		insertService(t, "Новогодний квест", "active", 1000)
		insertService(t, "Летний квест", "active", 1500)
		insertService(t, "Мастер-класс", "active", 2000)
		insertService(t, "Квест в реальности", "active", 2500)

		// Проверяем, что данные вставились
		var names []string
		err := db.SelectContext(ctx, &names, "SELECT name FROM services WHERE deleted_at IS NULL")
		require.NoError(t, err)
		t.Logf("Сервисы в БД: %v", names)

		search := "квест"
		query := models.BoxList{
			Search: &search,
			Limit:  10,
			Offset: 0,
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		t.Logf("Найденные сервисы: %v", result.Items)

		assert.Len(t, result.Items, 3, "Должно найти 3 сервиса со словом 'квест'")
		assert.Equal(t, 3, result.Total)

		for _, item := range result.Items {
			// Используем Contains с учетом регистра
			assert.True(t,
				strings.Contains(strings.ToLower(item.Name), "квест"),
				"Название '%s' должно содержать 'квест'", item.Name)
		}
	})

	t.Run("filters by status and search together", func(t *testing.T) {
		truncateTables()

		insertService(t, "Активный квест", "active", 1000)
		insertService(t, "Активный мастер-класс", "active", 1500)
		insertService(t, "Неактивный квест", "inactive", 2000)

		status := "active"
		search := "квест"
		query := models.BoxList{
			Status: &status,
			Search: &search,
			Limit:  10,
			Offset: 0,
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Items, 1)
		assert.Equal(t, 1, result.Total)
		assert.Equal(t, "Активный квест", result.Items[0].Name)
	})

	t.Run("sorts by name asc", func(t *testing.T) {
		truncateTables()

		insertService(t, "Бета", "active", 1000)
		insertService(t, "Альфа", "active", 1500)
		insertService(t, "Гамма", "active", 2000)

		query := models.BoxList{
			Limit:  10,
			Offset: 0,
			Sort:   "name",
			Order:  "asc",
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Items, 3)
		assert.Equal(t, "Альфа", result.Items[0].Name)
		assert.Equal(t, "Бета", result.Items[1].Name)
		assert.Equal(t, "Гамма", result.Items[2].Name)
	})

	t.Run("sorts by name desc", func(t *testing.T) {
		truncateTables()

		insertService(t, "Бета", "active", 1000)
		insertService(t, "Альфа", "active", 1500)
		insertService(t, "Гамма", "active", 2000)

		query := models.BoxList{
			Limit:  10,
			Offset: 0,
			Sort:   "name",
			Order:  "desc",
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Items, 3)
		assert.Equal(t, "Гамма", result.Items[0].Name)
		assert.Equal(t, "Бета", result.Items[1].Name)
		assert.Equal(t, "Альфа", result.Items[2].Name)
	})

	t.Run("sorts by created_at", func(t *testing.T) {
		truncateTables()

		id1 := insertService(t, "Первый", "active", 1000)
		time.Sleep(10 * time.Millisecond)
		id2 := insertService(t, "Второй", "active", 1500)
		time.Sleep(10 * time.Millisecond)
		id3 := insertService(t, "Третий", "active", 2000)

		query := models.BoxList{
			Limit:  10,
			Offset: 0,
			Sort:   "created_at",
			Order:  "asc",
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Items, 3)
		assert.Equal(t, id1, result.Items[0].ID)
		assert.Equal(t, id2, result.Items[1].ID)
		assert.Equal(t, id3, result.Items[2].ID)
	})

	t.Run("returns empty result when no services match", func(t *testing.T) {
		truncateTables()

		insertService(t, "Active Box", "active", 1000)

		status := "inactive"
		query := models.BoxList{
			Status: &status,
			Limit:  10,
			Offset: 0,
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Empty(t, result.Items)
		assert.Equal(t, 0, result.Total)
	})

	t.Run("returns services with correct slots", func(t *testing.T) {
		truncateTables()

		slots1 := []struct {
			date, start, end string
		}{
			{"2024-05-01", "10:00", "12:00"},
			{"2024-05-02", "14:00", "16:00"},
		}
		id1 := insertServiceWithSlots(t, "Квест с слотами", "active", 1000, slots1)

		query := models.BoxList{
			Limit:  10,
			Offset: 0,
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		var found *models.Service
		for i := range result.Items {
			if result.Items[i].ID == id1 {
				found = &result.Items[i]
				break
			}
		}
		require.NotNil(t, found)
		assert.Len(t, found.BoxAvailableSlots, 2)
		assert.Equal(t, "2024-05-01", found.BoxAvailableSlots[0].Date)
		assert.Equal(t, "10:00", found.BoxAvailableSlots[0].StartTime)
		assert.Equal(t, "12:00", found.BoxAvailableSlots[0].EndTime)
	})

	t.Run("returns services without slots", func(t *testing.T) {
		truncateTables()

		id1 := insertService(t, "Квест без слотов", "active", 1000)

		query := models.BoxList{
			Limit:  10,
			Offset: 0,
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		var found *models.Service
		for i := range result.Items {
			if result.Items[i].ID == id1 {
				found = &result.Items[i]
				break
			}
		}
		require.NotNil(t, found)
		assert.Empty(t, found.BoxAvailableSlots)
	})

	t.Run("soft-deleted services are excluded", func(t *testing.T) {
		truncateTables()

		id1 := insertService(t, "Активный", "active", 1000)
		id2 := insertService(t, "Удаленный", "active", 1500)

		// Мягко удаляем второй сервис
		err := boxRepo.SoftDeleteService(ctx, id2)
		require.NoError(t, err)

		query := models.BoxList{
			Limit:  10,
			Offset: 0,
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Items, 1)
		assert.Equal(t, id1, result.Items[0].ID)
		assert.Equal(t, 1, result.Total)
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		truncateTables()

		insertService(t, "Тестовый", "active", 1000)

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		query := models.BoxList{
			Limit:  10,
			Offset: 0,
		}

		result, err := boxRepo.List(cancelledCtx, query)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("limit exceeds max returns all available", func(t *testing.T) {
		truncateTables()

		// Вставляем 50 сервисов
		for i := 0; i < 50; i++ {
			insertService(t, "Service"+strconv.Itoa(i), "active", 1000)
		}

		query := models.BoxList{
			Limit:  100, // больше чем есть
			Offset: 0,
		}

		result, err := boxRepo.List(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Items, 50)
		assert.Equal(t, 50, result.Total)
	})
}

func TestGetServices(t *testing.T) {
	_, err := db.ExecContext(context.Background(), "TRUNCATE service_available_slots, services RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	// Вставляем 2 сервиса
	var serviceID1, serviceID2 int64
	err = db.QueryRowContext(context.Background(), `
			INSERT INTO services (name, slug, description, rules, location, price, status, organizer)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id`,
		"Box 1", "box-1", "Описание 1", "Правила 1", "Москва", 1000, "active", "Иван",
	).Scan(&serviceID1)
	require.NoError(t, err)

	err = db.QueryRowContext(context.Background(), `
			INSERT INTO services (name, slug, price, status, organizer)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
		"Box 2", "box-2", 2000, "active", "Петр",
	).Scan(&serviceID2)
	require.NoError(t, err)

	// Слоты для первого сервиса
	_, err = db.ExecContext(context.Background(), `
			INSERT INTO service_available_slots (service_id, slot_date, start_time, end_time)
			VALUES ($1, $2, $3, $4), ($1, $5, $6, $7)`,
		serviceID1,
		"2024-03-25", "09:00", "18:00",
		"2024-03-26", "10:00", "15:00",
	)
	require.NoError(t, err)

	// Вставляем удалённый сервис — не должен попасть в выборку
	_, err = db.ExecContext(context.Background(), `
			INSERT INTO services (name, slug, price, status, deleted_at)
			VALUES ($1, $2, $3, $4, NOW())`,
		"Deleted Box", "deleted-box", 500, "active",
	)
	require.NoError(t, err)

	t.Run("success - returns all non-deleted services", func(t *testing.T) {
		services, err := boxRepo.GetServices(context.Background(), 123)
		require.NoError(t, err)
		require.NotNil(t, services)
		assert.Len(t, services, 2) // удалённый не попал
	})

	t.Run("success - slots correctly mapped", func(t *testing.T) {
		services, err := boxRepo.GetServices(context.Background(), 123)
		require.NoError(t, err)

		// Находим первый сервис
		var box1 *models.Service
		for _, s := range services {
			if s.ID == serviceID1 {
				svc := s
				box1 = &svc
				break
			}
		}

		require.NotNil(t, box1)
		assert.Equal(t, "Box 1", box1.Name)
		assert.Equal(t, "Москва", box1.Location)
		assert.Equal(t, "Иван", box1.Organizer)
		assert.Len(t, box1.BoxAvailableSlots, 2)
		assert.Equal(t, "2024-03-25", box1.BoxAvailableSlots[0].Date)
		assert.Equal(t, "09:00", box1.BoxAvailableSlots[0].StartTime)
	})

	t.Run("success - service without slots", func(t *testing.T) {
		services, err := boxRepo.GetServices(context.Background(), 123)
		require.NoError(t, err)

		var box2 *models.Service
		for _, s := range services {
			if s.ID == serviceID2 {
				svc := s
				box2 = &svc
				break
			}
		}

		require.NotNil(t, box2)
		assert.Empty(t, box2.BoxAvailableSlots)
	})

	t.Run("empty - no services", func(t *testing.T) {
		_, err := db.ExecContext(context.Background(), "TRUNCATE service_available_slots, services RESTART IDENTITY CASCADE")
		require.NoError(t, err)

		services, err := boxRepo.GetServices(context.Background(), 123)
		require.NoError(t, err)
		assert.Empty(t, services)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		services, err := boxRepo.GetServices(ctx, 123)
		assert.Nil(t, services)
		assert.Error(t, err)
	})
}

func TestGetServiceByID(t *testing.T) {
	// Чистим перед тестом
	_, err := db.ExecContext(context.Background(), "TRUNCATE service_available_slots, services RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	// Вставляем тестовые данные
	var serviceID int64
	err = db.QueryRowContext(context.Background(), `
			INSERT INTO services (name, slug, description, rules, location, price, image, status, organizer)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`,
		"Test Box", "test-box", "Описание", "Правила", "Москва", 1000, nil, "active", "Иван",
	).Scan(&serviceID)
	require.NoError(t, err)

	// Вставляем слоты
	_, err = db.ExecContext(context.Background(), `
			INSERT INTO service_available_slots (service_id, slot_date, start_time, end_time)
			VALUES ($1, $2, $3, $4), ($1, $5, $6, $7)`,
		serviceID,
		"2024-03-25", "09:00", "18:00",
		"2024-03-26", "10:00", "15:00",
	)
	require.NoError(t, err)

	t.Run("success - service with slots", func(t *testing.T) {
		svc, err := boxRepo.GetServiceByID(context.Background(), serviceID)
		require.NoError(t, err)
		require.NotNil(t, svc)

		assert.Equal(t, serviceID, svc.ID)
		assert.Equal(t, "Test Box", svc.Name)
		assert.Equal(t, "test-box", svc.Slug)
		assert.Equal(t, "Описание", svc.Description)
		assert.Equal(t, "Москва", svc.Location)
		assert.Equal(t, 1000, svc.Price)
		assert.Equal(t, "active", svc.Status)
		assert.Equal(t, "Иван", svc.Organizer)
		assert.Nil(t, svc.Image)

		require.Len(t, svc.BoxAvailableSlots, 2)
		assert.Equal(t, "2024-03-25", svc.BoxAvailableSlots[0].Date)
		assert.Equal(t, "09:00", svc.BoxAvailableSlots[0].StartTime)
		assert.Equal(t, "18:00", svc.BoxAvailableSlots[0].EndTime)
	})

	t.Run("not found", func(t *testing.T) {
		svc, err := boxRepo.GetServiceByID(context.Background(), 99999)
		assert.Nil(t, svc)
		assert.ErrorIs(t, err, models.ErrBoxSolutionNotFound)
	})

	t.Run("service without slots", func(t *testing.T) {
		// Вставляем сервис без слотов
		var noSlotServiceID int64
		err = db.QueryRowContext(context.Background(), `
					INSERT INTO services (name, slug, price, status, organizer)
					VALUES ($1, $2, $3, $4, $5)
					RETURNING id`,
			"No Slots Box", "no-slots-box", 500, "active", "Петр",
		).Scan(&noSlotServiceID)
		require.NoError(t, err)

		svc, err := boxRepo.GetServiceByID(context.Background(), noSlotServiceID)
		require.NoError(t, err)
		require.NotNil(t, svc)
		assert.Empty(t, svc.BoxAvailableSlots)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // сразу отменяем

		svc, err := boxRepo.GetServiceByID(ctx, serviceID)
		assert.Nil(t, svc)
		assert.Error(t, err)
	})
}

func TestUpdateService(t *testing.T) {
	// Чистим и вставляем свежие данные
	_, err := db.ExecContext(context.Background(), "TRUNCATE service_available_slots, services RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	// Вставляем тестовый сервис
	var serviceID int64
	err = db.QueryRowContext(context.Background(), `
			INSERT INTO services (name, slug, description, rules, location, price, image, status, organizer)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`,
		"Test Box", "test-box", "Описание", "Правила", "Москва", 1000, nil, "active", "Иван",
	).Scan(&serviceID)
	require.NoError(t, err)

	t.Run("success - update all fields", func(t *testing.T) {
		newName := "Updated Box"
		newDesc := "Новое описание"
		newPrice := 2000
		newStatus := "inactive"

		err := boxRepo.UpdateService(context.Background(), serviceID, &models.BoxUpdate{
			Name:        &newName,
			Description: &newDesc,
			Price:       &newPrice,
			Status:      &newStatus,
		})
		require.NoError(t, err)

		// Проверяем что обновилось
		var name, desc, status string
		var price int
		err = db.QueryRowContext(context.Background(),
			`SELECT name, description, price, status FROM services WHERE id = $1`, serviceID,
		).Scan(&name, &desc, &price, &status)
		require.NoError(t, err)

		assert.Equal(t, "Updated Box", name)
		assert.Equal(t, "Новое описание", desc)
		assert.Equal(t, 2000, price)
		assert.Equal(t, "inactive", status)
	})

	t.Run("success - partial update", func(t *testing.T) {
		newName := "Partially Updated"

		err := boxRepo.UpdateService(context.Background(), serviceID, &models.BoxUpdate{
			Name: &newName, // только имя
		})
		require.NoError(t, err)

		// Проверяем что имя обновилось, остальное не тронуто
		var name, status string
		err = db.QueryRowContext(context.Background(),
			`SELECT name, status FROM services WHERE id = $1`, serviceID,
		).Scan(&name, &status)
		require.NoError(t, err)

		assert.Equal(t, "Partially Updated", name)
		assert.Equal(t, "inactive", status) // статус остался из прошлого теста
	})

	t.Run("not found", func(t *testing.T) {
		newName := "Ghost"
		err := boxRepo.UpdateService(context.Background(), 99999, &models.BoxUpdate{
			Name: &newName,
		})
		assert.ErrorIs(t, err, models.ErrBoxSolutionNotFound)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		newName := "Ghost"
		err := boxRepo.UpdateService(ctx, serviceID, &models.BoxUpdate{
			Name: &newName,
		})
		assert.Error(t, err)
	})
}

func TestDeleteServiceSlots(t *testing.T) {
	ctx := context.Background()

	// helper: вставить сервис, вернуть его ID
	insertService := func(t *testing.T, name, slug string) int64 {
		t.Helper()
		var id int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO services (name, slug, price, status, organizer)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			name, slug, 500, "active", "Тест",
		).Scan(&id)
		require.NoError(t, err)
		return id
	}

	// helper: вставить слот
	insertSlot := func(t *testing.T, serviceID int64, date, start, end string) {
		t.Helper()
		_, err := db.ExecContext(ctx, `
			INSERT INTO service_available_slots (service_id, slot_date, start_time, end_time)
			VALUES ($1, $2, $3, $4)`,
			serviceID, date, start, end,
		)
		require.NoError(t, err)
	}

	// helper: посчитать слоты по service_id
	countSlots := func(t *testing.T, serviceID int64) int {
		t.Helper()
		var count int
		err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM service_available_slots WHERE service_id=$1`, serviceID,
		).Scan(&count)
		require.NoError(t, err)
		return count
	}

	t.Run("deletes all slots for service", func(t *testing.T) {
		_, err := db.ExecContext(ctx, "TRUNCATE service_available_slots, services RESTART IDENTITY CASCADE")
		require.NoError(t, err)

		svcID := insertService(t, "Box A", "box-a")
		insertSlot(t, svcID, "2024-04-01", "09:00", "12:00")
		insertSlot(t, svcID, "2024-04-02", "10:00", "14:00")
		insertSlot(t, svcID, "2024-04-03", "11:00", "16:00")

		require.Equal(t, 3, countSlots(t, svcID))

		err = boxRepo.DeleteServiceSlots(ctx, svcID)
		require.NoError(t, err)

		assert.Equal(t, 0, countSlots(t, svcID))
	})

	t.Run("does not delete slots of other services", func(t *testing.T) {
		_, err := db.ExecContext(ctx, "TRUNCATE service_available_slots, services RESTART IDENTITY CASCADE")
		require.NoError(t, err)

		svcA := insertService(t, "Box A", "box-a")
		svcB := insertService(t, "Box B", "box-b")
		insertSlot(t, svcA, "2024-04-01", "09:00", "12:00")
		insertSlot(t, svcB, "2024-04-01", "09:00", "12:00")
		insertSlot(t, svcB, "2024-04-02", "10:00", "14:00")

		err = boxRepo.DeleteServiceSlots(ctx, svcA)
		require.NoError(t, err)

		assert.Equal(t, 0, countSlots(t, svcA))
		assert.Equal(t, 2, countSlots(t, svcB)) // чужие слоты не тронуты
	})

	t.Run("no-op on service with no slots", func(t *testing.T) {
		_, err := db.ExecContext(ctx, "TRUNCATE service_available_slots, services RESTART IDENTITY CASCADE")
		require.NoError(t, err)

		svcID := insertService(t, "Box Empty", "box-empty")
		require.Equal(t, 0, countSlots(t, svcID))

		err = boxRepo.DeleteServiceSlots(ctx, svcID)
		require.NoError(t, err) // DELETE без строк — не ошибка
	})

	t.Run("no-op on non-existent service id", func(t *testing.T) {
		err := boxRepo.DeleteServiceSlots(ctx, 999999999)
		require.NoError(t, err) // аналогично — просто 0 строк затронуто
	})

	t.Run("cancelled context", func(t *testing.T) {
		_, err := db.ExecContext(ctx, "TRUNCATE service_available_slots, services RESTART IDENTITY CASCADE")
		require.NoError(t, err)

		svcID := insertService(t, "Box Ctx", "box-ctx")
		insertSlot(t, svcID, "2024-04-01", "09:00", "12:00")

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		err = boxRepo.DeleteServiceSlots(cancelledCtx, svcID)
		assert.Error(t, err)

		// слот должен остаться — транзакция не прошла
		assert.Equal(t, 1, countSlots(t, svcID))
	})
}

func TestUpdateServiceSlots(t *testing.T) {
	ctx := context.Background()

	insertService := func(t *testing.T) int64 {
		t.Helper()
		suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
		var id int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO services (name, slug, price, status, organizer)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			"Box "+suffix, "box-"+suffix, 500, "active", "Тест",
		).Scan(&id)
		require.NoError(t, err)
		return id
	}

	countSlots := func(t *testing.T, serviceID int64) int {
		t.Helper()
		var count int
		err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM service_available_slots WHERE service_id=$1`, serviceID,
		).Scan(&count)
		require.NoError(t, err)
		return count
	}

	parseSlots := func(t *testing.T, serviceID int64) []struct {
		Date      string
		StartTime string
		EndTime   string
	} {
		t.Helper()
		rows, err := db.QueryContext(ctx, `
			SELECT slot_date::text, start_time::text, end_time::text
			FROM service_available_slots
			WHERE service_id=$1
			ORDER BY slot_date, start_time`, serviceID,
		)
		require.NoError(t, err)
		defer rows.Close()

		var result []struct {
			Date      string
			StartTime string
			EndTime   string
		}
		for rows.Next() {
			var s struct {
				Date      string
				StartTime string
				EndTime   string
			}
			require.NoError(t, rows.Scan(&s.Date, &s.StartTime, &s.EndTime))
			result = append(result, s)
		}
		require.NoError(t, rows.Err())
		return result
	}

	makeSlots := func(dates, starts, ends []string) *models.BoxNewSlots {
		parsedDates := make([]time.Time, len(dates))
		parsedStarts := make([]time.Time, len(starts))
		parsedEnds := make([]time.Time, len(ends))
		for i := range dates {
			parsedDates[i], _ = time.Parse("2006-01-02", dates[i])
			parsedStarts[i], _ = time.Parse("15:04", starts[i])
			parsedEnds[i], _ = time.Parse("15:04", ends[i])
		}
		return &models.BoxNewSlots{
			Date:      parsedDates,
			StartTime: parsedStarts,
			EndTime:   parsedEnds,
		}
	}

	t.Run("inserts slots for service", func(t *testing.T) {
		svcID := insertService(t)

		slots := makeSlots(
			[]string{"2024-05-01", "2024-05-02"},
			[]string{"09:00", "10:00"},
			[]string{"12:00", "14:00"},
		)

		err := boxRepo.UpdateServiceSlots(ctx, svcID, slots)
		require.NoError(t, err)

		require.Equal(t, 2, countSlots(t, svcID))

		got := parseSlots(t, svcID)
		assert.Equal(t, "2024-05-01", got[0].Date)
		assert.Equal(t, "09:00:00", got[0].StartTime)
		assert.Equal(t, "12:00:00", got[0].EndTime)
		assert.Equal(t, "2024-05-02", got[1].Date)
	})

	t.Run("inserts single slot", func(t *testing.T) {
		svcID := insertService(t)

		slots := makeSlots(
			[]string{"2024-06-01"},
			[]string{"08:00"},
			[]string{"17:00"},
		)

		err := boxRepo.UpdateServiceSlots(ctx, svcID, slots)
		require.NoError(t, err)
		assert.Equal(t, 1, countSlots(t, svcID))
	})

	t.Run("does not affect other services", func(t *testing.T) {
		svcA := insertService(t)
		svcB := insertService(t)

		slots := makeSlots(
			[]string{"2024-05-01"},
			[]string{"09:00"},
			[]string{"12:00"},
		)

		err := boxRepo.UpdateServiceSlots(ctx, svcA, slots)
		require.NoError(t, err)

		assert.Equal(t, 1, countSlots(t, svcA))
		assert.Equal(t, 0, countSlots(t, svcB)) // чужие не тронуты
	})

	t.Run("duplicate slot returns error", func(t *testing.T) {
		svcID := insertService(t)

		slots := makeSlots(
			[]string{"2024-05-01"},
			[]string{"09:00"},
			[]string{"12:00"},
		)

		err := boxRepo.UpdateServiceSlots(ctx, svcID, slots)
		require.NoError(t, err)

		// вставляем тот же слот повторно — нарушение uq_available_slots_service_date
		err = boxRepo.UpdateServiceSlots(ctx, svcID, slots)
		require.Error(t, err)
	})

	t.Run("cancelled context", func(t *testing.T) {
		svcID := insertService(t)

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		slots := makeSlots(
			[]string{"2024-05-01"},
			[]string{"09:00"},
			[]string{"12:00"},
		)

		err := boxRepo.UpdateServiceSlots(cancelledCtx, svcID, slots)
		assert.Error(t, err)
		assert.Equal(t, 0, countSlots(t, svcID))
	})
}

func TestSoftDeleteService(t *testing.T) {
	ctx := context.Background()

	insertService := func(t *testing.T) int64 {
		t.Helper()
		suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
		var id int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO services (name, slug, price, status, organizer)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			"Service "+suffix, "service-"+suffix, 1000, "active", "Тест",
		).Scan(&id)
		require.NoError(t, err)
		return id
	}

	isDeleted := func(t *testing.T, serviceID int64) bool {
		t.Helper()
		var deletedAt sql.NullTime
		err := db.QueryRowContext(ctx,
			`SELECT deleted_at FROM services WHERE id=$1`, serviceID,
		).Scan(&deletedAt)
		require.NoError(t, err)
		return deletedAt.Valid
	}

	t.Run("soft deletes existing service", func(t *testing.T) {
		svcID := insertService(t)

		err := boxRepo.SoftDeleteService(ctx, svcID)
		require.NoError(t, err)

		assert.True(t, isDeleted(t, svcID))
	})

	t.Run("idempotent - deleting already deleted service returns nil", func(t *testing.T) {
		svcID := insertService(t)

		// первый раз
		err := boxRepo.SoftDeleteService(ctx, svcID)
		require.NoError(t, err)

		// второй раз — не падаем
		err = boxRepo.SoftDeleteService(ctx, svcID)
		require.NoError(t, err) // не ошибка, всё ок

		assert.True(t, isDeleted(t, svcID))
	})

	t.Run("idempotent - deleting non-existent service returns nil", func(t *testing.T) {
		err := boxRepo.SoftDeleteService(ctx, 99999999)
		require.NoError(t, err) // просто успех
		// ничего не проверяем — такого ID нет, и ладно
	})

	t.Run("does not affect other services", func(t *testing.T) {
		svc1 := insertService(t)
		svc2 := insertService(t)

		err := boxRepo.SoftDeleteService(ctx, svc1)
		require.NoError(t, err)

		assert.True(t, isDeleted(t, svc1))
		assert.False(t, isDeleted(t, svc2))
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		svcID := insertService(t)

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		err := boxRepo.SoftDeleteService(cancelledCtx, svcID)
		assert.Error(t, err)                 // контекст отменён — ошибка
		assert.False(t, isDeleted(t, svcID)) // данные не тронуты
	})
}

func TestUpdateServiceStatus(t *testing.T) {
	ctx := context.Background()

	insertService := func(t *testing.T, status string) int64 {
		t.Helper()
		suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
		var id int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO services (name, slug, price, status, organizer)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			"Service "+suffix, "service-"+suffix, 1000, status, "Тест",
		).Scan(&id)
		require.NoError(t, err)
		return id
	}

	getServiceStatus := func(t *testing.T, serviceID int64) (string, bool) {
		t.Helper()
		var status string
		var deletedAt sql.NullTime
		err := db.QueryRowContext(ctx,
			`SELECT status, deleted_at FROM services WHERE id=$1`, serviceID,
		).Scan(&status, &deletedAt)
		require.NoError(t, err)
		return status, deletedAt.Valid
	}

	t.Run("updates status of existing active service", func(t *testing.T) {
		svcID := insertService(t, "active")

		result, err := boxRepo.UpdateServiceStatus(ctx, svcID, "inactive")
		require.NoError(t, err)

		assert.Equal(t, svcID, result.ID)
		assert.Equal(t, "inactive", result.Status)
		assert.NotZero(t, result.UpdatedAt)

		status, deleted := getServiceStatus(t, svcID)
		assert.Equal(t, "inactive", status)
		assert.False(t, deleted)
	})

	t.Run("returns error for non-existent service", func(t *testing.T) {
		_, err := boxRepo.UpdateServiceStatus(ctx, 99999999, "active")
		assert.ErrorIs(t, err, models.ErrBoxSolutionNotFound)
	})

	t.Run("returns error for soft-deleted service", func(t *testing.T) {
		svcID := insertService(t, "active")

		// мягко удаляем
		err := boxRepo.SoftDeleteService(ctx, svcID)
		require.NoError(t, err)

		// пробуем обновить статус
		_, err = boxRepo.UpdateServiceStatus(ctx, svcID, "inactive")
		assert.ErrorIs(t, err, models.ErrBoxSolutionNotFound)
	})

	t.Run("does not affect other services", func(t *testing.T) {
		svcA := insertService(t, "active")
		svcB := insertService(t, "active")

		_, err := boxRepo.UpdateServiceStatus(ctx, svcA, "inactive")
		require.NoError(t, err)

		statusA, _ := getServiceStatus(t, svcA)
		statusB, _ := getServiceStatus(t, svcB)

		assert.Equal(t, "inactive", statusA)
		assert.Equal(t, "active", statusB)
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		svcID := insertService(t, "active")

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err := boxRepo.UpdateServiceStatus(cancelledCtx, svcID, "inactive")
		assert.Error(t, err)

		status, _ := getServiceStatus(t, svcID)
		assert.Equal(t, "active", status) // не изменился
	})

	t.Run("updates updated_at timestamp", func(t *testing.T) {
		svcID := insertService(t, "active")

		var oldUpdatedAt time.Time
		err := db.QueryRowContext(ctx,
			`SELECT updated_at FROM services WHERE id=$1`, svcID,
		).Scan(&oldUpdatedAt)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond) // гарантируем разницу

		_, err = boxRepo.UpdateServiceStatus(ctx, svcID, "inactive")
		require.NoError(t, err)

		var newUpdatedAt time.Time
		err = db.QueryRowContext(ctx,
			`SELECT updated_at FROM services WHERE id=$1`, svcID,
		).Scan(&newUpdatedAt)
		require.NoError(t, err)

		assert.True(t, newUpdatedAt.After(oldUpdatedAt),
			"updated_at should change: old=%v, new=%v", oldUpdatedAt, newUpdatedAt)
	})
}

func TestGetServicesByStatus(t *testing.T) {
	ctx := context.Background()

	// helper для очистки таблиц
	truncateTables := func() {
		_, err := db.ExecContext(ctx, "TRUNCATE services, service_available_slots RESTART IDENTITY CASCADE")
		require.NoError(t, err)
	}

	// helper для вставки сервиса со слотами
	insertServiceWithSlots := func(t *testing.T, status string, slotDates ...string) int64 {
		t.Helper()
		suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
		var id int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO services (name, slug, price, status, organizer)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			"Service "+suffix, "service-"+suffix, 1000, status, "Тест",
		).Scan(&id)
		require.NoError(t, err)

		for _, date := range slotDates {
			_, err = db.ExecContext(ctx, `
				INSERT INTO service_available_slots (service_id, slot_date, start_time, end_time)
				VALUES ($1, $2, '09:00', '17:00')`,
				id, date,
			)
			require.NoError(t, err)
		}
		return id
	}

	t.Run("returns all services (active and inactive) when status is nil", func(t *testing.T) {
		truncateTables()

		svc1 := insertServiceWithSlots(t, "active", "2024-05-01")
		svc2 := insertServiceWithSlots(t, "active")
		svc3 := insertServiceWithSlots(t, "inactive")

		services, err := boxRepo.GetServicesByStatus(ctx, nil)
		require.NoError(t, err)

		ids := make([]int64, len(services))
		for i, s := range services {
			ids[i] = s.ID
		}

		assert.Contains(t, ids, svc1)
		assert.Contains(t, ids, svc2)
		assert.Contains(t, ids, svc3)
		assert.Len(t, services, 3)
	})

	t.Run("filters by active status when provided", func(t *testing.T) {
		truncateTables()

		svcActive := insertServiceWithSlots(t, "active")
		insertServiceWithSlots(t, "inactive")

		status := models.StatusActive
		services, err := boxRepo.GetServicesByStatus(ctx, &status)
		require.NoError(t, err)

		assert.Len(t, services, 1)
		assert.Equal(t, svcActive, services[0].ID)
		assert.Equal(t, "active", services[0].Status)
	})

	t.Run("filters by inactive status when provided", func(t *testing.T) {
		truncateTables()

		svcInactive := insertServiceWithSlots(t, "inactive")
		insertServiceWithSlots(t, "active")

		status := models.StatusInactive
		services, err := boxRepo.GetServicesByStatus(ctx, &status)
		require.NoError(t, err)

		assert.Len(t, services, 1)
		assert.Equal(t, svcInactive, services[0].ID)
		assert.Equal(t, "inactive", services[0].Status)
	})

	t.Run("returns services with correct slots count", func(t *testing.T) {
		truncateTables()

		svcID := insertServiceWithSlots(t, "active", "2024-05-01", "2024-05-02", "2024-05-03")

		services, err := boxRepo.GetServicesByStatus(ctx, nil)
		require.NoError(t, err)

		var found *models.Service
		for i := range services {
			if services[i].ID == svcID {
				found = &services[i]
				break
			}
		}
		require.NotNil(t, found)
		assert.Len(t, found.BoxAvailableSlots, 3)
	})

	t.Run("returns empty slice (not nil) for services without slots", func(t *testing.T) {
		truncateTables()

		svcID := insertServiceWithSlots(t, "active")

		services, err := boxRepo.GetServicesByStatus(ctx, nil)
		require.NoError(t, err)

		var found *models.Service
		for i := range services {
			if services[i].ID == svcID {
				found = &services[i]
				break
			}
		}
		require.NotNil(t, found)
		assert.NotNil(t, found.BoxAvailableSlots)
		assert.Empty(t, found.BoxAvailableSlots)
	})

	t.Run("returns empty slice when no services match status", func(t *testing.T) {
		truncateTables()

		insertServiceWithSlots(t, "active")

		status := models.StatusInactive
		services, err := boxRepo.GetServicesByStatus(ctx, &status)
		require.NoError(t, err)
		assert.Empty(t, services)
		assert.NotNil(t, services)
	})

	t.Run("returns empty slice when no services exist", func(t *testing.T) {
		truncateTables()

		services, err := boxRepo.GetServicesByStatus(ctx, nil)
		require.NoError(t, err)
		assert.Empty(t, services)
		assert.NotNil(t, services)
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		truncateTables()

		insertServiceWithSlots(t, "active")

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		services, err := boxRepo.GetServicesByStatus(cancelledCtx, nil)
		assert.Error(t, err)
		assert.Nil(t, services)
	})

	t.Run("soft-deleted services are excluded", func(t *testing.T) {
		truncateTables()

		svcID := insertServiceWithSlots(t, "active")

		err := boxRepo.SoftDeleteService(ctx, svcID)
		require.NoError(t, err)

		services, err := boxRepo.GetServicesByStatus(ctx, nil)
		require.NoError(t, err)

		for _, s := range services {
			assert.NotEqual(t, svcID, s.ID)
		}
	})

	t.Run("slot dates are formatted as strings (YYYY-MM-DD)", func(t *testing.T) {
		truncateTables()

		svcID := insertServiceWithSlots(t, "active", "2024-12-25")

		services, err := boxRepo.GetServicesByStatus(ctx, nil)
		require.NoError(t, err)

		var found *models.Service
		for i := range services {
			if services[i].ID == svcID {
				found = &services[i]
				break
			}
		}
		require.NotNil(t, found)
		require.Len(t, found.BoxAvailableSlots, 1)

		assert.Equal(t, "2024-12-25", found.BoxAvailableSlots[0].Date)
		assert.IsType(t, "", found.BoxAvailableSlots[0].Date)
	})

	t.Run("multiple slots for same service are ordered correctly", func(t *testing.T) {
		truncateTables()

		svcID := insertServiceWithSlots(t, "active", "2024-05-03", "2024-05-01", "2024-05-02")

		services, err := boxRepo.GetServicesByStatus(ctx, nil)
		require.NoError(t, err)

		var found *models.Service
		for i := range services {
			if services[i].ID == svcID {
				found = &services[i]
				break
			}
		}
		require.NotNil(t, found)
		require.Len(t, found.BoxAvailableSlots, 3)

		assert.Equal(t, "2024-05-01", found.BoxAvailableSlots[0].Date)
		assert.Equal(t, "2024-05-02", found.BoxAvailableSlots[1].Date)
		assert.Equal(t, "2024-05-03", found.BoxAvailableSlots[2].Date)
	})

	t.Run("multiple services with slots are returned correctly", func(t *testing.T) {
		truncateTables()

		svc1 := insertServiceWithSlots(t, "active", "2024-05-01")
		svc2 := insertServiceWithSlots(t, "active", "2024-05-02", "2024-05-03")
		svc3 := insertServiceWithSlots(t, "inactive")

		services, err := boxRepo.GetServicesByStatus(ctx, nil)
		require.NoError(t, err)

		assert.Len(t, services, 3)

		for _, svc := range services {
			switch svc.ID {
			case svc1:
				assert.Len(t, svc.BoxAvailableSlots, 1)
			case svc2:
				assert.Len(t, svc.BoxAvailableSlots, 2)
			case svc3:
				assert.Empty(t, svc.BoxAvailableSlots)
			default:
				t.Errorf("unexpected service ID: %d", svc.ID)
			}
		}
	})
}
