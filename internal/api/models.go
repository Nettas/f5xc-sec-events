package api

// SecurityEvent represents a single WAF/security event from the F5 XC API.
type SecurityEvent struct {
	Time         string `json:"time"`
	SrcIP        string `json:"src_ip"`
	ReqPath      string `json:"req_path"`
	Method       string `json:"method"`
	ResponseCode int    `json:"response_code"`
	ReqID        string `json:"req_id"`
	WAFAction    string `json:"waf_action"`
	AttackType   string `json:"attack_type"`
	Severity     string `json:"severity"`
	VirtualHost  string `json:"virtual_host"`
}

// EventsResponse is the top-level response envelope from the F5 XC events API.
type EventsResponse struct {
	Events []SecurityEvent `json:"events"`
}

// eventsRequest is the POST body sent to the F5 XC events API.
type eventsRequest struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Namespace string `json:"namespace"`
	Query     string `json:"query"`
}
