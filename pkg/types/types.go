package types

type Resp struct {
	Success bool `json:"success"`
}

type CommonResponse struct {
	Data       Resp   `json:"data"`
	TargetUser string `json:"targetUser"`
	ReactCount int64  `json:"reactCount"`
	Title      string `json:"title"`
}

type FeedbackDto struct {
}
