# -----------------------------------------------------------------------------
# 60-verification.mk (60s — Build & Verification)
#
# Responsibility: Prove code correctness.
# - format/lint/tests/static analysis/coverage
#
# Rule: Deterministic. Do not mutate machine state.
# -----------------------------------------------------------------------------

# -------------------------------------------------------------------
# QUALITY / TESTS / BOOTSTRAP
# -------------------------------------------------------------------

.PHONY: pre-commit format lint lint-docs test verify quality test-ci bootstrap

pre-commit: ## 🪝 Smart pre-commit gate (strict on main)
	@if [ "$(GIT_BRANCH)" = "main" ]; then \
	  printf "%b\n" "$(CYAN)🪝 pre-commit$(RESET): on '$(BOLD)main$(RESET)' → running $(BOLD)quality$(RESET)"; \
	  $(MAKE) quality; \
	else \
	  printf "%b\n" "$(CYAN)🪝 pre-commit$(RESET): on '$(BOLD)$(GIT_BRANCH)$(RESET)' → running fast gate ($(BOLD)format + lint + test$(RESET))"; \
	  $(MAKE) format lint test; \
	fi

format: ## ✨ Auto-format sources (gofmt)
	$(call group_start,format)
	$(call step,✨ gofmt)
	@if [ -n "$$(find . -name '*.go' -not -path './vendor/*' 2>/dev/null | head -1)" ]; then \
	  gofmt -w ./...; \
	else \
	  printf "%b\n" "$(YELLOW)⚠ no Go files found, skipping gofmt$(RESET)"; \
	fi
	$(call group_end)

lint: lint-docs ## 🔎 Static analysis + markdown lint
	$(call group_start,lint)
	$(call step,🔎 go vet)
	@if [ -n "$$(find . -name '*.go' -not -path './vendor/*' 2>/dev/null | head -1)" ]; then \
	  go vet ./...; \
	else \
	  printf "%b\n" "$(YELLOW)⚠ no Go files found, skipping go vet$(RESET)"; \
	fi
	$(call group_end)

lint-docs: ## 📝 Lint all markdown files (markdownlint-cli2)
	$(call group_start,lint-docs)
	$(call step,📝 markdownlint)
	@./node_modules/.bin/markdownlint-cli2 '**/*.md' '#node_modules'
	$(call group_end)

test: ## 🧪 Unit tests
	$(call group_start,test)
	$(call step,🧪 Unit tests)
	@if [ -n "$$(find . -name '*.go' -not -path './vendor/*' 2>/dev/null | head -1)" ]; then \
	  go test ./...; \
	else \
	  printf "%b\n" "$(YELLOW)⚠ no Go files found, skipping go test$(RESET)"; \
	fi
	$(call group_end)

verify: doctor lint test ## ✅ Doctor + lint + test
	@printf "%b\n" "$(GREEN)✅ verify complete$(RESET)"

quality: doctor ## ✅ Doctor + go vet + go test (matches CI intent)
	$(call group_start,quality)
	$(call step,✅ CI-parity quality gate)
	@go vet ./...
	@go test ./...
	$(call group_end)

test-ci: ## CI: Run CI-equivalent test suite locally
	$(call group_start,test-ci)
	$(call step,🧪 CI-like test run)
	@go test -count=1 -race ./...
	$(call group_end)

bootstrap: env-init hooks exec-bits quality ## 🚀 Install env + hooks + run full local quality gate
	$(call step,🚀 bootstrap complete)
	@printf "%b\n" "$(GREEN)✅ bootstrap complete$(RESET)"
