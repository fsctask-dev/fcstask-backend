package model

type PlatformStats struct {
	TotalCourses   int64 `json:"totalCourses"`
	PublicCourses  int64 `json:"publicCourses"`
	PrivateCourses int64 `json:"privateCourses"`
	TotalUsers     int64 `json:"totalUsers"`
}
