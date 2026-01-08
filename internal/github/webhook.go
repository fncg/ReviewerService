package github

type PullRequestEvent struct {
	Action      string `json:"action"`
	PullRequest struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		HTML  string `json:"html_url"`
		User  struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}
