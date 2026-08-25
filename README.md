# Career Progression Ledger

Career Progression Ledger records the longitudinal development of vocational-bachelor graduates. Graduates, verified employers, counselors and major administrators maintain consent-scoped employment histories, capability milestones, qualifications, training and promotion evidence without overwriting trusted historical facts.

## Run

```sh
GOTOOLCHAIN=local go run ./cmd/server
```

The service exposes `/healthz` and `/readyz`. Set `DB_PATH` to a writable SQLite file. All schema changes are versioned under `migrations/` and applied on startup.

## Roles and privacy

Users authenticate with a server-side revocable session. A graduate can manage their own consent and records; employers can submit verified feedback; counselors can conduct follow-ups; major administrators can freeze statistics and resolve appeals. Sensitive salary ranges and contact details are redacted according to role and consent.
