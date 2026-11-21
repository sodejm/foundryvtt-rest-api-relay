# CodeQL checks

The workflow runs the default and `security-extended` suites for Go,
JavaScript/TypeScript, Python, and GitHub Actions on pull requests, main pushes,
and a weekly schedule. Go uses `go build ./...` from `go-relay`, with CGO disabled
as in the container build. It does not start the relay or contact its services.
The workflow pins the CodeQL 2.27.4 bundle used for local query validation and
the matching `codeql/go-all` 8.1.0 pack.

## Admin API middleware invariant

`go/RequireAdminRouterMiddleware.ql` checks registrations reachable from
`internal/handler.AdminRouter` under `/admin/api`. It follows local router aliases,
chi `Group`, `Route`, and `With`, and statically resolved mounted constructors,
including a constructor held in a local variable. Each registration path must
inherit `internal/middleware.RequireAdmin`; a guard on a sibling path does not
authorize another mount of the same constructor. Direct `/api/...` registrations
and `Route`/`Mount` prefixes are included. Other prefixes, such as `/apiary` and
`/other/api`, are outside that boundary.

A guard must be a direct `Use(RequireAdmin(...))` statement before registration
in the same function body, on a receiver with one shared local origin, or a
`With(RequireAdmin(...))` on that path. Conditional, late, and ambiguous receiver
guards do not suppress an alert. This conservative static model does not prove
arbitrary helper functions, dynamic paths, middleware aliases, interface dispatch,
or control flow that the model does not recognize.

The existing intentional exceptions are narrow:

- The admin HTML shell and `/admin/auth` credential exchange/logout routes are
  outside `/admin/api`; their own access and CSRF behavior needs separate review.
- Only `GET /admin/api/tests/report` registered directly at the API root with
  `internal/handler.TestReportHandler` is exempt. The current endpoint intentionally
  relies on the admin IP allowlist. A different method, handler package, or path
  does not qualify. This exemption records the current policy; it does not prove
  that deployment's allowlist configuration.

## Regression tests

Run with a CodeQL CLI containing the Go extractor and the declared `codeql/go-all`
pack:

```sh
codeql test run .github/codeql/go/tests --threads=2
```

The Go analysis job runs this command using the CLI path returned by its init
step, after production analysis. The inert chi and middleware stubs are only for
extracting query fixtures; they do not implement HTTP behavior. Tests cover safe
nested groups, `With`, local aliases, self-protected mounts, intentional exceptions,
and missing guards on sibling mounts, conditional/late/ambiguous guards, direct API
registrations, and near-miss exemptions. The separate fixture module is outside
the production Go source root, and the interpreted-language configuration also
excludes it. Deliberately unsafe fixtures therefore do not emit production alerts.
