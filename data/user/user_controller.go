package user

type UserController struct {
	repo *SQLiteRepository
}

func NewUserController(repo *SQLiteRepository) *UserController {
	return &UserController{
		repo: repo,
	}
}
