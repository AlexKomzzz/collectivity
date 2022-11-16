package app

type User struct {
	Id         int    `json:"id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name" db:"first_name"`
	LastName   string `json:"last_name" db:"last_name"`
	MiddleName string `json:"middle_name" db:"middle_name"`
	Email      string `json:"email" db:"email" binding:"required"`
	Password   string `json:"password" db:"password_hash" binding:"required"`
	Debt       string ` db:"debt"`
}

type UserGoogle struct {
	Id        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"name"`        // ФИ
	FirstName string `json:"given_name"`  // имя
	LastName  string `json:"family_name"` // фамилия
}

type UserYandex struct {
	Id        string `json:"id"`
	Email     string `json:"default_email"`
	FullName  string `json:"real_name"` // ФИ
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	// Phone string `json:"number"`
}
