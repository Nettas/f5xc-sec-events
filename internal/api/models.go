package api

import "encoding/json"

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
	SuspicionScore                float64 `json:"suspicion_score"`
	WafSuspicionScore             float64 `json:"waf_suspicion_score"`
	BotDefenseSuspicionScore      float64 `json:"bot_defense_suspicion_score"`
	BehaviorAnomalyScore          float64 `json:"behavior_anomaly_score"`
	FeatureScore                  string  `json:"feature_score"`
	IpReputationSuspicionScore    float64 `json:"ip_reputation_suspicion_score"`
	ForbiddenAccessSuspicionScore float64 `json:"forbidden_access_suspicion_score"`
	FailedLoginSuspicionScore     float64 `json:"failed_login_suspicion_score"`
	RateLimitSuspicionScore       float64 `json:"rate_limit_suspicion_score"`
	WafSecEventCount              int     `json:"waf_sec_event_count"`
	BotDefenseSecEventCount       int     `json:"bot_defense_sec_event_count"`
	ReqCount                      int     `json:"req_count"`
	ErrCount                      int     `json:"err_count"`
	FailedLoginCount              int     `json:"failed_login_count"`
	ForbiddenAccessCount          int     `json:"forbidden_access_count"`
	PageNotFoundCount             int     `json:"page_not_found_count"`
	RateLimitingCount             int     `json:"rate_limiting_count"`
	ApiepAnomaly                  int             `json:"apiep_anomaly"`
	OriginalTopicName             string          `json:"original_topic_name"`
	// Per-request fields (present on waf_sec_event type)
	Method            string          `json:"method,omitempty"`
	RspCode           int             `json:"rsp_code,omitempty"`
	Action            string          `json:"action,omitempty"`
	Domain            string          `json:"domain,omitempty"`
	ReqPath           string          `json:"req_path,omitempty"`
	ReqID             string          `json:"req_id,omitempty"`
	Authority         string          `json:"authority,omitempty"`
	ApiEndpoint       string          `json:"api_endpoint,omitempty"`
	ReqSize           int             `json:"req_size,omitempty"`
	RspSize           int             `json:"rsp_size,omitempty"`
	ReqRisk           string          `json:"req_risk,omitempty"`
	UpstreamRspCode   int             `json:"upstream_rsp_code,omitempty"`
	BrowserType       string          `json:"browser_type,omitempty"`
	DeviceType        string          `json:"device_type,omitempty"`
	UserAgent         string          `json:"user_agent,omitempty"`
	TLSFingerprint    string          `json:"tls_fingerprint,omitempty"`
	JA4TLSFingerprint string          `json:"ja4_tls_fingerprint,omitempty"`
	SrcSite           string          `json:"src_site,omitempty"`
	Src               string          `json:"src,omitempty"`
	ReqParams         string          `json:"req_params,omitempty"`
	Signatures        json.RawMessage `json:"signatures,omitempty"`
	ReqRiskReasons    json.RawMessage `json:"req_risk_reasons,omitempty"`
	// Complex / variable-shape fields
	PolicyHits        json.RawMessage `json:"policy_hits,omitempty"`
	TimeseriesEnabled bool            `json:"timeseries_enabled,omitempty"`
	Extra             map[string]json.RawMessage `json:"-"`
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
