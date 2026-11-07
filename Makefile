# Go 1.27.0 lives at ~/.local/go and is NOT on the default PATH.
#   export PATH="$HOME/.local/go/bin:$PATH"
#
# The contract (`make test`): every answer under testdata/ was produced by
# executing Node (oracle/generate.mjs), and the package must give each one,
# except the lines in testdata/known-differences.txt (README).
# The watch (`make watch`): the same generator on Bun (Envoy's runtime) and on
# whatever Node is installed, compared the same way; it reports and never
# fails. WATCH.txt is its last report.

GO    ?= go
NICE  ?= nice -n 10
CORES ?= 2

.PHONY: all fmt fmt-check vet test goldens bun-goldens watch known-differences

all: fmt-check vet test watch

fmt:
	gofmt -w .

fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "unformatted:"; echo "$$out"; exit 1; fi

vet:
	CGO_ENABLED=0 $(GO) vet ./...

test:
	GOMAXPROCS=$(CORES) CGO_ENABLED=1 $(NICE) $(GO) test -race -count=1 ./...

# Goldens come from EXECUTING the runtime. Never hand-edit testdata/.
goldens:
	$(NICE) node oracle/generate.mjs

# RuntimeBun's contract (TestBunContract): Bun's own answers, from Bun.
bun-goldens:
	$(NICE) bun oracle/bun.mjs

known-differences:
	-KNOWN_DIFFERENCES_WRITE=1 CGO_ENABLED=0 $(GO) test -count=1 -run '^TestContract$$' . > /dev/null
	@wc -l testdata/known-differences.txt

watch:
	@rm -rf .watch && mkdir -p .watch/bun .watch/node
	@$(NICE) bun oracle/generate.mjs .watch/bun 2>/dev/null
	@$(NICE) node oracle/generate.mjs .watch/node 2>/dev/null
	@: > WATCH.txt
	@for r in node bun; do \
		echo "== $$r" >> WATCH.txt; \
		WATCH_DIR=.watch/$$r WATCH_REPORT=.watch/$$r.txt CGO_ENABLED=0 $(GO) test -count=1 -run '^TestWatch$$' . > /dev/null; \
		cat .watch/$$r.txt >> WATCH.txt; \
	done
	@rm -rf .watch
	@grep 'answers compared' WATCH.txt
