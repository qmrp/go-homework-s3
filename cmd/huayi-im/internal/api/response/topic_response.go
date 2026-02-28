package response

type TopicResponse struct {
	Topic string `json:"topic"`
}

type TopicListResponse struct {
	List  []TopicResponse `json:"topics"`
	Total int             `json:"total"`
}
