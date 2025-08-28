package types

type Resp struct {
	Success bool `json:"success"`
}

type CommonResponse struct {
	Data       Resp   `json:"data"`
	TargetUser string `json:"targetUser"`
}

type FeedbackDto struct {
}
