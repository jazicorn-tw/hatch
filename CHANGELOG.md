# 📦 Release History

## 📦 Release 0.1.0

### ✨ Features

- **hooks:** use gum for output when available, fall back to printf (14b7022)
- implement M2 ingestion pipeline (d25be11)
- implement M3 quiz engine with LLM providers and CLI (d8a28fc)
- **ingest:** add Gemini embedder and split dev into .dev/ modules (08ac3a7)
- **kata:** implement M3b kata engine with multi-language sandbox and TUI editor (3ecb1e1)
- scaffold M1 Go packages, config, store, and test infrastructure (a028a87)

### 🐛 Fixes

- **build:** add cmd/hatch entry point hidden by gitignore (1f51af0)
- **ci:** remove leftover Spring Boot gradle build step from releaserc (6f6331b)
- **docker:** add sqlite-dev to xx-apk for sqlite-vec CGO headers (9a8a3e0)
- **docker:** pin tonistiigi/xx to digest and enable CGO cross-compilation (d239d4a)
- **docker:** switch builder to debian bookworm for sqlite-vec CGO compatibility (e260622)
- **docker:** use xx canonical COPY pattern and pin image to digest (6dd23eb)
- **hooks:** add gum output to commit-msg scope validation (dd0d133)
- **hooks:** skip tag validation for template files in pre-add (f915de8)
- **kata:** resolve sandbox commands to absolute paths via exec.LookPath (334bada)
- make cz optional in commit-msg hook and fix doctor false positives (81f5b30)
- **make:** guard format, lint, and test targets when no Go files exist ⚠️ unknown-scope (67f3ac0)
- **qa:** resolve code smells and optimize data structures (36beae7)
- **qa:** resolve remaining SonarQube issues in semantic-release-impact.mjs (2472eea)
- **qa:** resolve SonarQube issues across workflows, Helm, Go, and scripts (8a0cc88)
- **tooling:** detect and recover stale Colima state on env up (f33ea72)
- **tooling:** remove duplicate release-notes-generator and sync scopes from tags.yml (d16b608)

### 📦 Build

- add Dockerfile and Helm chart (02c0a26)

### 🤖 CI / CD

- **actions:** add Dependabot for automated dependency updates (cd5447a)
- **actions:** normalize variable comparisons to uppercase (235bae3)
- **actions:** pin all GitHub Actions to full commit SHAs (4534e3f)
- **deploy:** add validate-helm gate to ensure both artifacts publish together (28c71a3)
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
- **deps:** update yarn.lock for js-yaml (413972e)
- **dev:** add tags command and extend tag taxonomy (464d7d0)
- **dev:** fix test-ci usage comment (6b78c43)
- initialize go module (github.com/jazicorn/hatch) (c12a106)
- migrate make/ and scripts from Java/Gradle/act to Go (b2d5df9)
- move tags.yml to .github/, add yarn node-modules linker (8cfd0d8)
- sync scope validation with tags.yml in releaserc (f30b969)
- **tooling:** remove stale inspect scripts (b517bc1)
- update pre-commit hook and local-settings for Go stack (acf8b92)

### 💄 Style

- **store:** gofmt alignment on minHeap method signatures (814b81c)

### ♻️ Refactors

- **kata:** extract shared generator helpers into internal/genutil (e9a9c9f)
- replace Makefile with gum-powered dev script (bd8edb7)

### 📝 Docs

- add commit, devops, and tooling documentation (91dea7d)
- add deeper onboarding docs, Charmbracelet ecosystem, and ADR-014 (692373a)
- add M2 ingestion pipeline doc and mark complete in roadmap (1b2004d)
- add M3 quiz engine milestone doc (d0ae8d3)
- **adr:** add ADR-012 sqlite-vec over dedicated vector databases (c01096e)
- **adr:** add ADR-013 Go as implementation language (586ef1b)
- **adr:** add architecture decision records for Go stack (f0a2395)
- **adr:** align adrs with current project state and roadmap (b1259f5)
- **adr:** fill Scope and Deciders fields across all ADRs (8f97937)
- **commit:** update cheat sheet and commit docs (3a4ab8d)
- **devops:** add git and github operation guides (657d7d6)
- **devops:** reorganize release and status docs into subdirectories (4f3dce8)
- **onboarding:** add onboarding guides and move project setup (2a3ecd1)
- **planning:** mark M3 quiz engine complete in roadmap (d34cbc4)
- **providers:** add Gemini provider support and 1Password integration guide (f8e3abb)
- split TESTING.md, rename EMBEDDINGS to EMBEDDER, update references (52fbf36)
- update roadmap, docs template, and tooling docs (00c777d)
- update roadmap, readme, env examples, and CI docs (f3fa6b1)
- update stale Java/Spring references for Go stack (3358225)
