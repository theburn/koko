package model

type SharingSession struct {
	ID          string `json:"id"`
	IsActive    bool   `json:"is_active"`
	ExpiredTime int    `json:"expired_time"`
	Session     string `json:"session"`
	OrgId       string `json:"org_id"`
	OrgName     string `json:"org_name"`
	Code        string `json:"code"`
}
