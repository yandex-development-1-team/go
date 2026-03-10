-- +goose Up
INSERT INTO
	special_projects (id, title, description, image, is_active_in_bot)
VALUES
	(
		1,
		'AI Assistant Development',
		'Разработка ИИ-помощника для автоматизации задач',
		'https://example.com/ai-project.jpg',
		true
	),
	(
		2,
		'Mobile App Redesign',
		'Редизайн мобильного приложения с новым UX/UI',
		'https://example.com/mobile-app.jpg',
		false
	),
	(
		3,
		'Data Analytics Platform',
		'Платформа для анализа больших данных',
		NULL,
		true
	),
	(
		4,
		'E-commerce Integration',
		'Интеграция с популярными e-commerce платформами',
		'https://example.com/ecommerce.jpg',
		false
	),
	(
		5,
		'Security Audit Tool',
		'Инструмент для автоматического аудита безопасности',
		NULL,
		false
	);

-- +goose Down
DELETE FROM special_projects
WHERE
	id IN (1, 2, 3, 4, 5);