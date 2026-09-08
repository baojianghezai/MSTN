# Public Company Directory

All endpoints in this document are public and use the HRC response envelope:

```json
{ "code": 0, "message": "success", "data": {} }
```

Only company profiles with `audit=1`, `user_status=1`, and a non-empty company name are visible. Contact people, phone numbers, email addresses, member data, and business certificates are never returned.

## GET /api/v1/companies

Search the company directory. This preserves the previous system's keyword, company nature, and industry filtering, and adds scale and district.

| Query | Type | Description |
| --- | --- | --- |
| `keyword` | string | Matches company name, short name, or short description |
| `nature` | number | Company nature category ID |
| `trade` | number | Industry category ID |
| `scale` | number | Company scale category ID |
| `district` | string | District code or display name |
| `page` | number | Page number, default `1` |
| `pageSize` | number | Page size, default `10`, maximum `100` |

Response data:

```json
{
  "list": [{
    "id": 12,
    "companyname": "Example Technology",
    "natureCn": "Private company",
    "tradeCn": "Internet",
    "districtCn": "Beijing",
    "scaleCn": "100-499 people",
    "logo": "uploads/file/logo.png",
    "shortDesc": "We build recruitment software.",
    "tag": "Five insurance,Weekend off",
    "jobsCount": 3
  }],
  "total": 1,
  "page": 1,
  "pageSize": 10
}
```

`jobsCount` includes only public jobs with `display=1`, `audit=1`, no deletion flag, and a valid deadline.

## GET /api/v1/companies/{id}

Returns safe public company information, increments the company view count, and includes the first four active jobs.

```json
{
  "company": { "id": 12, "companyname": "Example Technology", "jobsCount": 3 },
  "contents": "Company introduction",
  "address": "Public office address",
  "website": "https://example.com",
  "jobs": [{ "id": 91, "jobsName": "Go Engineer", "minwage": 12000, "maxwage": 18000 }],
  "jobsTotal": 3
}
```

## GET /api/v1/companies/{id}/jobs

Returns all active jobs for the company with standard pagination. Query parameters are `page` and `pageSize`.

The job list is ordered by top placement, urgent placement, and refresh time. A company that is no longer public returns business code `1001` with no data.
