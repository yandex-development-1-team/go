package models

type StaffAdminCreate struct {
	Name         string
	Email        string
	Role         string
	Status       string
	Permissions  []string
	TelegramNick *string
	InviteToken  string
}

type StaffAdminUpdate struct {
	Name         *string
	Email        *string
	Role         *string
	Status       *string
	Permissions  *[]string
	TelegramNick *string
}
