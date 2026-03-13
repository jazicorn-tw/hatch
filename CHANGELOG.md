# 📦 Release History

## 📦 Release 0.1.0

### ✨ Features

- **hooks:** use gum for output when available, fall back to printf (14b7022)
- scaffold M1 Go packages, config, store, and test infrastructure (a028a87)

### 🐛 Fixes

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



## 0.1.0

### ✨ Features

- **hooks:** use gum for output when available, fall back to printf
- scaffold M1 Go packages, config, store, and test infrastructure

### 🐛 Fixes

- **ci:** remove leftover Spring Boot gradle build step from releaserc
- **hooks:** add gum output to commit-msg scope validation
- **hooks:** skip tag validation for template files in pre-add
- make cz optional in commit-msg hook and fix doctor false positives
- **make:** guard format, lint, and test targets when no Go files exist
- **tooling:** detect and recover stale Colima state on env up

### 📦 Build

- add Dockerfile and Helm chart

### 🤖 CI / CD

- **actions:** normalize variable comparisons to uppercase
- gate go vet and staticcheck behind ENABLE_GO_ANALYSIS
- remove pull_request trigger from doctor workflow

### 🧹 Chores

- add example env, vars, and secrets template files
- add local-settings config examples and inspect script
- add scripts infrastructure and configurable Colima doctor gate
- **ci:** add GitHub Actions workflows
- **commit:** update scope taxonomy and commitizen config
- configure semantic-release and add local release tooling
- configure Yarn 4 and expand .gitignore
- **dev:** fix test-ci usage comment
- initialize go module (github.com/jazicorn/hatch)
- migrate make/ and scripts from Java/Gradle/act to Go
- move tags.yml to .github/, add yarn node-modules linker
- **release:** 0.1.0
- **tooling:** remove stale inspect scripts
- update pre-commit hook and local-settings for Go stack

### 💄 Style

- **store:** gofmt alignment on minHeap method signatures

### ♻️ Refactors

- replace Makefile with gum-powered dev script

### 📝 Docs

- add commit, devops, and tooling documentation
- **adr:** add architecture decision records for Go stack
- **adr:** fill Scope and Deciders fields across all ADRs
- **commit:** update cheat sheet and commit docs
- **devops:** reorganize release and status docs into subdirectories
- **onboarding:** add onboarding guides and move project setup
- update roadmap, docs template, and tooling docs
- update roadmap, readme, env examples, and CI docs
- update stale Java/Spring references for Go stack

## 📦 Release 0.1.0

### ✨ Features

- **hooks:** use gum for output when available, fall back to printf (14b7022)
- scaffold M1 Go packages, config, store, and test infrastructure (a028a87)

### 🐛 Fixes

- **ci:** remove leftover Spring Boot gradle build step from releaserc (6f6331b)
- **hooks:** add gum output to commit-msg scope validation (dd0d133)
- **hooks:** skip tag validation for template files in pre-add (f915de8)
- make cz optional in commit-msg hook and fix doctor false positives (81f5b30)
- **make:** guard format, lint, and test targets when no Go files exist (67f3ac0)
- **tooling:** detect and recover stale Colima state on env up (f33ea72)

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



## 0.1.0

### ✨ Features

- **hooks:** use gum for output when available, fall back to printf
- scaffold M1 Go packages, config, store, and test infrastructure

### 🐛 Fixes

- **ci:** remove leftover Spring Boot gradle build step from releaserc
- **hooks:** add gum output to commit-msg scope validation
- **hooks:** skip tag validation for template files in pre-add
- make cz optional in commit-msg hook and fix doctor false positives
- **make:** guard format, lint, and test targets when no Go files exist
- **tooling:** detect and recover stale Colima state on env up

### 🤖 CI / CD

- **actions:** normalize variable comparisons to uppercase
- gate go vet and staticcheck behind ENABLE_GO_ANALYSIS
- remove pull_request trigger from doctor workflow

### 🧹 Chores

- add example env, vars, and secrets template files
- add local-settings config examples and inspect script
- add scripts infrastructure and configurable Colima doctor gate
- **ci:** add GitHub Actions workflows
- **commit:** update scope taxonomy and commitizen config
- configure semantic-release and add local release tooling
- configure Yarn 4 and expand .gitignore
- **dev:** fix test-ci usage comment
- initialize go module (github.com/jazicorn/hatch)
- migrate make/ and scripts from Java/Gradle/act to Go
- move tags.yml to .github/, add yarn node-modules linker
- **tooling:** remove stale inspect scripts
- update pre-commit hook and local-settings for Go stack

### 💄 Style

- **store:** gofmt alignment on minHeap method signatures

### ♻️ Refactors

- replace Makefile with gum-powered dev script

### 📝 Docs

- add commit, devops, and tooling documentation
- **adr:** add architecture decision records for Go stack
- **adr:** fill Scope and Deciders fields across all ADRs
- **commit:** update cheat sheet and commit docs
- **devops:** reorganize release and status docs into subdirectories
- **onboarding:** add onboarding guides and move project setup
- update roadmap, docs template, and tooling docs
- update roadmap, readme, env examples, and CI docs
- update stale Java/Spring references for Go stack
