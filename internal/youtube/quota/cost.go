package quota

// Method is a YouTube Data API v3 method name (resource.method).
type Method string

const (
	// VideosList is videos.list (default unit bucket).
	VideosList Method = "videos.list"
	// LiveChatMessagesList is liveChatMessages.list (default unit bucket).
	LiveChatMessagesList Method = "liveChatMessages.list"
	// SearchList is search.list (separate daily search bucket).
	SearchList Method = "search.list"
	// Unknown is any other /youtube/v3/ path. Google bills at least 1 unit.
	Unknown Method = "unknown"
)

const (
	// DefaultUnitsPerDay is Google’s documented default for the combined unit
	// bucket (not search.list / videos.insert).
	DefaultUnitsPerDay = 10_000
	// DefaultSearchPerDay is Google’s documented default for search.list calls.
	DefaultSearchPerDay = 100
	// LiveChatMessagesListCost is what Cloud Console bills per liveChatMessages.list.
	// The public calculator table currently lists 1; measured project usage is 5.
	LiveChatMessagesListCost = 5
)

// Cost is the quota cost for m. Unknown or empty methods cost 1 (Google’s minimum).
func Cost(m Method) int {
	switch m {
	case LiveChatMessagesList:
		return LiveChatMessagesListCost
	default:
		return 1
	}
}

// searchBucket reports whether m spends the separate search.list quota.
func searchBucket(m Method) bool {
	return m == SearchList
}
