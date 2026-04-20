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
  "query": "{virtual_host=\"my-lb\"}"
}
```

### Response Shape (approximate)

```json
{
  "events": [
    {
      "time": "...",
      "src_ip": "...",
      "req_path": "...",
      "method": "...",
      "response_code": 403,
      "req_id": "...",
      "waf_action": "BLOCK|ALLOW",
      "attack_type": "...",
      "severity": "...",
      "virtual_host": "..."
    }
  ]
}
```

> **Note:** Verify exact field names against the live API before finalizing models.go.
> The API may paginate — check response for continuation tokens.
