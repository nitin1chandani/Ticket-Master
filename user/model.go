package user

type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	RoleID       int    `json:"roleID"`
	PasswordHash string `json:"-"`
	CreatedBy    string `json:"created_by"`
	CreatedAt    string `json:"created_at"`
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
