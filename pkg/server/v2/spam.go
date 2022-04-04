package v2

type spamManagement struct {
	userChatSpamCounts  map[string]int
	userLoginSpamCounts map[string]int
}

func (sm spamManagement) init() {
	// TODO: maybe add some IP-based factor to disallow rapid key-gen attempts
	sm.userChatSpamCounts = make(map[string]int)
	sm.userLoginSpamCounts = make(map[string]int)
}
