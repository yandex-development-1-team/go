package service

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/signintech/gopdf"
	"github.com/yandex-development-1-team/go/internal/models"
)

//go:embed fonts/Roboto-Regular.ttf
var fontBytes []byte

const (
	pageWidth    = 210.0
	pageHeight   = 297.0
	marginLeft   = 20.0
	marginRight  = 20.0
	marginBottom = 20.0
	contentWidth = pageWidth - marginLeft - marginRight
)

func (s *APIBoxService) generatePDF(services []models.Service) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	if err := pdf.AddTTFFontData("roboto", fontBytes); err != nil {
		return nil, fmt.Errorf("load font: %w", err)
	}

	for _, svc := range services {
		pdf.AddPage()
		y := 20.0

		// Заголовок
		y = writeText(&pdf, svc.Name, "roboto", 16, marginLeft, y, contentWidth)
		y += 6

		// Основные поля
		fields := []struct{ label, value string }{
			{"Статус", svc.Status},
			{"Цена", fmt.Sprintf("%d руб.", svc.Price)},
			{"Место", svc.Location},
			{"Организатор", svc.Organizer},
		}
		for _, f := range fields {
			if f.value == "" {
				continue
			}
			y = checkNewPage(&pdf, y, 10)
			y = writeText(&pdf, f.label+": "+f.value, "roboto", 12, marginLeft, y, contentWidth)
			y += 2
		}

		// Описание
		if svc.Description != "" {
			y += 4
			y = checkNewPage(&pdf, y, 10)
			y = writeText(&pdf, "Описание:", "roboto", 12, marginLeft, y, contentWidth)
			y += 2
			y = writeWrappedText(&pdf, svc.Description, "roboto", 11, marginLeft, y, contentWidth)
			y += 4
		}

		// Правила
		if svc.Rules != "" {
			y += 4
			y = checkNewPage(&pdf, y, 10)
			y = writeText(&pdf, "Правила:", "roboto", 12, marginLeft, y, contentWidth)
			y += 2
			y = writeWrappedText(&pdf, svc.Rules, "roboto", 11, marginLeft, y, contentWidth)
			y += 4
		}

		// Слоты
		if len(svc.BoxAvailableSlots) > 0 {
			y += 4
			y = checkNewPage(&pdf, y, 10)
			y = writeText(&pdf, "Доступные слоты:", "roboto", 13, marginLeft, y, contentWidth)
			y += 4

			for _, slot := range svc.BoxAvailableSlots {
				y = checkNewPage(&pdf, y, 8)
				line := fmt.Sprintf("  %s   %s — %s", slot.Date, slot.StartTime, slot.EndTime)
				y = writeText(&pdf, line, "roboto", 11, marginLeft, y, contentWidth)
				y += 2
			}
		}
	}

	var buf bytes.Buffer
	if _, err := pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}

	return buf.Bytes(), nil
}

// checkNewPage добавляет новую страницу если не хватает места
func checkNewPage(pdf *gopdf.GoPdf, y, needed float64) float64 {
	if y+needed > pageHeight-marginBottom {
		pdf.AddPage()
		return 20.0
	}
	return y
}

// writeText пишет одну строку и возвращает новый y
func writeText(pdf *gopdf.GoPdf, text, font string, size int, x, y, width float64) float64 {
	_ = pdf.SetFont(font, "", size)
	pdf.SetXY(x, y)
	_ = pdf.Cell(&gopdf.Rect{W: width, H: float64(size) * 1.2}, text)
	return y + float64(size)*1.4
}

// writeWrappedText пишет текст с переносом по словам
func writeWrappedText(pdf *gopdf.GoPdf, text, font string, size int, x, y, width float64) float64 {
	_ = pdf.SetFont(font, "", size)
	lineHeight := float64(size) * 1.4

	words := strings.Fields(text)
	line := ""

	for _, word := range words {
		test := line
		if test != "" {
			test += " "
		}
		test += word

		w, _ := pdf.MeasureTextWidth(test)
		if w > width && line != "" {
			y = checkNewPage(pdf, y, lineHeight)
			pdf.SetXY(x, y)
			_ = pdf.Cell(&gopdf.Rect{W: width, H: lineHeight}, line)
			y += lineHeight
			line = word
		} else {
			line = test
		}
	}

	// Последняя строка
	if line != "" {
		y = checkNewPage(pdf, y, lineHeight)
		pdf.SetXY(x, y)
		_ = pdf.Cell(&gopdf.Rect{W: width, H: lineHeight}, line)
		y += lineHeight
	}

	return y
}

func (s *APIBoxService) generateCSV(services []models.Service) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Заголовки
	if err := w.Write([]string{
		"ID", "Название", "Статус", "Цена", "Место", "Организатор", "Описание", "Правила", "Дата", "Начало", "Конец",
	}); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, svc := range services {
		// Если слотов нет — одна строка без слотов
		if len(svc.BoxAvailableSlots) == 0 {
			if err := w.Write([]string{
				fmt.Sprintf("%d", svc.ID),
				svc.Name,
				svc.Status,
				fmt.Sprintf("%d", svc.Price),
				svc.Location,
				svc.Organizer,
				svc.Description,
				svc.Rules,
				"", "", "",
			}); err != nil {
				return nil, fmt.Errorf("write csv row: %w", err)
			}
			continue
		}

		// Одна строка на каждый слот
		for _, slot := range svc.BoxAvailableSlots {
			if err := w.Write([]string{
				fmt.Sprintf("%d", svc.ID),
				svc.Name,
				svc.Status,
				fmt.Sprintf("%d", svc.Price),
				svc.Location,
				svc.Organizer,
				svc.Description,
				svc.Rules,
				slot.Date,
				slot.StartTime,
				slot.EndTime,
			}); err != nil {
				return nil, fmt.Errorf("write csv row: %w", err)
			}
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return buf.Bytes(), nil
}
