## Summary

<!-- Why this change exists. One to three bullets. -->

## Test plan

- [ ] `go test ./...`
- [ ] If API changed: note FE impact or link paired PR in `crm-fe`

## Checklist

- [ ] No unrelated changes in this PR
- [ ] Handler stays thin; business rules in application/domain
- [ ] No secrets (`.env`, keys, DSN) committed
- [ ] AutoMigrate / permission seed impact considered for deploy
