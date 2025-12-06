package model

import "database/sql"

type User struct {
	ID        int
	Username  string
	Password  string
	CreatedAt string
}

type Talent struct {
	ID            int
	UserID        int
	Name          string
	Affiliation   sql.NullString
	Beauty        int
	Cuteness      int
	Talent        int
	TotalBeauty   int
	TotalCuteness int
	TotalTalent   int
	CreatedAt     string
}

type Adjustment struct {
	ID             int
	TalentID       int
	AdjustmentType string
	Points         int
	Reason         string
	CreatedAt      string
}
