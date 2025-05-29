#!/usr/bin/env bash
# shellcheck disable=SC2155,SC2163



# Função para verificar modificações
__check_modifications() {
    if [[ ! -d "${PWD}/logs" ]]; then
        mkdir -p "${PWD}/logs" || ("$(command -v pretty_log || echo "echo")" "Failed to create the logs directory" "error" && return 1)
    fi

    local _dir=$1
    local _build_file_name=$(basename "${_dir%/*}")
    _build_file_name="${_build_file_name}_$(basename "$_dir")"

    local _hash_file="${PWD}/logs/${_build_file_name}.hash"
    if [[ ! -f $_hash_file ]]; then
        touch "${PWD}/logs/${_build_file_name}.hash" || ("$(command -v pretty_log || echo "echo")" "Failed to create the hash file" "error" && return 1)
    fi

    local _current_hash=$(find "${_dir}" -type f -exec md5sum {} + | md5sum | awk '{ print $1 }')

    if [[ -f "${_hash_file}" ]]; then
        local _last_hash=$(cat "$_hash_file")
        if [[ $_current_hash == "$_last_hash" ]]; then
            echo "Old build hash: $_last_hash"
            echo "New build hash: $_current_hash"
            return 1  # Sem modificações
        else
            echo "$_current_hash" > "$_hash_file"
            echo "Old build hash: $_last_hash"
            echo "New build hash: $_current_hash"
            return 0 # Com modificações
        fi
    else
        echo "$_current_hash" > "$_hash_file"
        echo "New build hash: $_current_hash"
        return 0  # Com modificações
    fi
}

###############################################################################
# __env_sourced_name
#
# This function is used to generate a unique environment variable name
# based on the script name. It will be used to check if the script is sourced
# after all other operations are done and validated. This will prevent
# the script from being run directly and will ensure that the environment
# is set up correctly before running any commands. It is not meant to be run
# directly.
# 
###############################################################################
__loader_sourced_name() {
  local _self="${BASH_SOURCE-}"
  _self="${_self//${_kbx_root:-$(git rev-parse --show-toplevel)}/}"
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
    local _ws_name="$(__loader_sourced_name)"

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


#############################################################################
# clear_script_cache
#
# This function is used to export only functions without '__' prefix.
# It prevents the script from exporting functions that are not meant to be used
# outside the script. It is used to clean up the environment before running
# any commands. It is not meant to be run directly.
# 
############################################################################
# shellcheck disable=SC2207,SC2116,SC2155
__loader_list_functions() {
  local _str_functions=$(declare -F | awk '{print $3}' | grep -v "^__") >/dev/null || return 1
  declare -a _functions=( $(echo "$_str_functions") ) > /dev/null || return 1
  echo "${_functions[@]}"
  return 0
}
__loader_main_functions() {
  # shellcheck disable=SC2207
  local _exported_functions=( $(__loader_list_functions) ) >/dev/null || return 1
  for _exported_function in "${_exported_functions[@]}"; do
    export -f "${_exported_function}" >/dev/null || return 61
  done
  return 0
}


#################################################################################
# get_current_shell
# 
# This function retrieves the current shell being used. It checks the process
# name and the shebang line of the script to determine the shell. It is not
# meant to be run directly.
# 
#################################################################################
get_current_shell() {
  _CURRENT_SHELL="$(cat /proc/$$/comm)"

  case "${0##*/}" in
    ${_CURRENT_SHELL}*)
      shebang="$(head -1 "${0}")"
      _CURRENT_SHELL="${shebang##*/}"
      ;;
  esac

  return 0
}

#########################################################################
# get_release_url
#
# This function is used to get the release URL for the binary.
# It can be customized to change the URL format or add additional parameters.
# Actually im using the default logic to construct the URL with the release version, the platform and the architecture
# with the format .tar.gz or .zip (for windows). Sweet yourself.
#
#############################################################################
get_release_url() {
    # Default logic for constructing the release URL
    local _os="${_PLATFORM%%-*}"
    local _arch="${_PLATFORM##*-}"
    # If os is windows, set the format to .zip, otherwise .tar.gz
    local _format="${_os:zip=tar.gz}"

    local _url=$(printf "https://github.com/%s/%s/releases/download/%s/%s_.%s" "${_OWNER}" "${_PROJECT_NAME}" "${_VERSION}" "${_PROJECT_NAME}" "${_format}")

    echo "${_url}"

    return 0
}


###############################################################################
# detect_shell_rc
#
# This function is used to detect the shell configuration file for the current
# user. It will return the path to the shell configuration file based on the
# current shell. It is used to add the installation directory to the PATH
# in the shell configuration file. It is not meant to be run directly.
# It is used to ensure that the installation directory is added to the PATH
# 
#############################################################################
detect_shell_rc() {
    shell_rc_file=""
    user_shell=$(basename "$SHELL")
    case "$user_shell" in
        bash) shell_rc_file="$HOME/.bashrc" ;;
        zsh) shell_rc_file="$HOME/.zshrc" ;;
        sh) shell_rc_file="$HOME/.profile" ;;
        fish) shell_rc_file="$HOME/.config/fish/config.fish" ;;
        *)
            kbx_die "warn" "Unsupported shell, modify PATH manually."
            return 1
            ;;
    esac
    kbx_die "info" "$shell_rc_file"
    if [ ! -f "$shell_rc_file" ]; then
        kbx_die "error" "Shell configuration file not found: $shell_rc_file"
        return 1
    fi
    echo "$shell_rc_file"
    return 0
}


###############################################################################
# __main
#
# This function is the main entry point for the script. It will call the
# function passed as argument and pass the arguments to it, but only after 
# checking all validations and setting up the environment correctly.
# 
###############################################################################
_main() {
  # if ! __check_modifications "${_kbx_path_source:-$(dirname "${0}")}"; then
  #   echo "No modifications detected, skipping build." > /dev/tty
  #   return 0
  # fi
  local _ws_name="$(__loader_sourced_name)"
  eval "local _ws_name=\$${_ws_name}" >/dev/null
  if [ $(echo "$_ws_name") != "true" ]; then
      kbx_die "error" "This script is not meant to be run directly. Please source it instead."
      exit 1 || kill -9 $$
  else
    __loader_main_functions "$@"
  fi
}

_main "$@"
