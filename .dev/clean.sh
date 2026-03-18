# .dev/clean.sh — Clean tasks: clean:local, clean:docker, clean:colima
# Sourced by ./dev — not executed directly.

run_clean_local() {
  header "clean local"
  $GUM confirm "Clean local build artifacts?" || exit 0
  ./scripts/cache/clean-local.sh
  log_done "clean local"
  return 0
}

run_clean_docker() {
  header "clean docker"
  $GUM confirm "Clean Docker build cache?" || exit 0
  ./scripts/cache/cache-docker.sh
  log_done "clean docker"
  return 0
}

run_clean_colima() {
  header "clean colima"
  $GUM confirm "Reset Colima? This will delete all containers and volumes." || exit 0
  ./scripts/cache/clean-colima.sh
  log_done "clean colima"
  return 0
}
