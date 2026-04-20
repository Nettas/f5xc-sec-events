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
  "query": "{vh_name=\"ves-io-s-iannetta-my-lb\"}"
}
```

- Query filter field is `vh_name` (NOT `virtual_host`)
- `vh_name` pattern: `ves-io-{namespace}-{lb-name}` — e.g. `ves-io-s-iannetta-webuiaz`
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

### Known Event Fields (from live API — namespace s-iannetta)

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

### Field Type Quirks

| Field | JSON type | Go type |
|-------|-----------|---------|
| `latitude`, `longitude` | string (not number) | `string` |
| `start_time`, `end_time` (in event) | number (Unix epoch) | `int64` |
| `start_time`, `end_time` (in request) | string (RFC3339) | `string` |
| score fields (`suspicion_score`, etc.) | number | `float64` |
| count fields (`req_count`, etc.) | number | `int` |

> **Pagination:** Not yet confirmed. Check response for continuation tokens if large result sets are truncated.
