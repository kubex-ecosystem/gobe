#!/usr/bin/env bash


###############################################################################
# __wrapper
#
# This is a wrapper script to run the kubex modules installation,
# building, uninstalling, cleaning and testing. It is not meant to be run
# directly to avoid any issues and protect the environment and the user.
# It is meant to be run by the Makefile or other scripts. 
#
###############################################################################
# shellcheck disable=SC2317,SC2155
__wrapper(){
  # This a secure script to run kubex modules installation, building, uninstalling, cleaning and testing.
  # It is not meant to be run directly to avoid any issues and protect the environment and the user.
  # It is meant to be run by the Makefile or other scripts.
  # __check(){
  #   return 0
  #   # # Check if folders and files exist
  #   # local _current_hash=$(find "${_kbx_path_scripts}" -type f -exec md5sum {} + | md5sum | awk '{ print $1 }')
  #   # local _last_hash=$(cat "${_kbx_path_build}/logs/.hash")
  #   # local _git_hash=$(_ghh=$(git rev-parse HEAD) && curl -s "https://raw.githubusercontent.com/kubex-io/kubex/${_ghh}/support/scripts/.hash" 2>/dev/null || echo "0")
  #   # if test "${_current_hash}" != "${_last_hash}"; then
  #   #   echo "KUBEX: Please, run the script: ${_kbx_path_build}/logs/.hash" > /dev/tty
  #   #   exit 1 || kill -9 $$
  #   # fi

  #   # local _curl_hash_check=$(cat "${_kbx_path_helpers}/curl_hash.txt" 2>/dev/null || echo "0")
  #   # if test "${_current_hash}" != "${_curl_hash_check}"; then
  #   #   echo "KUBEX: Please, run the script: ${_kbx_path_helpers}/curl_hash.sh" > /dev/tty
  #   #   exit 1 || kill -9 $$
  #   # fi

  #   # if test -d "${_kbx_path_source}" && test -d "${_kbx_path_scripts}" && test -d "${_kbx_path_helpers}"; then
  #   #   if test -f "${_kbx_path_source}/main.go" && test -f "${_kbx_path_scripts}/main.sh" && test -f "${_kbx_path_helpers}/logger.sh"; then
  #   #     return 0
  #   #   else
  #   #     return 1
  #   #   fi
  #   # else
  #   #   return 1
  #   # fi
  # }

  _kbx_path_run="$(readlink -e "${0}")" || return 1
  _kbx_path_root="$(realpath "$(dirname "${_kbx_path_run}")/../..")" || return 1
  _kbx_path_build="${_kbx_path_root}/build" || return 1
  _kbx_path_source="${_kbx_path_root}/cmd" || return 1
  _kbx_path_scripts="${_kbx_path_root}/support/scripts" || return 1
  _kbx_path_helpers="${_kbx_path_root}/support/utils" || return 1
  _kbx_path_settings="${_kbx_path_root}/support/settings" || return 1
  
  if test $EUID && test $UID -eq 0; then
    kbx_log error "KUBEX: Please do not run as root."
    exit 1 || kill -9 $$
  elif test -n "${SUDO_USER:-}"; then
    kbx_log error "KUBEX: Please do not run as root, but with sudo privileges." > /dev/tty
    exit 1 || kill -9 $$
  else
    # shellcheck disable=SC2065,SC1091
    test -z "$(declare -f kbx_log)" >/dev/null && source "${_kbx_path_helpers:-"$(dirname "${0}")"}/logger.sh"

    # if ! __check; then
    #   kbx_log error "KUBEX: Please, this script is not meant to be sourced!" > /dev/tty
    #   exit 1 || kill -9 $$
    # else
      set -o errexit
      set -o nounset
      set -o pipefail
      set -o errtrace
      set -o functrace
      shopt -s inherit_errexit

      readonly _kbx_path_root="${_kbx_path_root}" && export _kbx_path_root || return 1
      readonly _kbx_path_build="${_kbx_path_build}" && export _kbx_path_build || return 1
      readonly _kbx_path_source="${_kbx_path_source}" && export _kbx_path_source || return 1
      readonly _kbx_path_scripts="${_kbx_path_scripts}" && export _kbx_path_scripts || return 1
      readonly _kbx_path_helpers="${_kbx_path_helpers}" && export _kbx_path_helpers || return 1
      readonly _kbx_path_settings="${_kbx_path_settings}" && export _kbx_path_settings || return 1

      local _kbx_args=( "$@" )
      local _kbx_args_len="${#_kbx_args[@]}"
      local _kbx_run_file=""
      if test "${_kbx_args_len}" -gt 0; then
        _kbx_run_file="${_kbx_args[0]}"
        unset "_kbx_args[0]"
      else
        _kbx_run_file="help"
      fi
      _kbx_run_file="${_kbx_path_scripts}/${_kbx_run_file}.sh"

      # Check if the file exists and if it is inside the scripts folder.
      # after that, 
      if test ! -f "${_kbx_run_file}"; then
        kbx_log error "KUBEX: File not found: ${_kbx_run_file}" > /dev/tty
        exit 1 || kill -9 $$
      else 
        kbx_log debug "KUBEX: File found: ${_kbx_run_file}" > /dev/tty
      fi
      
      # shellcheck disable=SC1091
      . "${_kbx_path_settings}/loader.sh" || return 1

      declare -a _run_cmd=( "${_kbx_run_file}" "$@" ) || return 1
      kbx_run info "${_run_cmd[@]}" || return 1
    fi
  #fi
}


export -f __wrapper
