package response

type UserListResponse struct {
	List  []string `json:"list"`
	Total int      `json:"total"`
}
