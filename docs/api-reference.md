# F5 XC API Reference

## Events Endpoint

**Method:** POST  
**URL:** `https://{tenant}.console.ves.volterra.io/api/data/namespaces/{namespace}/app_security/events`

**Auth Header:** `Authorization: APIToken <F5XC_API_KEY>`

### Request Body

```json
{
  "start_time": "<RFC3339>",
  "end_time":   "<RFC3339>",
  "namespace":  "s-iannetta",
  "query": "{vh_name=\"ves-io-http-loadbalancer-webuiaz\"}"
}
```

- Query filter field is `vh_name` (NOT `virtual_host`)
- `vh_name` pattern confirmed 2026-04-27: `ves-io-http-loadbalancer-{lbname}`
  (earlier docs used `ves-io-{namespace}-{lbname}` — that pattern is wrong for this tenant)
- Omit `query` entirely (empty string) to return all events for the namespace

### Response Shape (confirmed against live API 2026-04-20)

**IMPORTANT:** The `events` array contains JSON-encoded strings, not objects.
Each element must be unmarshalled a second time to get the event fields.

```json
{
  "events": [
    "{\"@timestamp\":\"...\",\"time\":\"...\",\"src_ip\":\"1.2.3.4\",\"vh_name\":\"ves-io-s-iannetta-webuiaz\", ...}",
    "{\"@timestamp\":\"...\",\"time\":\"...\",\"src_ip\":\"5.6.7.8\",\"vh_name\":\"ves-io-s-iannetta-webuiaz\", ...}"
  ]
}
```

### Known Event Fields (from live API — namespace s-iannetta, confirmed 2026-04-27)

**Aggregated event fields** (present on both `malicious_user_sec_event` and `waf_sec_event` aggregate types):
```
@timestamp, apiep_anomaly, app, app_type, asn, behavior_anomaly_score,
bot_defense_sec_event_count, bot_defense_suspicion_score, city, cluster_name,
country, end_time, err_count, failed_login_count, failed_login_suspicion_score,
feature_score, forbidden_access_count, forbidden_access_suspicion_score,
hostname, incremental_activity_info, ip_reputation_suspicion_score,
latitude, longitude, message, messageid, message_key, method_counts,
mitigation_activity_info, namespace, network, original_topic_name,
page_not_found_count, rate_limiting_count, rate_limit_suspicion_score,
region, req_count, sec_event_type, site, src_ip, start_time, stream,
summary_msg, suspicion_log_type, suspicion_score, tenant, threat_level,
time, user, vh_name, waf_sec_event_count, waf_suspicion_score
```

**Per-request fields** (expected on individual `waf_sec_event` hits — types unconfirmed):
```
action, api_endpoint, authority, browser_type, device_type, domain,
ja4_tls_fingerprint, method, req_id, req_params, req_path, req_risk,
req_risk_reasons, req_size, rsp_code, rsp_size, signatures, src,
src_site, tls_fingerprint, upstream_rsp_code, user_agent
```

### Field Type Map — AUTHORITATIVE (confirmed by live curl 2026-04-27)

> Locked for aggregate fields. Per-request fields are unconfirmed — `string` used as safe default.

| Field(s) | JSON type | Go type | Notes |
|----------|-----------|---------|-------|
| `latitude`, `longitude` | string | `string` | Quoted string despite being numeric |
| `start_time`, `end_time` (event payload) | number | `int64` | Unix epoch seconds |
| `start_time`, `end_time` (request body) | string | `string` | RFC3339 — required by API |
| `suspicion_score`, `waf_suspicion_score`, `bot_defense_suspicion_score`, `behavior_anomaly_score`, `ip_reputation_suspicion_score`, `forbidden_access_suspicion_score`, `failed_login_suspicion_score`, `rate_limit_suspicion_score` | float | `float64` | All confirmed score fields |
| `feature_score` | string | `string` | Exception — API sends `"{}"` (JSON-encoded string) |
| `req_count`, `waf_sec_event_count`, `err_count`, `failed_login_count`, `forbidden_access_count`, `page_not_found_count`, `rate_limiting_count`, `bot_defense_sec_event_count` | number | `int` | Confirmed JSON integers |
| `apiep_anomaly` | number | `int` | JSON integer (0 or 1) |
| `policy_hits` | object/null | `json.RawMessage` | Variable shape — stored raw |
| `timeseries_enabled` | bool | `bool` | |
| `rsp_code`, `req_size`, `rsp_size`, `upstream_rsp_code` | unknown | `string` | Changed int→string 2026-04-27; caused Windows 502 when typed as int; confirmed absent from aggregate events |
| `signatures` | array/string | `json.RawMessage` | Array of sig objects; may be double-encoded; parse with `parseJsonField()` in JS |
| `req_risk_reasons` | array/string | `json.RawMessage` | Array of reason strings; may be double-encoded |
| `incremental_activity_info`, `method_counts`, `mitigation_activity_info` | string | not mapped | Double-encoded nested objects — silently ignored by json.Unmarshal |

### Fields Not in Struct (silently ignored)

Go's `json.Unmarshal` discards unknown fields by default (no `DisallowUnknownFields`):
`incremental_activity_info`, `method_counts`, `mitigation_activity_info`

> **Pagination:** Not yet confirmed. Check response for continuation tokens if large result sets are truncated.
