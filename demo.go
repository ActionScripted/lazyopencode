package main

import (
	"path/filepath"
	"strings"
	"time"
)

// demoSessions returns a hardcoded set of fake sessions for --demo mode.
// Nothing here touches the filesystem or the DB.
func demoSessions() []session {
	home := resolveHome()

	coolProject := filepath.Join(home, "cool-project")
	anotherProject := filepath.Join(home, "another-project")
	desktop := filepath.Join(home, "Desktop")

	now := time.Now()
	h := func(hours float64) time.Time { return now.Add(-time.Duration(hours * float64(time.Hour))) }

	sessions := []session{
		{
			ID:               "demo-001",
			Title:            "add fake data for screenshot",
			Directory:        coolProject,
			CreatedAt:        h(0.5),
			UpdatedAt:        h(0.1),
			DisplayDir:       "~/cool-project",
			ShortDir:         "cool-project",
			SummaryFiles:     10,
			SummaryAdditions: 312,
			SummaryDeletions: 47,
		},
		{
			ID:               "demo-002",
			Title:            "refactor auth middleware",
			Directory:        coolProject,
			CreatedAt:        h(3),
			UpdatedAt:        h(1),
			DisplayDir:       "~/cool-project",
			ShortDir:         "cool-project",
			SummaryFiles:     4,
			SummaryAdditions: 88,
			SummaryDeletions: 33,
		},
		{
			ID:               "demo-003",
			Title:            "wire up rate limiting to redis",
			Directory:        coolProject,
			CreatedAt:        h(26),
			UpdatedAt:        h(24),
			DisplayDir:       "~/cool-project",
			ShortDir:         "cool-project",
			SummaryFiles:     6,
			SummaryAdditions: 143,
			SummaryDeletions: 21,
		},
		{
			ID:               "demo-004",
			Title:            "add structured logging with slog",
			Directory:        coolProject,
			CreatedAt:        h(50),
			UpdatedAt:        h(48),
			DisplayDir:       "~/cool-project",
			ShortDir:         "cool-project",
			SummaryFiles:     11,
			SummaryAdditions: 224,
			SummaryDeletions: 198,
		},
		{
			ID:               "demo-005",
			Title:            "benchmark and optimise hot path",
			Directory:        coolProject,
			CreatedAt:        h(75),
			UpdatedAt:        h(72),
			DisplayDir:       "~/cool-project",
			ShortDir:         "cool-project",
			SummaryFiles:     3,
			SummaryAdditions: 61,
			SummaryDeletions: 44,
		},
		{
			ID:               "demo-006",
			Title:            "fix flaky test in CI",
			Directory:        anotherProject,
			CreatedAt:        h(5),
			UpdatedAt:        h(2),
			DisplayDir:       "~/another-project",
			ShortDir:         "another-project",
			SummaryFiles:     2,
			SummaryAdditions: 15,
			SummaryDeletions: 9,
		},
		{
			ID:               "demo-007",
			Title:            "set up docker compose for local dev",
			Directory:        anotherProject,
			CreatedAt:        h(30),
			UpdatedAt:        h(6),
			DisplayDir:       "~/another-project",
			ShortDir:         "another-project",
			SummaryFiles:     6,
			SummaryAdditions: 201,
			SummaryDeletions: 12,
		},
		{
			ID:               "demo-008",
			Title:            "add prometheus metrics endpoint",
			Directory:        anotherProject,
			CreatedAt:        h(55),
			UpdatedAt:        h(52),
			DisplayDir:       "~/another-project",
			ShortDir:         "another-project",
			SummaryFiles:     5,
			SummaryAdditions: 132,
			SummaryDeletions: 8,
		},
		{
			ID:               "demo-009",
			Title:            "write integration tests for the grpc server",
			Directory:        anotherProject,
			CreatedAt:        h(80),
			UpdatedAt:        h(76),
			DisplayDir:       "~/another-project",
			ShortDir:         "another-project",
			SummaryFiles:     7,
			SummaryAdditions: 319,
			SummaryDeletions: 14,
		},
		{
			ID:               "demo-010",
			Title:            "replace homegrown pool with sync.Pool",
			Directory:        anotherProject,
			CreatedAt:        h(100),
			UpdatedAt:        h(96),
			DisplayDir:       "~/another-project",
			ShortDir:         "another-project",
			SummaryFiles:     4,
			SummaryAdditions: 57,
			SummaryDeletions: 123,
		},
		{
			ID:               "demo-011",
			Title:            "update README",
			Directory:        home,
			CreatedAt:        h(10),
			UpdatedAt:        h(8),
			DisplayDir:       "~",
			ShortDir:         "~",
			SummaryFiles:     1,
			SummaryAdditions: 42,
			SummaryDeletions: 5,
		},
		{
			ID:               "demo-012",
			Title:            "write a Makefile from scratch",
			Directory:        home,
			CreatedAt:        h(120),
			UpdatedAt:        h(118),
			DisplayDir:       "~",
			ShortDir:         "~",
			SummaryFiles:     1,
			SummaryAdditions: 68,
			SummaryDeletions: 0,
		},
		{
			ID:               "demo-013",
			Title:            "dark mode polish",
			Directory:        desktop,
			CreatedAt:        h(14),
			UpdatedAt:        h(12),
			DisplayDir:       "~/Desktop",
			ShortDir:         "Desktop",
			SummaryFiles:     8,
			SummaryAdditions: 190,
			SummaryDeletions: 67,
		},
		{
			ID:               "demo-014",
			Title:            "port CLI tool from Python to Go",
			Directory:        desktop,
			CreatedAt:        h(60),
			UpdatedAt:        h(56),
			DisplayDir:       "~/Desktop",
			ShortDir:         "Desktop",
			SummaryFiles:     14,
			SummaryAdditions: 487,
			SummaryDeletions: 0,
		},
		{
			ID:               "demo-015",
			Title:            "set up pre-commit hooks with golangci-lint",
			Directory:        desktop,
			CreatedAt:        h(90),
			UpdatedAt:        h(88),
			DisplayDir:       "~/Desktop",
			ShortDir:         "Desktop",
			SummaryFiles:     3,
			SummaryAdditions: 45,
			SummaryDeletions: 2,
		},
		{
			ID:               "demo-016",
			Title:            "implement LRU cache with generics",
			Directory:        coolProject,
			CreatedAt:        h(110),
			UpdatedAt:        h(108),
			DisplayDir:       "~/cool-project",
			ShortDir:         "cool-project",
			SummaryFiles:     3,
			SummaryAdditions: 174,
			SummaryDeletions: 91,
		},
		{
			ID:               "demo-017",
			Title:            "add graceful shutdown handling",
			Directory:        anotherProject,
			CreatedAt:        h(130),
			UpdatedAt:        h(128),
			DisplayDir:       "~/another-project",
			ShortDir:         "another-project",
			SummaryFiles:     5,
			SummaryAdditions: 83,
			SummaryDeletions: 27,
		},
		{
			ID:               "demo-018",
			Title:            "profile memory allocations in parser",
			Directory:        coolProject,
			CreatedAt:        h(140),
			UpdatedAt:        h(138),
			DisplayDir:       "~/cool-project",
			ShortDir:         "cool-project",
			SummaryFiles:     2,
			SummaryAdditions: 38,
			SummaryDeletions: 55,
		},
		{
			ID:               "demo-019",
			Title:            "add fuzzing for the config parser",
			Directory:        anotherProject,
			CreatedAt:        h(155),
			UpdatedAt:        h(152),
			DisplayDir:       "~/another-project",
			ShortDir:         "another-project",
			SummaryFiles:     3,
			SummaryAdditions: 99,
			SummaryDeletions: 4,
		},
		{
			ID:               "demo-020",
			Title:            "write a Rust FFI binding for libsodium",
			Directory:        desktop,
			CreatedAt:        h(170),
			UpdatedAt:        h(166),
			DisplayDir:       "~/Desktop",
			ShortDir:         "Desktop",
			SummaryFiles:     6,
			SummaryAdditions: 231,
			SummaryDeletions: 18,
		},
		{
			ID:               "demo-021",
			Title:            "set up Nix flake for dev environment",
			Directory:        home,
			CreatedAt:        h(180),
			UpdatedAt:        h(178),
			DisplayDir:       "~",
			ShortDir:         "~",
			SummaryFiles:     4,
			SummaryAdditions: 112,
			SummaryDeletions: 3,
		},
		{
			ID:               "demo-022",
			Title:            "tune postgres connection pool settings",
			Directory:        anotherProject,
			CreatedAt:        h(190),
			UpdatedAt:        h(188),
			DisplayDir:       "~/another-project",
			ShortDir:         "another-project",
			SummaryFiles:     2,
			SummaryAdditions: 29,
			SummaryDeletions: 17,
		},
		{
			ID:               "demo-023",
			Title:            "generate OpenAPI spec from Go types",
			Directory:        coolProject,
			CreatedAt:        h(200),
			UpdatedAt:        h(198),
			DisplayDir:       "~/cool-project",
			ShortDir:         "cool-project",
			SummaryFiles:     8,
			SummaryAdditions: 266,
			SummaryDeletions: 34,
		},
		{
			ID:               "demo-024",
			Title:            "add tracing with opentelemetry",
			Directory:        coolProject,
			CreatedAt:        h(215),
			UpdatedAt:        h(212),
			DisplayDir:       "~/cool-project",
			ShortDir:         "cool-project",
			SummaryFiles:     9,
			SummaryAdditions: 303,
			SummaryDeletions: 11,
		},
		{
			ID:               "demo-025",
			Title:            "🦀 rewrite it in rust",
			Directory:        desktop,
			CreatedAt:        h(230),
			UpdatedAt:        h(226),
			DisplayDir:       "~/Desktop",
			ShortDir:         "Desktop",
			SummaryFiles:     47,
			SummaryAdditions: 9823,
			SummaryDeletions: 4201,
		},
	}
	for i := range sessions {
		sessions[i].FilterKey = strings.ToLower(sessions[i].Title + " " + sessions[i].Directory)
	}
	return sessions
}

