package model

type Course struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	RepoTemplate string `json:"repoTemplate"`
	Description  string `json:"description"`
	URL          string `json:"url"`
}

type BoardDeadline struct {
	ID      string  `json:"id"`
	Label   string  `json:"label"`
	Percent float64 `json:"percent"`
	DueAt   string  `json:"dueAt"`
	Status  string  `json:"status"`
}

type BoardTask struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Score       int     `json:"score"`
	ScoreEarned int     `json:"scoreEarned"`
	Stats       float64 `json:"stats"`
	IsBonus     bool    `json:"isBonus,omitempty"`
	IsSpecial   bool    `json:"isSpecial,omitempty"`
	URL         string  `json:"url,omitempty"`
}

type BoardGroup struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	IsSpecial bool            `json:"isSpecial,omitempty"`
	StartedAt string          `json:"startedAt"`
	EndsAt    string          `json:"endsAt"`
	Deadlines []BoardDeadline `json:"deadlines"`
	Tasks     []BoardTask     `json:"tasks"`
}

type TaskBoardSummary struct {
	CourseName    string       `json:"courseName"`
	CourseStatus  string       `json:"courseStatus"`
	SolvedScore   int          `json:"solvedScore"`
	MaxScore      int          `json:"maxScore"`
	SolvedPercent int          `json:"solvedPercent"`
	Groups        []BoardGroup `json:"groups"`
}
