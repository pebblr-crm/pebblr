Audit the codebase for secret handling violations.

Search for patterns that indicate secrets may be read from environment variables instead of file mounts:

```bash
# Find os.Getenv calls in Go code
grep -rn "os\.Getenv" --include="*.go" .

# Find process.env in TypeScript
grep -rn "process\.env" --include="*.ts" --include="*.tsx" .
```

For each match found:
1. Determine if it's reading a secret value (password, token, key, connection string) vs. a non-secret config value (e.g., PORT, LOG_LEVEL)
2. If it's a secret: flag it as a violation — secrets must come from file mounts (e.g., `/run/secrets/`)
3. If it's non-sensitive config: note it as acceptable

Report a summary of violations vs. acceptable uses. Violations must be fixed before merge.
