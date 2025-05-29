#!/usr/bin/env bash
# shellcheck disable=SC2155,SC2163,SC2207,SC2116

###################################################################################
# __build_sourced_name
#
# This function is used to generate a unique environment variable name
# based on the script name. It will be used to check if the script is sourced
# after all other operations are done and validated. This will prevent
# the script from being run directly and will ensure that the environment
# is set up correctly before running any commands. It is not meant to be run
# directly.
#
###################################################################################
__build_sourced_name() {
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
# to a file. It is not meant to be run directly. This method is used to
# ensure that the script is loaded only if it is not already loaded and throws
# an error if the script is not found automatically.
# 
#################################################################################
# shellcheck disable=SC2065,SC1091
test -z "$(declare -f kbx_log)" >/dev/null && source "${_kbx_path_helpers:-"$(dirname "${0}")"}/logger.sh"

# Here you can load other scripts that are needed for the script to work
# test -z "$(declare -f kbx_log)" >/dev/null && source "$HOME/.${SHELL:-bash}rc

#############################################################################
# __build_list_functions
#
# This function is used to export only functions without '__' prefix.
# It prevents the script from exporting functions that are not meant to be used
# outside the script. It is used to clean up the environment before running
# any commands. It is not meant to be run directly.
# 
############################################################################
# shellcheck disable=SC2155
__build_list_functions() {
  local _str_functions=$(declare -F | awk '{print $3}' | grep -v "^__") >/dev/null || return 1
  declare -a _functions=( $(echo "$_str_functions") ) > /dev/null || return 1
  echo "${_functions[@]}"
  return 0
}
__build_main_functions() {
  local _exported_functions=( $(__build_list_functions) ) >/dev/null || return 1
  for _exported_function in "${_exported_functions[@]}"; do
    export -f "${_exported_function}" >/dev/null || return 1
  done
  return 0
}

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
#################################################################################
__first(){
  if [ "$EUID" -eq 0 ] || [ "$UID" -eq 0 ]; then
    echo "Please do not run as root." 1>&2 > /dev/tty
    exit 1 || kill -9 $$ || true
  elif [ -n "${SUDO_USER:-}" ]; then
    echo "Please do not run as root, but with sudo privileges." 1>&2 > /dev/tty
    exit 1 || kill -9 $$ || true
  else
    local _ws_name="$(__build_sourced_name)"

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

################################################################################
### HERE BEGINS THE SCRIPT LOGIC IN A SECURE ENVIRONMENT:
### WITHOUT ROOT, WITHOUT SUDO, WITHOUT EXPORTING WHAT IS NOT NEEDED (JUST PUT __ IN FRONT OF THE FUNCTION NAME)

# Some functions here.. 

# Some functions here..

# Some functions here..

### HERE ENDS THE WRAPPED SCRIPT LOGIC, HERE BELOW THE SCRIPT WILL ENSURE THE ISOLATION ABOVE
###############################################################################
# __main
#
# This function is the main entry point for the script. It will call the
# function passed as argument and pass the arguments to it, but only after 
# checking all validations and setting up the environment correctly.
# 
###############################################################################
__build_main() {
  # Restore the dynamic variable name based on the script path/name from a variable 
  # that we set the name dynamic and we need to restore it to its original name,
  # so we can use it in the script programmatically. Because of the way
  # bash works we have oportunty, for example:
  #     - check if the script is sourced or run directly and define what to do
  #     - prevent the script from exporting functions that are not meant to be used outside the script
  #     - prevent the script from being run directly and ensure that the environment is set up correctly
  #     - prevent the script from being run as root or with sudo privileges
  # this also helps to:
  #     - executing the script by mistake
  #     - sourcing the script by mistake
  #     - functions and methods with equal names in different scripts for being triggered by mistake/error
  #     - scripts with equal names in different directories for being triggered by mistake/error
  local _ws_name="$(__build_sourced_name)"
  eval "local _ws_name=\$${_ws_name}" >/dev/null

  # This echo is used like a workaround some bash versions that do not support
  # the "test" command like we expect, not retrieve the value of the variable like we expect, etc..
  if test "$(echo "${_ws_name}")" != "true"; then
    __build_main_functions "$@"
    exit $?
  else
    kbx_die 33 "This script is not meant to be sourced" || echo "This script is not meant to be sourced" > /dev/tty
    exit 3
  fi
}
__build_main "$@"