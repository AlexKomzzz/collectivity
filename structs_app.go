package app

type User struct {
	Id         int    `json:"-"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name" db:"first_name"`
	LastName   string `json:"last_name" db:"last_name"`
	MiddleName string `json:"middle_name" db:"middle_name"`
	Email      string `json:"email" db:"email" binding:"required"`
	Password   string `json:"password" db:"password_hash" binding:"required"`
}

type UserGoogle struct {
	Id        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"name"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
}

type UserYandex struct {
	Id        string `json:"id"`
	Email     string `json:"default_email"`
	FullName  string `json:"real_name"` // ФИ
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	// Phone string `json:"number"`
}

/*type Message struct {
	Id       int    `json:"id" db:"id"`
	Date     string `json:"date" db:"date"`
	Username string `json:"username" db:"username"`
	Body     string `json:"message" db:"message"`
}

type GroupChat struct {
	Id           int    `json:"-" db:"id"`
	Title        string `json:"title" db:"title"`
	Admin        int    `json:"admin" db:"admin"`
	Participants []int  `json:"participants"`
}
*/
