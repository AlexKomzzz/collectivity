package app

type User struct {
	Id       int    `json:"-"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserGoogle struct {
	Id         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

type UserYandex struct {
	Id         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
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