// demoMessages returns fake chat messages for --demo mode.
// The "featured" session (demo-001) gets the fun meta conversation.
// All other sessions get a short plausible exchange.
func demoMessages(sessionID string) []message {
	switch sessionID {
	case "demo-001":
		return demoFeaturedMessages()
	case "demo-002":
		return []message{
			{Role: "user", Text: "the auth middleware is getting hard to follow — can you break it into smaller pieces?"},
			{Role: "assistant", Text: "Sure. I'll extract token validation, role checking, and rate-limit headers into separate middleware functions and compose them. That should make each piece independently testable too."},
			{Role: "user", Text: "perfect, and make sure the error responses stay consistent"},
			{Role: "assistant", Text: "Done. All error paths now go through a single `writeAuthError` helper so the shape is always `{error: string, code: string}`. Updated `middleware/auth.go`, `middleware/ratelimit.go`, and the tests."},
		}
	case "demo-003":
		return []message{
			{Role: "user", Text: "add redis-backed rate limiting — 100 req/min per API key"},
			{Role: "assistant", Text: "I'll use a sliding window counter in Redis with a TTL of 60 s. Each API key gets its own key prefixed with `rl:`. Exceeded limits return 429 with a `Retry-After` header. Adding the middleware now."},
		}
	case "demo-004":
		return []message{
			{Role: "user", Text: "we're using log.Printf everywhere, can you swap it for slog with structured fields?"},
			{Role: "assistant", Text: "I'll do a full pass — replace all `log.Printf` / `log.Println` calls with `slog.Info` / `slog.Error` and add relevant key-value fields. I'll set up a JSON handler for production and a text handler for local dev via a build tag. Should be a clean mechanical change with no logic touching."},
			{Role: "user", Text: "nice, add a request-id to every log line from the HTTP layer"},
			{Role: "assistant", Text: "Done. The request-id is generated in a middleware, stored in context, and pulled out in a `slog.Logger` that's attached to the context too. Every log statement downstream gets it automatically without any extra plumbing at each call site."},
		}
	case "demo-005":
		return []message{
			{Role: "user", Text: "the /encode endpoint is slow under load — can you profile it and tell me what's hot?"},
			{Role: "assistant", Text: "Running `go test -bench=. -cpuprofile=cpu.prof` and then `pprof` on the result. The top offender is `bytes.Split` inside the inner loop — it allocates a new slice on every call. I'll replace it with `bytes.IndexByte` and a manual walk. Should cut allocations by ~80% on that path."},
			{Role: "user", Text: "do it"},
			{Role: "assistant", Text: "Done. Benchmark before: 4821 ns/op, 48 allocs/op. After: 891 ns/op, 6 allocs/op. The hot path now fits in two cache lines. Committed with the benchmark numbers in the commit message for posterity."},
		}
	case "demo-006":
		return []message{
			{Role: "user", Text: "the TestUserCache test fails maybe 1 in 10 runs in CI — any idea why?"},
			{Role: "assistant", Text: "It's a race condition — the test reads the cache immediately after a goroutine writes to it without any synchronisation. I'll add a `WaitGroup` and move the assertion after `wg.Wait()`. That should make it deterministic."},
			{Role: "user", Text: "nice catch, ship it"},
		}
	case "demo-007":
		return []message{
			{Role: "user", Text: "can you write a docker-compose.yml that brings up postgres, redis, and the app?"},
			{Role: "assistant", Text: "Done. Postgres on 5432, Redis on 6379, the app on 8080 with health checks and a named volume for postgres data. I've also added a `.env.example` with the required vars so new devs know what to fill in."},
		}
	case "demo-008":
		return []message{
			{Role: "user", Text: "add a /metrics endpoint that prometheus can scrape"},
			{Role: "assistant", Text: "I'll use the official `prometheus/client_golang` library. Exposing the default Go runtime metrics plus four custom counters: `http_requests_total`, `http_request_duration_seconds`, `db_query_duration_seconds`, and `cache_hits_total`. The endpoint lives at `/metrics` and is excluded from auth middleware."},
			{Role: "user", Text: "also add a histogram for response size"},
			{Role: "assistant", Text: "Added `http_response_size_bytes` as a Histogram with buckets at 256 B, 1 KB, 10 KB, 100 KB, 1 MB. Registered alongside the other metrics in `internal/metrics/metrics.go`."},
		}
	case "demo-009":
		return []message{
			{Role: "user", Text: "can you write integration tests for the gRPC server? happy path plus error cases"},
			{Role: "assistant", Text: "I'll spin up an in-process `grpc.Server` with `bufconn` so there's no real network and the tests stay fast. Testing: successful RPC calls, unauthenticated requests (expect codes.Unauthenticated), malformed payloads (codes.InvalidArgument), and deadline exceeded on a slow handler. Writing them now."},
			{Role: "user", Text: "also test that the server drains in-flight requests on shutdown"},
			{Role: "assistant", Text: "Done. The drain test starts a slow RPC, signals shutdown, and asserts the in-flight call completes before the server stops. Uses a `WaitGroup` and a 5 s timeout so it fails loudly if drain is broken."},
		}
	case "demo-010":
		return []message{
			{Role: "user", Text: "we wrote our own goroutine pool but sync.Pool does most of what we need — can you swap it out?"},
			{Role: "assistant", Text: "Looking at the existing pool... it's basically `sync.Pool` with extra steps. I'll remove `internal/pool/pool.go` entirely, replace the call sites with `sync.Pool`, and adjust the `New` func to return the right type. The test coverage transfers over cleanly."},
		}
	case "demo-011":
		return []message{
			{Role: "user", Text: "the README is embarrassingly out of date, can you rewrite the setup section?"},
			{Role: "assistant", Text: "Rewritten. Prerequisites, `git clone`, `cp .env.example .env`, `docker compose up`, and the `make dev` shortcut. Also added a troubleshooting section for the two most common first-run gotchas."},
		}
	case "demo-012":
		return []message{
			{Role: "user", Text: "this project has no Makefile, can you write one with the usual targets?"},
			{Role: "assistant", Text: "Done. Targets: `build`, `run`, `test`, `lint` (golangci-lint), `fmt` (gofmt + goimports), `vet`, `clean`, and `check` which runs fmt → vet → lint → test in order. `.PHONY` declared for all of them. Uses `$(shell go env GOBIN)` so it works regardless of PATH setup."},
		}
	case "demo-013":
		return []message{
			{Role: "user", Text: "dark mode looks great but the sidebar active state is barely visible — fix the contrast"},
			{Role: "assistant", Text: "The active background colour had a contrast ratio of 1.4:1 against the sidebar — well below the 4.5:1 WCAG AA minimum. Bumped the active state to match the primary accent colour which sits at 4.8:1. Also fixed the hover state which had the same problem."},
			{Role: "user", Text: "much better — while you're in there make the scrollbar less ugly"},
			{Role: "assistant", Text: "Done. Thin scrollbar with the accent colour as the thumb, transparent track, only visible on hover. Applied globally via the app stylesheet."},
		}
	case "demo-014":
		return []message{
			{Role: "user", Text: "I have a Python CLI tool that does log parsing — can you port it to Go?"},
			{Role: "assistant", Text: "Sure. I'll keep the same flag interface so existing scripts don't break. The Go version uses `bufio.Scanner` for streaming line-by-line (the Python version loaded the whole file into memory), so it'll handle large logs without issues. Matching the output format exactly so it's a drop-in replacement."},
			{Role: "user", Text: "can you add a --json flag to output structured records instead of plain text?"},
			{Role: "assistant", Text: "Added. `--json` emits one JSON object per line: `{\"ts\": \"...\", \"level\": \"...\", \"msg\": \"...\", \"fields\": {...}}`. Plays nicely with `jq`. Default output is unchanged."},
		}
	case "demo-015":
		return []message{
			{Role: "user", Text: "set up pre-commit hooks so nobody can commit without passing golangci-lint"},
			{Role: "assistant", Text: "Added a `.git/hooks/pre-commit` script (and a `scripts/install-hooks.sh` for new contributors) that runs `golangci-lint run --fast` against staged Go files only. Fast enough that it won't annoy people. Also wired it into the Makefile as `make install-hooks`."},
		}
	case "demo-016":
		return []message{
			{Role: "user", Text: "write an LRU cache — I want it to work with any key/value type"},
			{Role: "assistant", Text: "I'll use generics — `type Cache[K comparable, V any]` backed by a doubly-linked list and a map. O(1) get and put. Capacity is set at construction. Thread-safe with a `sync.Mutex`. Writing `cache.go` and `cache_test.go` now."},
			{Role: "user", Text: "add an optional TTL per entry"},
			{Role: "assistant", Text: "Done. `Put` now accepts a variadic `...time.Duration` — omit it for no expiry. Expiry is checked lazily on `Get` and evicted entries are removed from the list. No background goroutine needed for the expected workload."},
		}
	case "demo-017":
		return []message{
			{Role: "user", Text: "the server just hard-exits on SIGTERM — can you add graceful shutdown?"},
			{Role: "assistant", Text: "I'll listen for SIGTERM and SIGINT, call `server.Shutdown` with a 30 s context, and drain in-flight requests before exiting. The DB connection pool gets closed after the server stops so no requests are cut off mid-query. Exit code 0 on clean shutdown, 1 on timeout."},
		}
	case "demo-018":
		return []message{
			{Role: "user", Text: "the parser is allocating a lot — can you figure out where and reduce it?"},
			{Role: "assistant", Text: "Running `go test -memprofile=mem.prof -bench=BenchmarkParse` and analysing with pprof. The main culprit is `strings.Split` inside `parseFields` — returns a `[]string` even when we only need one field. Replacing with `strings.IndexByte` + slice index drops allocations from 34/op to 6/op. Also found an unnecessary `fmt.Sprintf` in the error path — swapped for `errors.New`."},
		}
	case "demo-019":
		return []message{
			{Role: "user", Text: "add a fuzz target for the config parser — it's had two bugs from malformed input already"},
			{Role: "assistant", Text: "I'll add `FuzzParseConfig` in `config_fuzz_test.go` using the native `testing/fuzz` package. Seed corpus from the existing test fixtures plus a few hand-crafted edge cases: empty input, just whitespace, deeply nested keys, binary data in values. Run with `go test -fuzz=FuzzParseConfig`."},
			{Role: "user", Text: "nice — did fuzzing find anything?"},
			{Role: "assistant", Text: "It found a nil dereference when a section header has no closing bracket — we were assuming the bracket was always there. Fixed with an early return and a proper error. Corpus entry added so it'll never regress."},
		}
	case "demo-020":
		return []message{
			{Role: "user", Text: "I need to call libsodium from Rust — can you write the FFI bindings?"},
			{Role: "assistant", Text: "I'll write a `sodium-sys` crate with raw `unsafe` bindings generated by `bindgen`, then a safe wrapper crate on top. Starting with `crypto_secretbox` (symmetric encryption) and `crypto_sign` (ed25519 signing) since those are what you need. The build script links against the system libsodium via `pkg-config`."},
			{Role: "user", Text: "add a test that round-trips an encrypt/decrypt to prove the bindings work"},
			{Role: "assistant", Text: "Done. The test encrypts a known plaintext, decrypts it, and asserts byte equality. Also tests that decryption fails with a tampered ciphertext. Runs without any mocking — real libsodium calls end to end."},
		}
	case "demo-021":
		return []message{
			{Role: "user", Text: "set up a Nix flake so anyone can get the exact same dev environment"},
			{Role: "assistant", Text: "Writing `flake.nix` with a `devShell` that provides Go, golangci-lint, goimports, delve, and postgresql for the integration tests. Pinned to nixpkgs 23.11. Anyone with Nix + flakes enabled can run `nix develop` and have the full toolchain in 30 seconds. Adding a `.envrc` for direnv users too."},
		}
	case "demo-022":
		return []message{
			{Role: "user", Text: "we're getting connection pool exhaustion under load — how do I tune pgxpool?"},
			{Role: "assistant", Text: "The defaults are conservative. For your workload I'd set `MaxConns` to `(num_cpus * 4)`, `MinConns` to `2`, `MaxConnLifetime` to 1 hour, and `MaxConnIdleTime` to 30 minutes. Also worth enabling `HealthCheckPeriod` at 1 minute so stale connections get recycled. I'll update `db/pool.go` with these values and add a comment explaining the reasoning."},
		}
	case "demo-023":
		return []message{
			{Role: "user", Text: "can you generate an OpenAPI spec automatically from our Go handler types?"},
			{Role: "assistant", Text: "I'll use `swaggo/swag` — you annotate handlers with structured comments and `swag init` generates the spec. I'll add the annotations to all existing handlers, wire up `/docs` to serve the Swagger UI, and add `make spec` to regenerate on demand. The spec will be committed to the repo so CI can diff it."},
			{Role: "user", Text: "make sure the auth endpoints document the bearer token requirement"},
			{Role: "assistant", Text: "Done. Added a `securityDefinitions.bearerAuth` scheme at the top level and annotated every protected endpoint with `@Security bearerAuth`. The Swagger UI now shows the Authorize button and all locked endpoints are marked with the padlock icon."},
		}
	case "demo-024":
		return []message{
			{Role: "user", Text: "add opentelemetry tracing — I want to see spans in jaeger locally"},
			{Role: "assistant", Text: "I'll set up the OTLP exporter pointing at a local Jaeger instance (already in docker-compose), initialise a `TracerProvider` in `main.go`, and instrument the HTTP middleware and DB layer with spans. Every request gets a root span with http method, route, and status code. DB spans include the query text (sanitised — no parameter values)."},
			{Role: "user", Text: "propagate the trace context to downstream gRPC calls too"},
			{Role: "assistant", Text: "Done. Added `otelgrpc.UnaryClientInterceptor()` to the gRPC client dial options. Trace context is now propagated via the `traceparent` metadata header, so the Jaeger UI shows the full call tree across both services."},
		}
	case "demo-025":
		return []message{
			{Role: "user", Text: "🦀 rewrite it in rust"},
			{Role: "assistant", Text: "I admire the commitment. Which part are we rewriting first — the whole thing, or just the hot path to justify the decision?"},
			{Role: "user", Text: "all of it obviously. the whole thing. every line"},
			{Role: "assistant", Text: "Bold. We're going to need a `Cargo.toml`, strongly held opinions about async runtimes, and at least three arguments about whether to use `Arc<Mutex<T>>` or channels. I'll start with the core types and we can fight about the rest later."},
			{Role: "user", Text: "use tokio, and make sure the borrow checker knows who's boss"},
			{Role: "assistant", Text: "The borrow checker IS the boss. That's the whole bit. But yes — tokio for the runtime, `async-trait` for the interfaces, `thiserror` for errors, and `serde` for everything serialisation-shaped. First pass compiles clean. Zero unsafe blocks. The borrow checker approves."},
			{Role: "user", Text: "how's the performance vs the old version?"},
			{Role: "assistant", Text: "Benchmarks are in. Throughput up 3.2x. Latency p99 down from 14 ms to 4 ms. Memory usage down 60%. Compile time up 40x. Worth it? Absolutely. Would we do it again? Ask me after the next dependency audit."},
		}
	default:
		return []message{}
	}
}

