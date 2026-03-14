# .dev/docker.sh — Docker tasks: docker:up, docker:down, docker:reset
# Sourced by ./dev — not executed directly.

run_docker_up() {
  header "docker up"
  spin "Starting Docker Compose..." docker compose up -d
  log_done "docker up"
}

run_docker_down() {
  header "docker down"
  spin "Stopping Docker Compose..." docker compose down
  log_done "docker down"
}

run_docker_reset() {
  header "docker reset"
  $GUM confirm "Reset Docker Compose? This will delete volumes." || exit 0
  spin "Resetting Docker Compose..." bash -c "docker compose down -v && docker compose up -d"
  log_done "docker reset"
}
