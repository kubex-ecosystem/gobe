#!/usr/bin/env bash
# shellcheck disable=SC2155,SC2163

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
__env_sourced_name() {
  local _self="${BASH_SOURCE-}"
  _self="${_self//${_kbx_root:-$(dirname "$(dirname "$(dirname "$(readlink -e "$0")")")")}/}"
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
    local _ws_name="$(__env_sourced_name)"

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
__env_list_functions() {
  local _str_functions=$(declare -F | awk '{print $3}' | grep -v "^_") >/dev/null || return 1
  declare -a _functions=( $(echo "$_str_functions") ) > /dev/null || return 1
  echo "${_functions[@]}"
  return 0
}
__env_main_functions() {
  # shellcheck disable=SC2207
  local _exported_functions=( $(__env_list_functions) ) >/dev/null || return 1
  for _exported_function in "${_exported_functions[@]}"; do
    export -f "${_exported_function}" >/dev/null || return 61
  done
  return 0
}
## </editor-fold>


############################################################################
# get_globals
#
# get_globals is a function to get the global variables for the project.
# It is called by the main script to ensure that the environment is set up
# correctly before running any commands. It is not meant to be run directly.
# It transfers the global variables to the caller script without exporting them
# to the environment. This is useful for debugging, testing purposes and security.
#
##############################################################################
get_globals() {


    local _BANNER='################################################################################

    ██   ██ ██     ██ ██████   ████████ ██     ██
    ░██  ██ ░██    ░██░█░░░░██ ░██░░░░░ ░░██   ██
    ░██ ██  ░██    ░██░█   ░██ ░██       ░░██ ██
    ░████   ░██    ░██░██████  ░███████   ░░███
    ░██░██  ░██    ░██░█░░░░ ██░██░░░░     ██░██
    ░██░░██ ░██    ░██░█    ░██░██        ██ ░░██
    ░██ ░░██░░███████ ░███████ ░████████ ██   ░░██
    ░░   ░░  ░░░░░░░  ░░░░░░░  ░░░░░░░░ ░░     ░░'

    local _DEBUG=${DEBUG:-false}
    local _HIDE_ABOUT=${HIDE_ABOUT:-false}
    # This variable are used to customize the script behavior, like repository url and owner
    local _OWNER="faelmori"
    # The _REPO_ROOT variable is set to the root directory of the repository. One above the script directory.
    local _REPO_ROOT="${ROOT_DIR:-$(dirname "$(dirname "$(dirname "$(realpath "$0")")")")}"
    # The _APP_NAME variable is set to the name of the repository. It is used to identify the application.
    local _APP_NAME="${APP_NAME:-$(basename "$_REPO_ROOT")}"
    # The _PROJECT_NAME variable is set to the name of the project. It is used for display purposes.
    local _PROJECT_NAME="$_APP_NAME"
    # The _VERSION variable is set to the version of the project. It is used for display purposes.
    local _VERSION=$(cat "$_REPO_ROOT/version/CLI_VERSION" 2>/dev/null || echo "v0.0.0")
    # The _VERSION_GO variable is set to the version of the Go required by the project.
    local _VERSION_GO=$(grep '^go ' go.mod | awk '{print $2}')
    # The _VERSION variable is set to the version of the project. It is used for display purposes.
    local _LICENSE="MIT"
    # The _ABOUT variable contains information about the script and its usage.
    local _ABOUT="################################################################################
This Script is used to install ${_PROJECT_NAME} project, version ${_VERSION}.
Supported OS: Linux, MacOS, Windows
Supported Architecture: amd64, arm64, 386
Source: https://github.com/${_OWNER}/${_PROJECT_NAME}
Binary Release: https://github.com/${_OWNER}/${_PROJECT_NAME}/releases/latest
License: ${_LICENSE}
Notes:
    - [version] is optional; if omitted, the latest version will be used.
    - If the script is run locally, it will try to resolve the version from the
    repo tags if no version is provided.
    - The script will install the binary in the ~/.local/bin directory if the
    user is not root. Otherwise, it will install in /usr/local/bin.
    - The script will add the installation directory to the PATH in the shell
    configuration file.
    - The script will also install UPX if it is not already installed.
    - The script will build the binary if the build option is provided.
    - The script will download the binary from the release URL
    - The script will clean up build artifacts if the clean option is provided.
    - The script will check if the required dependencies are installed.
    - The script will validate the Go version before building the binary.
    - The script will check if the installation directory is in the PATH.
################################################################################"
    # Variable to store the current running shell
    local _CURRENT_SHELL=""
    # The _CMD_PATH variable is set to the path of the cmd directory. It is used to
    # identify the location of the main application code.
    local _CMD_PATH="${_REPO_ROOT}/cmd"
    # The _BUILD_PATH variable is set to the path of the build directory. It is used
    # to identify the location of the build artifacts.
    local _BUILD_PATH="$(dirname "${_CMD_PATH}")/build"
    # The _BINARY variable is set to the path of the binary file. It is used to
    # identify the location of the binary file.
    local _BINARY="${_BUILD_PATH}/${_APP_NAME}"
    # The _LOCAL_BIN variable is set to the path of the local bin directory. It is
    # used to identify the location of the local bin directory.
    local _LOCAL_BIN="${HOME:-"~"}/.local/bin"
    # The _GLOBAL_BIN variable is set to the path of the global bin directory. It is
    # used to identify the location of the global bin directory.
    local _GLOBAL_BIN="/usr/local/bin"
    # For internal use only
    local __PLATFORMS=( "windows" "darwin" "linux" )
    local __ARCHs=( "amd64" "386" "arm64" )
    # The _PLATFORM variable is set to the platform name. It is used to identify the
    # platform on which the script is running.
    local _PLATFORM_WITH_ARCH=""
    local _PLATFORM=""
    local _ARCH=""
    local __envvars_loaded="1"

    declare -A _envs=(
        ['_DEBUG']="${_DEBUG}"
        ['_HIDE_ABOUT']="${_HIDE_ABOUT}"
        ['_OWNER']="${_OWNER}"
        ['_REPO_ROOT']="${_REPO_ROOT}"
        ['_APP_NAME']="${_APP_NAME}"
        ['_PROJECT_NAME']="${_PROJECT_NAME}"
        ['_VERSION']="${_VERSION}"
        ['_VERSION_GO']="${_VERSION_GO}"
        ['_LICENSE']="${_LICENSE}"
        ['_ABOUT']="${_ABOUT}"
        ['_BANNER']="${_BANNER}"
        ['_CURRENT_SHELL']="${_CURRENT_SHELL}"
        ['_CMD_PATH']="${_CMD_PATH}"
        ['_BUILD_PATH']="${_BUILD_PATH}"
        ['_BINARY']="${_BINARY}"
        ['_LOCAL_BIN']="${_LOCAL_BIN}"
        ['_GLOBAL_BIN']="${_GLOBAL_BIN}"
        ['__PLATFORMS']="${__PLATFORMS}"
        ['__ARCHs']="${__ARCHs}"
        ['_PLATFORM_WITH_ARCH']="${_PLATFORM_WITH_ARCH}"
        ['_PLATFORM']="${_PLATFORM}"
        ['_ARCH']="${_ARCH}"
        ['__envvars_loaded']="${__envvars_loaded}"
    )

    # Export the variables to make them available in the environment
    for _key in "${!_envs[@]}"; do
        echo "${_key}"="${_envs[${_key}]}"
    done
    # Set the _PLATFORM variable based on the current platform
}

################################################################################
# __set_globals
# 
# This function is used to set the global variables for the project.
# It is called by the main script to ensure that the environment is set up
# correctly before running any commands. It is not meant to be run directly.
#
##################################################################################
__set_globals() {
    # Set the _PLATFORM variable based on the current platform
    local _globals=()
    _
    # Export the global variables to make them available in the environment
    for _key in "${!_globals[@]}"; do
        export "${_key}"="${_globals[${_key}]}"
    done
}

###############################################################################
# __unset_globals
#
# This function is used to unset exported global variables.
# It isn't exported to the environment, is used internally just if needed.
# 
#################################################################################
__unset_globals() {
    # Unset the global variables to clean up the environment
    for _key in "${!_globals[@]}"; do
        unset "${_key}"
    done
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
  local _ws_name="$(__env_sourced_name)"
  eval "local _ws_name=\$${_ws_name}" >/dev/null
  if [ $(echo "$_ws_name") != "true" ]; then
    __env_main_functions "$@"
    exit $?
  else
    kbx_die 33 "This script is not meant to be sourced" || echo "This script is not meant to be sourced" > /dev/tty
    exit 3
  fi
}
_main "$@"