// demoFeaturedMessages returns the fun meta conversation for demo-001.
func demoFeaturedMessages() []message {
	return []message{
		{Role: "user", Text: "hey, I need to take a screenshot of lazyopencode for the README but I don't have any interesting sessions in my DB. can you generate some fake ones?"},
		{Role: "assistant", Text: "Oh this is very meta — you want me to fabricate a fake history of conversations so the screenshot looks lived-in. I'm into it. What kind of projects should I pretend we've been working on?"},
		{Role: "user", Text: "make it look realistic — a mix of feature work, bug fixes, some infra stuff. spread them across a few directories: ~/cool-project, ~/another-project, ~/Desktop. and no JS/TS, I hate that language"},
		{Role: "assistant", Text: "Excellent taste. Go, Rust, Python, Nix — nothing that requires a node_modules folder. Here's the lineup:\n\n  ~/cool-project: auth refactor, redis rate limiting, structured logging, benchmarking, LRU cache, OpenAPI spec, tracing, and this very session\n  ~/another-project: flaky CI fix, docker compose, prometheus metrics, gRPC tests, sync.Pool swap, graceful shutdown, fuzzing, postgres tuning\n  ~/Desktop: dark mode polish, Python-to-Go port, pre-commit hooks, Rust FFI bindings\n  ~: README update, Makefile from scratch, Nix flake\n\nThat's 24 sessions. I'll make the diffs look plausible too."},
		{Role: "user", Text: "yes please, and make the chat preview for this session extra fun — say hi to anyone who reads it in the screenshot"},
		{Role: "assistant", Text: "Obviously. Hi to everyone reading this in a screenshot 👋 You're looking at completely fabricated session data, lovingly handcrafted to make this README look impressive. Every session title, every diff stat, every message — made up. The irony is not lost on us.\n\nHere's what I'm touching to make this work:\n\n  • demo.go — new file, all fake data lives here (sessions, messages, stats)\n  • model.go — demoMode bool on the model struct\n  • model.go — loadSessionsCmd / loadMessagesCmd / loadStatsCmd branch on demoMode\n  • main.go — --demo flag wired to newModel\n\nZero real data. The perfect crime."},
		{Role: "user", Text: "love it — what files did you actually change?"},
		{Role: "assistant", Text: "Here's the full list:\n\n   1. demo.go (new) — demoSessions(), demoMessages(), demoStats(), demoFeaturedMessages()\n   2. model.go — added demoMode field to model struct\n   3. model.go — updated newModel() signature to accept demoMode bool\n   4. model.go — loadSessionsCmd() returns fake sessions when demoMode is true\n   5. model.go — loadMessagesCmd() returns fake messages when demoMode is true\n   6. model.go — loadStatsCmd() returns fake stats when demoMode is true\n   7. main.go — registered --demo flag\n   8. main.go — passed *demo to newModel()\n   9. demo.go — 24 fake sessions across 4 workspaces with realistic diffs\n  10. demo.go — this very conversation as the featured preview\n\nRun it with: go run . --demo"},
		{Role: "user", Text: "this is genuinely one of my favourite things I've asked an AI to do"},
		{Role: "assistant", Text: "Generating fake data for screenshots is an honourable and underappreciated art form. The screenshot will look great. Go take it."},
	}
}

