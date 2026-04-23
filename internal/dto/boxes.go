package dto

import (
	"database/sql"
	"time"
)

type BoxListQuery struct {
	Status *string `form:"status" binding:"omitempty,oneof=active inactive"`
	Search *string `form:"search"`
	Limit  int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int     `form:"offset" binding:"omitempty,min=0"`
	Sort   string  `form:"sort" binding:"omitempty,oneof=name created_at updated_at"`
	Order  string  `form:"order" binding:"omitempty,oneof=asc desc"`
}

type BoxListResponse struct {
	Items      []*BoxDetailResponse `json:"items"`
	Pagination Pagination           `json:"pagination"`
}

type BoxDetailRequest struct {
	ID                int64              `json:"id"`
	Name              string             `json:"name" binding:"required,min=1"`
	Slug              string             `json:"slug"`
	Description       string             `json:"description"`
	Rules             string             `json:"rules"`
	BoxAvailableSlots []BoxAvailableSlot `json:"slots"`
	Location          string             `json:"location"`
	Price             int                `json:"price" binding:"required,gt=0"`
	Image             *string            `json:"image"`
	Status            string             `json:"status" binding:"required,oneof=active inactive"`
	Organizer         string             `json:"organizer"`
}

type BoxDetail struct {
	ID                int64              `json:"id"`
	Name              string             `json:"name" binding:"required,min=1"`
	Slug              string             `json:"slug"`
	Description       string             `json:"description"`
	Rules             string             `json:"rules"`
	BoxAvailableSlots []BoxAvailableSlot `json:"slots"`
	Location          string             `json:"location"`
	Price             int                `json:"price" binding:"required,gt=0"`
	Image             *string            `json:"image"`
	Status            string             `json:"status"`
	Organizer         string             `json:"organizer"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type BoxDetailResponse struct {
	ID                int64              `json:"id"`
	Name              string             `json:"name"`
	Slug              string             `json:"slug"`
	Description       string             `json:"description"`
	Rules             string             `json:"rules"`
	BoxAvailableSlots []BoxAvailableSlot `json:"slots"`
	Location          string             `json:"location"`
	Price             int                `json:"price"`
	Image             *string            `json:"image"`
	Status            string             `json:"status"`
	Organizer         string             `json:"organizer"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type BoxAvailableSlot struct {
	Date      string `json:"date"       binding:"required,datetime=2006-01-02"`
	StartTime string `json:"time_from"  binding:"required,datetime=15:04"`
	EndTime   string `json:"time_to"    binding:"required,datetime=15:04"`
}

type BoxCreateRequest struct {
	Name        *string            `json:"name" binding:"required"`
	Description *string            `json:"description,omitempty"`
	Rules       *string            `json:"rules,omitempty"`
	Location    *string            `json:"location,omitempty"`
	Price       *int               `json:"price" binding:"required,gt=0"`
	Image       *string            `json:"image,omitempty" binding:"omitempty,url"`
	Status      *string            `json:"status" binding:"required,oneof=active inactive"`
	Organizer   *string            `json:"organizer,omitempty"`
	Slots       []BoxAvailableSlot `json:"slots,omitempty"`
}

type BoxUpdateRequest struct {
	// ID          int64              `json:"id"            binding:"required,min=1"`
	Name        *string            `json:"name"          binding:"omitempty,min=1,max=255"`
	Description *string            `json:"description"   binding:"omitempty,max=1000"`
	Rules       *string            `json:"rules"         binding:"omitempty,max=1000"`
	Slots       []BoxAvailableSlot `json:"slots"         binding:"omitempty,dive"`
	Location    *string            `json:"location"      binding:"omitempty,max=255"`
	Price       *int               `json:"price"         binding:"omitempty,min=0"`
	Image       *string            `json:"image"         binding:"omitempty,max=500"`
	Status      *string            `json:"status"        binding:"required,oneof=active inactive"`
	Organizer   *string            `json:"organizer"     binding:"omitempty,max=255"`
}

type BoxUpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

type BoxStatusResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
}

type BoxUpdateStatusResult struct {
	ID        int64     `db:"id"`
	Status    string    `db:"status"`
	UpdatedAt time.Time `db:"updated_at"`
}

type BoxRaw struct {
	ID          int64        `db:"id"`
	Name        string       `db:"name"`
	Slug        string       `db:"slug"`
	Description *string      `db:"description"`
	Rules       *string      `db:"rules"`
	Location    *string      `db:"location"`
	Price       int          `db:"price"`
	Image       *string      `db:"image"`
	Status      string       `db:"status"`
	Organizer   *string      `db:"organizer"`
	CreatedBy   int64        `db:"created_by"`
	SlotDate    sql.NullTime `db:"slot_date"`
	StartTime   sql.NullTime `db:"start_time"`
	EndTime     sql.NullTime `db:"end_time"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

type BoxExportRequest struct {
	Status *string `json:"status" binding:"omitempty,oneof=active inactive"`
	Format *string `json:"format"`
}
