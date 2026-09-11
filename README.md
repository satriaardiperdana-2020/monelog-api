# Monelog planning pack

Draft v0.1 • 11 September 2026 • Planning documentation, not application code.

## Read in this order

1. [Requirements](docs/requirements.md) — what the product must do; confirm assumptions.
2. [Architecture](docs/architecture.md) — proposed technical boundaries.
3. [Database](docs/database.md) — proposed data model and invariants.
4. [API](docs/api.md) — draft HTTP contract, not yet a complete OpenAPI specification.
5. [Roadmap](docs/roadmap.md) — backend-first delivery sequence.
6. [Issue index](docs/issues.md) — task order and dependencies.
7. [ISSUE-001](docs/issues/ISSUE-001-project-setup.md) — first backend task.
8. [PLAN-001](docs/plans/PLAN-001.md) — proposed implementation plan for Issue 001.
9. [Workflow](docs/workflow.md) — repeatable planning, implementation and review process.

The Markdown issue files are the portable source of truth. They are not GitHub Issues yet.
Start with backend Issues 001–007 in `monelog-api`. Build the Vue frontend in
`monelog-app` beginning with Issue 008 after the backend MVP contract is stable.
