#!/usr/bin/env bash


###################################################################################
# __env_sourced_name
#
# This function is used to generate a unique environment variable name
# based on the script name. It will be used to check if the script is sourced
# after all other operations are done and validated. This will prevent
# the script from being run directly and will ensure that the environment
# is set up correctly before running any commands. It is not meant to be run
# directly.
#
###################################################################################
__cleanner_sourced_name() {
  local _self="${BASH_SOURCE-}"
  _self="${_self//${_kbx_root:-$()}/}"
  _self="${_self//\.sh/}"
  _self="${_self//\-/_}"
  _self="${_self//\//_}"
  echo "_was_sourced_${_self//__/_}"
  return 0
}

###############################################################################
# logger.sh load
#
# This line loads the logger.sh script only if the function kbx_log is not already
# defined. This is to prevent loading the script multiple times and causing
# conflicts. The logger.sh script is used to kbx_log messages to the console and
# to a file. It is not meant to be run directly.
# 
#################################################################################
# shellcheck disable=SC2065,SC1091
test -z "$(declare -f kbx_log)" >/dev/null && source "${_kbx_path_helpers:-"$(dirname "${0}")"}/logger.sh"


###############################################################################
# __first
#
# This function is the validation entry point for the script. It will trigger
# the unique environment variable name generation and check if the script is
# sourced or run directly. It will also check if the script is run as root
# or with sudo and set the shell options to ensure that the script is run
# correctly. If the script is run as root or with sudo privileges, it will
# print an error message and exit with a non-zero status code.
#
###############################################################################
__first(){
  if [ "$EUID" -eq 0 ] || [ "$UID" -eq 0 ]; then
    echo "Please do not run as root." 1>&2 > /dev/tty
    exit 1 || kill -9 $$ || true
  elif [ -n "${SUDO_USER:-}" ]; then
    echo "Please do not run as root, but with sudo privileges." 1>&2 > /dev/tty
    exit 1 || kill -9 $$ || true
  else
    local _ws_name="$(__cleanner_sourced_name)"

    if test "${BASH_SOURCE-}" != "${0}"; then
      export "${_ws_name}"="true"
    else
      export "${_ws_name}"="false"
      # This script is used to install the project binary and manage its dependencies.
      set -o errexit
      set -o nounset

      #set -euo pipefail
      set -o pipefail

      set -o errtrace
      set -o functrace
      # set -o posix

      shopt -s inherit_errexit
    fi
  fi
}
__first "$@" >/dev/tty || exit 1
## </editor-fold>


###################################################################################
# clear_script_cache
#
# This function is used to clear the script cache. It will remove the
# temporary directory and all its contents. It is not meant to be run
# directly. It is meant to be used as a cleanup function in the script.
# It will be called when the script exits or when the user interrupts
# the script. It will also check if the script is run as root or with
# sudo and will remove the temporary directory with sudo if necessary.
# 
###################################################################################
clear_script_cache() {
  # Disable the trap for cleanup
  trap - EXIT HUP INT QUIT ABRT ALRM TERM

  # Check if the temporary directory exists, if not, return
  if [ ! -d "${_TEMP_DIR}" ]; then
    exit 0
  fi

  # Remove the temporary directory
  rm -rf "${_TEMP_DIR}" || true
  # shellcheck disable=SC2046
  if test -d "${_TEMP_DIR}" && test $(sudo -v 2>/dev/null); then
    sudo rm -rf "${_TEMP_DIR}"
    if [[ -d "${_TEMP_DIR}" ]]; then
      printf '%b[_ERROR]%b ❌  %s\n' "$_ERROR" "$_NC" "Failed to remove temporary directory: ${_TEMP_DIR}"
    else
      printf '%b[_SUCCESS]%b ✅  %s\n' "$_SUCCESS" "$_NC" "Temporary directory removed: ${_TEMP_DIR}"
    fi
  fi
  exit 0
}