// demoStats returns plausible sessionStats for --demo mode.
func demoStats(sessionID string) sessionStats {
	statsMap := map[string]sessionStats{
		"demo-001": {MsgCount: 10, InputTokens: 31200, OutputTokens: 4821, ContextTokens: 12400, Models: []string{"claude-sonnet-4-6"}},
		"demo-002": {MsgCount: 4, InputTokens: 8900, OutputTokens: 1203, ContextTokens: 3800, Models: []string{"claude-sonnet-4-6"}},
		"demo-003": {MsgCount: 2, InputTokens: 5100, OutputTokens: 892, ContextTokens: 2100, Models: []string{"claude-opus-4-6"}},
		"demo-004": {MsgCount: 4, InputTokens: 9800, OutputTokens: 1544, ContextTokens: 4200, Models: []string{"claude-sonnet-4-6"}},
		"demo-005": {MsgCount: 4, InputTokens: 11400, OutputTokens: 1876, ContextTokens: 5100, Models: []string{"claude-sonnet-4-6"}},
		"demo-006": {MsgCount: 3, InputTokens: 3900, OutputTokens: 634, ContextTokens: 1500, Models: []string{"claude-haiku-4-5-20251001"}},
		"demo-007": {MsgCount: 2, InputTokens: 4700, OutputTokens: 741, ContextTokens: 1900, Models: []string{"claude-sonnet-4-6"}},
		"demo-008": {MsgCount: 4, InputTokens: 7200, OutputTokens: 1102, ContextTokens: 2900, Models: []string{"claude-sonnet-4-6"}},
		"demo-009": {MsgCount: 4, InputTokens: 13100, OutputTokens: 2104, ContextTokens: 5600, Models: []string{"claude-opus-4-6", "claude-sonnet-4-6"}},
		"demo-010": {MsgCount: 2, InputTokens: 4300, OutputTokens: 688, ContextTokens: 1700, Models: []string{"claude-sonnet-4-6"}},
		"demo-011": {MsgCount: 2, InputTokens: 3100, OutputTokens: 519, ContextTokens: 1200, Models: []string{"claude-haiku-4-5-20251001"}},
		"demo-012": {MsgCount: 2, InputTokens: 3600, OutputTokens: 601, ContextTokens: 1400, Models: []string{"claude-sonnet-4-6"}},
		"demo-013": {MsgCount: 4, InputTokens: 8200, OutputTokens: 1388, ContextTokens: 3300, Models: []string{"claude-sonnet-4-6"}},
		"demo-014": {MsgCount: 4, InputTokens: 14800, OutputTokens: 2233, ContextTokens: 6100, Models: []string{"claude-opus-4-6"}},
		"demo-015": {MsgCount: 2, InputTokens: 2800, OutputTokens: 477, ContextTokens: 1100, Models: []string{"claude-sonnet-4-6"}},
		"demo-016": {MsgCount: 4, InputTokens: 10600, OutputTokens: 1699, ContextTokens: 4500, Models: []string{"claude-sonnet-4-6"}},
		"demo-017": {MsgCount: 2, InputTokens: 5000, OutputTokens: 812, ContextTokens: 2000, Models: []string{"claude-sonnet-4-6"}},
		"demo-018": {MsgCount: 2, InputTokens: 5900, OutputTokens: 934, ContextTokens: 2400, Models: []string{"claude-haiku-4-5-20251001"}},
		"demo-019": {MsgCount: 4, InputTokens: 6700, OutputTokens: 1055, ContextTokens: 2700, Models: []string{"claude-sonnet-4-6"}},
		"demo-020": {MsgCount: 4, InputTokens: 12200, OutputTokens: 1914, ContextTokens: 5200, Models: []string{"claude-sonnet-4-6"}},
		"demo-021": {MsgCount: 2, InputTokens: 4500, OutputTokens: 723, ContextTokens: 1800, Models: []string{"claude-opus-4-6"}},
		"demo-022": {MsgCount: 2, InputTokens: 3300, OutputTokens: 566, ContextTokens: 1300, Models: []string{"claude-sonnet-4-6"}},
		"demo-023": {MsgCount: 4, InputTokens: 11300, OutputTokens: 1801, ContextTokens: 4800, Models: []string{"claude-sonnet-4-6"}},
		"demo-024": {MsgCount: 4, InputTokens: 12900, OutputTokens: 2047, ContextTokens: 5500, Models: []string{"claude-sonnet-4-6"}},
		"demo-025": {MsgCount: 8, InputTokens: 74100, OutputTokens: 3102, ContextTokens: 89200, Models: []string{"claude-sonnet-4-6"}},
	}
	if s, ok := statsMap[sessionID]; ok {
		return s
	}
	return sessionStats{}
}

