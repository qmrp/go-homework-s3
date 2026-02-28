package response

type UserResponse struct {
	Username string `json:"username"`
}

type UserListResponse struct {
	List  []UserResponse `json:"list"`
	Total int            `json:"total"`
}

type UserDetailResponse struct {
	Username string `json:"username"`
	Sid      string `json:"sid"`
}
