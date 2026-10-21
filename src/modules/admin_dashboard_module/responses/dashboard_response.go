package responses

type StatWithChange struct {
	Total     int64 `json:"total"`
	ThisMonth int64 `json:"thisMonth"`
}

type DashboardResponse struct {
	PendingReports       StatWithChange `json:"pendingReports"`
	TotalUsers           StatWithChange `json:"totalUsers"`
	ActiveProperties     StatWithChange `json:"activeProperties"`
	PendingVerifications StatWithChange `json:"pendingVerifications"`
	BannedHosts          StatWithChange `json:"bannedHosts"`
}
