package model

import "time"

type Command struct {
	SessionID  string `json:"session"`
	OrgID      string `json:"org_id"`
	Input      string `json:"input"`
	Output     string `json:"output"`
	User       string `json:"user"` // 系统用户的 user(username)
	Server     string `json:"asset"`
	SystemUser string `json:"system_user"`
	Timestamp  int64  `json:"timestamp"`
	RiskLevel  int64  `json:"risk_level"`
	Owner      string `json:"owner"` //  用户 user(username)
	OwnerID    string `json:"owner_id"`

	DateCreated time.Time `json:"@timestamp"`
}

const (
	HighRiskFlag = "1"
	LessRiskFlag = "0"
)

const (
	DangerLevel = 5
	NormalLevel = 0
)
