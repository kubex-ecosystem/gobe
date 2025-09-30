#!/usr/bin/env bash
# shellcheck disable=SC2015,SC1091

# Script Metadata
__secure_logic_version="1.0.0"
__secure_logic_date="$( date +%Y-%m-%d )"
__secure_logic_author="Rafael Mori"
__secure_logic_use_type="exec"
__secure_logic_init_timestamp="$(date +%s)"
__secure_logic_elapsed_time=0

set -o errexit # Exit immediately if a command exits with a non-zero status
set -o nounset # Treat unset variables as an error when substituting
set -o pipefail # Return the exit status of the last command in the pipeline that failed
set -o errtrace # If a command fails, the shell will exit immediately
set -o functrace # If a function fails, the shell will exit immediately
shopt -s inherit_errexit # Inherit the errexit option in functions

# Get the root directory of the git project
_SCRIPT_DIR="$(git rev-parse --show-toplevel)"
cd "$_SCRIPT_DIR" || exit 1

echo "🚀 Configurando pre-commit hooks..."

_default_pre_commit_config() {
  # Create support/pre-commit-config.yaml if it doesn't exist
  if [[ ! -f support/pre-commit-config.yaml ]]; then
    echo "🛠️  Creating support/pre-commit-config.yaml..."
    touch support/pre-commit-config.yaml
    # shellcheck disable=SC2155
    local _DEFAULT_PRE_COMMIT_CONFIG=$(cat <<'EOF'
# Pre-commit configuration file
# Documentation: https://pre-commit.com/
repos:
  # -------- Hygiene básica --------
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-json
      - id: check-added-large-files
        args: ["--maxkb=1000"]

  # -------- Segurança: detect-secrets --------
  - repo: https://github.com/Yelp/detect-secrets
    rev: v1.4.0
    hooks:
      - id: detect-secrets
        args: ["--baseline", ".secrets.baseline"]

  # -------- Segurança: gitleaks --------
  - repo: https://github.com/zricethezav/gitleaks
    rev: v8.17.0
    hooks:
      - id: gitleaks
        name: gitleaks
        entry: gitleaks protect --staged --no-banner
        language: system
        pass_filenames: false

  # -------- Go tools --------
  - repo: https://github.com/dnephin/pre-commit-golang
    rev: v0.5.1
    hooks:
      - id: go-fmt
      - id: go-vet
      - id: go-mod-tidy
      - id: golangci-lint
        args: ["--fast"]

  # # -------- Python tools --------
  # - repo: https://github.com/pre-commit/mirrors-autopep8
  #   rev: v1.5.7
  #   hooks:
  #     - id: autopep8
  #       args: ["--aggressive", "--aggressive"]
  # - repo: https://github.com/pre-commit/mirrors-isort
  #   rev: v5.10.1
  #   hooks:
  #     - id: isort
  #       args: ["--profile", "black"]
  # - repo: https://github.com/pre-commit/mirrors-black
  #   rev: 22.3.0
  #   hooks:
  #     - id: black
  #       args: ["--line-length", "88"]

EOF
)
  echo "$_DEFAULT_PRE_COMMIT_CONFIG"
  # else
  #   cat support/pre-commit-config.yaml
  fi
}

_install_pre_commit_tools() {
  # Create and activate a virtual environment for hooks
  if [[ ! -d .venv-hooks ]]; then
    python3 -m venv .venv-hooks
  fi
  if [[ ! -f .venv-hooks/bin/activate ]]; then
    echo "❌ Falha ao encontrar o ambiente virtual em .venv-hooks"
    exit 1
  fi

  # shellcheck source=/dev/null
  . .venv-hooks/bin/activate

  # Install requirements file from support/ if it exists
  if [[ -f support/requirements-hooks.txt ]]; then
    pip install -r support/requirements-hooks.txt
  fi

  pip install -U pip setuptools wheel
  pip install pre-commit detect-secrets

  # Install pre-commit hooks
  pre-commit install --config support/pre-commit-config.yaml --install-hooks
}

_create_baseline() {
  # Create a baseline for detect-secrets if it doesn't exist
  if [[ ! -f .secrets.baseline ]]; then
    detect-secrets scan > .secrets.baseline
    git add .secrets.baseline
    git commit -m "chore(secrets): add baseline" || true
  fi
}

_main() {
  # First we check if pre-commit is already configured
  if git config --get core.hooksPath &>/dev/null; then
    echo "⚠️  Pre-commit hooks are already configured. Aborting..."
    return 0
  fi

  _default_pre_commit_config > support/pre-commit-config.yaml
  _install_pre_commit_tools
  _create_baseline
  echo "✅ Pre-commit hooks configured successfully!"
}

_main "$@"

__secure_logic_elapsed_time=$(( $(date +%s) - __secure_logic_init_timestamp ))
echo "⏱️  Script executed in $__secure_logic_elapsed_time seconds."
