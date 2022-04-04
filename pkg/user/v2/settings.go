package v2

type settings struct {
	Bell      bool     `json:"bell"`
	Ping      bool     `json:"ping"`
	IsSlack   bool     `json:"isSlack"`
	FmtHour24 bool     `json:"fmtHour24"`
	Pronouns  []string `json:"pronouns"`
	Nick      string   `json:"nick"`
}
