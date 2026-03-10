# -----------------------------------------------------------------------------
# 30-interface.mk (30s — Interface)
#
# Responsibility: Public Makefile API discoverability.
# - help output, usage patterns, docs pointers
#
# Rule: This is the CLI contract. Keep stable.
# -----------------------------------------------------------------------------

# -------------------------------------------------------------------
# HELP / DOCS
# -------------------------------------------------------------------

.PHONY: help help-short help-auto explain debug

help: ## 🧰 Show developer help (curated)
	$(call section,🧰  {{project-name}} — Make Targets)

	$(call println,$(YELLOW)🚀 Recommended flow$(RESET))
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make help-categories" "→ discover help by category"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make help-roles" "→ discover role entrypoints"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make env-init" "→ create .env from example"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make bootstrap" "→ first-time setup (dev)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make verify" "→ before pushing"
	$(call println,)

	$(call println,$(YELLOW)🧑‍💼 Roles (opinionated entrypoints)$(RESET))
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make contributor" "→ run PR-ready checks (verify)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make reviewer" "→ CI-parity checks (quality)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make maintainer" "→ full local confidence (quality)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make help-roles" "→ explain roles and expectations"
	$(call println,)

	$(call println,$(YELLOW)🧪 Quality gates$(RESET))
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make doctor" "→ local environment sanity checks"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make check-env" "→ verify required env file (.env)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make env-init" "→ init baseline env from examples (safe)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make env-init-force" "→ overwrite baseline env from examples ($(RED)⚠️ destructive$(RESET))"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make env-help" "→ docs: local environment setup"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "go vet ./..." "→ static analysis (fast)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make test" "→ unit tests"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make verify" "→ doctor + lint + test"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make quality" "→ doctor + vet + test"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make pre-commit" "→ smart gate (main strict, branches fast)"
	$(call println,)

	$(call println,$(YELLOW)🐳 Docker / DB$(RESET))
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make docker-up" "→ start local Docker Compose services"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make docker-down" "→ stop local Docker Compose services"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make docker-reset" "→ stop + delete volumes + restart"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make run" "→ start the server (loads .env, runs go run)"
	$(call println,)

	$(call println,$(YELLOW)🧼 Local hygiene (disk pressure relief)$(RESET))
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make clean-local-info" "→ snapshot (docker + colima status)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make clean-local" "→ docker hygiene (Colima reset is explicit)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make clean-docker" "→ docker prune (explicit; supports auto mode)"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make docker-cache-info" "→ docker disk usage breakdown"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make colima-info" "→ show colima status"
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make clean-colima" "→ reset colima VM ($(RED)☢️ nuclear$(RESET))"
	@printf "%b\n" "$(GRAY)Docs: docs/tooling/LOCAL_HYGIENE.md$(RESET)"
	$(call println,)

	$(call println,$(YELLOW)🧭 Inspection / Navigation$(RESET))
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make tree [path]" "→ inspect repo structure (read-only). Docs: docs/TREE.md"
	$(call println,)

	$(call println,$(YELLOW)📦 Delivery$(RESET))
	@printf "  $(BOLD)%-22s$(RESET) %s\n" "make release-dry-run" "→ preview next semantic-release version (no publish)"
	$(call println,)

	$(call println,$(GRAY)Discover more: make help-categories | make help-roles$(RESET))
	$(call println,)

help-short: ## 🧰 Quick help (minimal)
	$(call section,🧰  Quick Make Targets)
	@printf "  $(BOLD)%-16s$(RESET) %s\n" "help" "curated help (recommended)"
	@printf "  $(BOLD)%-16s$(RESET) %s\n" "help-categories" "discover help by category"
	@printf "  $(BOLD)%-16s$(RESET) %s\n" "help-roles" "discover help by role (contributor/reviewer/maintainer)"
	@printf "  $(BOLD)%-16s$(RESET) %s\n" "contributor" "role gate: run PR-ready checks"
	@printf "  $(BOLD)%-16s$(RESET) %s\n" "doctor" "local environment sanity checks"
	@printf "  $(BOLD)%-16s$(RESET) %s\n" "verify" "doctor + lint + test"
	@printf "  $(BOLD)%-16s$(RESET) %s\n" "quality" "CI-parity gate"
	@printf "  $(BOLD)%-16s$(RESET) %s\n" "clean-local" "local disk hygiene (docker)"
	$(call println,)

help-auto: ## 🧾 Auto-generated help (from ## comments)
	$(call section,🧾  Auto-generated help)
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_.-]+:.*## / {printf "  $(BOLD)%-24s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	$(call println,)

explain: ## 🧠 Explain a target: make explain <target>
	@t="$(word 2,$(MAKECMDGOALS))"; \
	if [[ -z "$$t" ]]; then \
	  printf "%b\n" "$(RED)❌ Usage: make explain <target>$(RESET)"; \
	  printf "%b\n" "$(GRAY)Try one of: doctor check-env env-init env-init-force env-help bootstrap verify quality pre-commit clean-local clean-docker clean-colima$(RESET)"; \
	  exit 1; \
	fi; \
	$(call section,🧠  explain → $${t}); \
		case "$$t" in \
	  doctor)  printf "%b\n" "  $(BOLD)doctor$(RESET): runs local sanity checks (docker, colima, env files)";; \
	  check-env) printf "%b\n" "  $(BOLD)check-env$(RESET): verifies required baseline env file (.env)";; \
	  env-init) printf "%b\n" "  $(BOLD)env-init$(RESET): create .env from example file (safe, non-destructive)";; \
	  env-init-force) printf "%b\n" "  $(BOLD)env-init-force$(RESET): overwrite .env from example ($(RED)⚠️ destructive$(RESET))";; \
	  env-help) printf "%b\n" "  $(BOLD)env-help$(RESET): prints link to local environment setup documentation";; \
	  bootstrap) printf "%b\n" "  $(BOLD)bootstrap$(RESET): env-init + hooks + exec-bits + full quality gate (first-time dev setup)";; \
	  verify)  printf "%b\n" "  $(BOLD)verify$(RESET): doctor + lint + test (recommended before pushing)";; \
	  quality) printf "%b\n" "  $(BOLD)quality$(RESET): doctor + go vet + go test (matches CI intent)";; \
	  pre-commit) printf "%b\n" "  $(BOLD)pre-commit$(RESET): smart gate (main → strict CI parity, branches → faster checks)";; \
	  clean-local) printf "%b\n" "  $(BOLD)clean-local$(RESET): local disk hygiene (docker prune). Colima reset is explicit via clean-colima";; \
	  clean-docker) printf "%b\n" "  $(BOLD)clean-docker$(RESET): docker prune (explicit opt-in; supports auto mode keyed off Colima containerd filesystem)";; \
	  clean-colima) printf "%b\n" "  $(BOLD)clean-colima$(RESET): reset Colima VM ($(RED)☢️ nuclear$(RESET)); interactive confirmation required";; \
	  *) \
	    printf "%b\n" "$(YELLOW)⚠️  No extended explanation available for '$$t'.$(RESET)"; \
	    printf "%b\n" "$(GRAY)Tip: try 'make help', 'make help-categories', or 'make help-roles'.$(RESET)"; \
	    printf "%b\n" "$(GRAY)Docs: docs/tooling/MAKEFILE.md$(RESET)"; \
	    ;; \
	esac;
	$(call println,)
