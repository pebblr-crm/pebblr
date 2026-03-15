Run all quality gates for pebblr and report results.

```bash
echo "=== Tests ===" && make test
echo "=== Lint ===" && make lint
echo "=== Typecheck ===" && make typecheck
```

Report which gates passed and which failed. If any fail, summarize the errors and suggest fixes. Do not proceed with commits until all three pass.
