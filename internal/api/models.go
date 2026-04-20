package api

// SecurityEvent represents a single WAF/security event from the F5 XC API.
type SecurityEvent struct {
	Timestamp                     string  `json:"@timestamp"`
	Time                          string  `json:"time"`
	StartTime                     int64   `json:"start_time"`
	EndTime                       int64   `json:"end_time"`
	Namespace                     string  `json:"namespace"`
	Tenant                        string  `json:"tenant"`
	SrcIP                         string  `json:"src_ip"`
	Country                       string  `json:"country"`
	City                          string  `json:"city"`
	Region                        string  `json:"region"`
	ASN                           string  `json:"asn"`
	Network                       string  `json:"network"`
	Latitude                      string  `json:"latitude"`
	Longitude                     string  `json:"longitude"`
	VhName                        string  `json:"vh_name"`
	AppType                       string  `json:"app_type"`
	App                           string  `json:"app"`
	Site                          string  `json:"site"`
	Hostname                      string  `json:"hostname"`
	ClusterName                   string  `json:"cluster_name"`
	SummaryMsg                    string  `json:"summary_msg"`
	Message                       string  `json:"message"`
	MessageID                     string  `json:"messageid"`
	MessageKey                    string  `json:"message_key"`
	SecEventType                  string  `json:"sec_event_type"`
	SuspicionLogType              string  `json:"suspicion_log_type"`
	Stream                        string  `json:"stream"`
	User                          string  `json:"user"`
	ThreatLevel                   string  `json:"threat_level"`
	SuspicionScore                string `json:"suspicion_score"`
	WafSuspicionScore             string `json:"waf_suspicion_score"`
	BotDefenseSuspicionScore      string `json:"bot_defense_suspicion_score"`
	BehaviorAnomalyScore          string `json:"behavior_anomaly_score"`
	FeatureScore                  string `json:"feature_score"`
	IpReputationSuspicionScore    string `json:"ip_reputation_suspicion_score"`
	ForbiddenAccessSuspicionScore string `json:"forbidden_access_suspicion_score"`
	FailedLoginSuspicionScore     string `json:"failed_login_suspicion_score"`
	RateLimitSuspicionScore       string `json:"rate_limit_suspicion_score"`
	WafSecEventCount              int     `json:"waf_sec_event_count"`
	BotDefenseSecEventCount       int     `json:"bot_defense_sec_event_count"`
	ReqCount                      int     `json:"req_count"`
	ErrCount                      int     `json:"err_count"`
	FailedLoginCount              int     `json:"failed_login_count"`
	ForbiddenAccessCount          int     `json:"forbidden_access_count"`
	PageNotFoundCount             int     `json:"page_not_found_count"`
	RateLimitingCount             int     `json:"rate_limiting_count"`
	ApiepAnomaly                  string  `json:"apiep_anomaly"`
	OriginalTopicName             string  `json:"original_topic_name"`
}

// EventsResponse is the top-level response envelope from the F5 XC events API.
// The API returns each event as a JSON-encoded string inside the events array.
type EventsResponse struct {
	RawEvents []string `json:"events"`
}

// eventsRequest is the POST body sent to the F5 XC events API.
type eventsRequest struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Namespace string `json:"namespace"`
	Query     string `json:"query"`
}
