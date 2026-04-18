package models

import "time"

type Application struct {
	ID           int64
	Status       string
	ManagerID    int64
	ManagerName  string
	FormAnswerId string
	CustomerName string
	ContactInfo  string
	Description  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ApplicationFilter struct {
	Status       string
	ManagerID    int64
	CustomerName string
	Limit        int
	Offset       int
}

type ApplicationList struct {
	Items  []Application
	Total  int
	Limit  int
	Offset int
}
