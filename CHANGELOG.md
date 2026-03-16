# 📦 Release History

## 📦 Release 0.5.0

### ✨ Features

- **kata:** implement Sr-provided quiz and kata import commands (ac02ef2)

### 🤖 CI / CD

- **actions:** add missing permissions to validate-helm job (940a729)
- **actions:** gate releases on SonarCloud quality gate (da4b067)
- **actions:** opt into Node.js 24 runner and upgrade Node to 22 (0d1e398)
- **actions:** replace deprecated sonarcloud-github-action with sonarqube-scan-action v7.0.0 (fb42464)

### ♻️ Refactors

- **kata:** extract setupStore helper to eliminate duplicate store-opening code (59ee69d)
- **qa:** reduce duplicated lines across cmd, internal/genutil, and fs tests (2e05d7f)

### 📝 Docs

- **ci:** document SonarCloud quality gate setup and CI job table (2c7fd7c)

## 📦 Release 0.4.0

### ✨ Features

- **kata:** implement M3b kata engine with multi-language sandbox and TUI editor (3ecb1e1)

### 🐛 Fixes

- **kata:** resolve sandbox commands to absolute paths via exec.LookPath (334bada)
- **tooling:** remove duplicate release-notes-generator and sync scopes from tags.yml (d16b608)

### 🧹 Chores

- **deps:** update yarn.lock for js-yaml (413972e)
- sync scope validation with tags.yml in releaserc (f30b969)

### ♻️ Refactors

- **kata:** extract shared generator helpers into internal/genutil (e9a9c9f)

### 📝 Docs

- **planning:** mark M3 quiz engine complete in roadmap (d34cbc4)



## 📦 Release 0.3.0

### ✨ Features

- implement M3 quiz engine with LLM providers and CLI (9666574)

### 📝 Docs

- **adr:** align adrs with current project state and roadmap (96a452d)



## 📦 Release 0.2.0

### ✨ Features

- implement M2 ingestion pipeline (2046fc3)
- **ingest:** add Gemini embedder and split dev into .dev/ modules (97c5923)

### 🐛 Fixes

- **docker:** add sqlite-dev to xx-apk for sqlite-vec CGO headers (aaaaae5)
- **docker:** pin tonistiigi/xx to digest and enable CGO cross-compilation (8bf8768)
- **docker:** switch builder to debian bookworm for sqlite-vec CGO compatibility (661f353)
- **docker:** use xx canonical COPY pattern and pin image to digest (0fd0d35)
- **qa:** resolve code smells and optimize data structures (0982ec8)

### 🤖 CI / CD

- **deploy:** add validate-helm gate to ensure both artifacts publish together (bac747c)

### 🧹 Chores

- **dev:** add tags command and extend tag taxonomy (b365cf6)

### 📝 Docs

- add deeper onboarding docs, Charmbracelet ecosystem, and ADR-014 (7753e27)
- add M2 ingestion pipeline doc and mark complete in roadmap (845a18b)
- add M3 quiz engine milestone doc (80f3491)
- **adr:** add ADR-012 sqlite-vec over dedicated vector databases (c24aa76)
- **adr:** add ADR-013 Go as implementation language (4c424d2)
- **devops:** add git and github operation guides (771272f)
- **providers:** add Gemini provider support and 1Password integration guide (0fa5d23)
- split TESTING.md, rename EMBEDDINGS to EMBEDDER, update references (c3c3671)



## 📦 Release 0.1.2

### 🐛 Fixes

- **qa:** resolve remaining SonarQube issues in semantic-release-impact.mjs (50e9ca7)

### 🤖 CI / CD

- **actions:** add Dependabot for automated dependency updates (a475b37)



## 📦 Release 0.1.1

### 🐛 Fixes

- **qa:** resolve SonarQube issues across workflows, Helm, Go, and scripts (1b931cf)

### 🤖 CI / CD

- **actions:** pin all GitHub Actions to full commit SHAs (94fd7df)



## 📦 Release 0.1.0

### ✨ Features

- **hooks:** use gum for output when available, fall back to printf (14b7022)
- scaffold M1 Go packages, config, store, and test infrastructure (a028a87)

### 🐛 Fixes

- **build:** add cmd/hatch entry point hidden by gitignore (7388bb8)
- **ci:** remove leftover Spring Boot gradle build step from releaserc (6f6331b)
- **hooks:** add gum output to commit-msg scope validation (dd0d133)
- **hooks:** skip tag validation for template files in pre-add (f915de8)
- make cz optional in commit-msg hook and fix doctor false positives (81f5b30)
- **make:** guard format, lint, and test targets when no Go files exist (67f3ac0)
- **tooling:** detect and recover stale Colima state on env up (f33ea72)

### 📦 Build

- add Dockerfile and Helm chart (bd6a7db)

### 🤖 CI / CD

- **actions:** normalize variable comparisons to uppercase (235bae3)
- gate go vet and staticcheck behind ENABLE_GO_ANALYSIS (ddf5557)
- remove pull_request trigger from doctor workflow (f151875)

### 🧹 Chores

- add example env, vars, and secrets template files (9a25938)
- add local-settings config examples and inspect script (58cb49b)
- add scripts infrastructure and configurable Colima doctor gate (f4931b1)
- **ci:** add GitHub Actions workflows (f4ae4f3)
- **commit:** update scope taxonomy and commitizen config (0e14960)
- configure semantic-release and add local release tooling (aa5ffc5)
- configure Yarn 4 and expand .gitignore (f13330d)
- **dev:** fix test-ci usage comment (6b78c43)
- initialize go module (github.com/jazicorn/hatch) (c12a106)
- migrate make/ and scripts from Java/Gradle/act to Go (b2d5df9)
- move tags.yml to .github/, add yarn node-modules linker (8cfd0d8)
- **release:** 0.1.0 (6e1b9ab)
- **release:** 0.1.0 (28de9a7)
- **tooling:** remove stale inspect scripts (b517bc1)
- update pre-commit hook and local-settings for Go stack (acf8b92)

### 💄 Style

- **store:** gofmt alignment on minHeap method signatures (814b81c)

### ♻️ Refactors

- replace Makefile with gum-powered dev script (bd8edb7)

### 📝 Docs

- add commit, devops, and tooling documentation (91dea7d)
- **adr:** add architecture decision records for Go stack (f0a2395)
- **adr:** fill Scope and Deciders fields across all ADRs (8f97937)
- **commit:** update cheat sheet and commit docs (3a4ab8d)
- **devops:** reorganize release and status docs into subdirectories (4f3dce8)
- **onboarding:** add onboarding guides and move project setup (2a3ecd1)
- update roadmap, docs template, and tooling docs (00c777d)
- update roadmap, readme, env examples, and CI docs (f3fa6b1)
- update stale Java/Spring references for Go stack (3358225)
