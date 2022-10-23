package app

type User struct {
	Id        int    `json:"-"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type UserGoogle struct {
	Id         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

type UserYandex struct {
	Id        string `json:"id"`
	Email     string `json:"default_email"`
	RealName  string `json:"real_name"` // ФИ
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