// demoGlobalStats returns plausible globalStats for --demo mode.
func demoGlobalStats() globalStats {
	return globalStats{
		TotalSessions:   25,
		TotalPrompts:    82,
		TotalInput:      312400,
		TotalOutput:     48210,
		TotalCacheRead:  93720,
		TotalCacheWrite: 62480,
		TotalFiles:      234,
		TotalAdditions:  12400,
		TotalDeletions:  4200,
		TotalDurationMS: 93_600_000, // 26h total (~1h 3m avg)

		RecentSessions:   4,
		RecentPrompts:    14,
		RecentInput:      38100,
		RecentOutput:     6820,
		RecentCacheRead:  11430,
		RecentCacheWrite: 7620,
		RecentFiles:      18,
		RecentAdditions:  820,
		RecentDeletions:  310,
		RecentDurationMS: 10_800_000, // 3h total (45m avg)

		Models: []modelStat{
			{Name: "claude-sonnet-4-6", Sessions: 18, Prompts: 54, InputTokens: 241300, OutputTokens: 37200, DurationMS: 64_800_000},
			{Name: "claude-opus-4-6", Sessions: 5, Prompts: 12, InputTokens: 58400, OutputTokens: 8900, DurationMS: 21_600_000},
			{Name: "claude-haiku-4-5-20251001", Sessions: 2, Prompts: 4, InputTokens: 12700, OutputTokens: 2110, DurationMS: 7_200_000},
		},

		Projects: []projectStat{
			{
				Dir: "/Users/demo/code/myapp", DisplayDir: "~/code/myapp",
				Sessions: 9, Prompts: 28, InputTokens: 142300, OutputTokens: 21400, DurationMS: 36_000_000,
				Models: []modelStat{
					{Name: "claude-sonnet-4-6", Sessions: 6, Prompts: 18, InputTokens: 98100, OutputTokens: 14800, DurationMS: 21_600_000},
					{Name: "claude-opus-4-6", Sessions: 3, Prompts: 8, InputTokens: 44200, OutputTokens: 6600, DurationMS: 14_400_000},
				},
			},
			{
				Dir: "/Users/demo/code/lazyopencode", DisplayDir: "~/code/lazyopencode",
				Sessions: 7, Prompts: 20, InputTokens: 98200, OutputTokens: 14700, DurationMS: 25_200_000,
				Models: []modelStat{
					{Name: "claude-sonnet-4-6", Sessions: 5, Prompts: 14, InputTokens: 71400, OutputTokens: 10900, DurationMS: 18_000_000},
					{Name: "claude-haiku-4-5-20251001", Sessions: 2, Prompts: 4, InputTokens: 26800, OutputTokens: 3800, DurationMS: 7_200_000},
				},
			},
			{
				Dir: "/Users/demo/code/api", DisplayDir: "~/code/api",
				Sessions: 5, Prompts: 14, InputTokens: 48100, OutputTokens: 7300, DurationMS: 18_000_000,
				Models: []modelStat{
					{Name: "claude-sonnet-4-6", Sessions: 5, Prompts: 12, InputTokens: 48100, OutputTokens: 7300, DurationMS: 18_000_000},
				},
			},
			{
				Dir: "/Users/demo/code/infra", DisplayDir: "~/code/infra",
				Sessions: 4, Prompts: 10, InputTokens: 23800, OutputTokens: 4810, DurationMS: 14_400_000,
				Models: []modelStat{
					{Name: "claude-opus-4-6", Sessions: 2, Prompts: 4, InputTokens: 14200, OutputTokens: 2900, DurationMS: 7_200_000},
					{Name: "claude-sonnet-4-6", Sessions: 2, Prompts: 6, InputTokens: 9600, OutputTokens: 1910, DurationMS: 7_200_000},
				},
			},
		},
	}
}
