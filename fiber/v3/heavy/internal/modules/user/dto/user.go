package dto

type CreateUserRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	Status string `json:"status"`
}

type UpdateUserRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	Status string `json:"status"`
}

type ListUsersRequest struct {
	PageNum  int    `json:"page_num"`
	PageSize int    `json:"page_size"`
	Keyword  string `json:"keyword"`
}
