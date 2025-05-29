#!/usr/bin/env bash

###################################################################################
# __inst_sourced_name
#
# This function is used to generate a unique environment variable name
# based on the script name. It will be used to check if the script is sourced
# after all other operations are done and validated. This will prevent
# the script from being run directly and will ensure that the environment
# is set up correctly before running any commands. It is not meant to be run
# directly.
#
###################################################################################
__inst_sourced_name() {
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
#################################################################################
__first(){
  if [ "$EUID" -eq 0 ] || [ "$UID" -eq 0 ]; then
    echo "Please do not run as root." 1>&2 > /dev/tty
    exit 1 || kill -9 $$ || true
  elif [ -n "${SUDO_USER:-}" ]; then
    echo "Please do not run as root, but with sudo privileges." 1>&2 > /dev/tty
    exit 1 || kill -9 $$ || true
  else
    local _ws_name="$(__inst_sourced_name)"

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


# Set the field separator to handle spaces and tabs
IFS=$'\n\t'


#################################################################################
# set_trap
#
# This function sets a trap for the script to handle various signals and
# clean up resources. It will also set the shell options for error handling
# and debugging. It is not meant to be run directly. It will be called when
# the script exits or when the user interrupts the script. It will also
# check if the script is run as root or with sudo and will remove the
# temporary directory with sudo if necessary.
#
#################################################################################
# set_trap(){
#   # Get the current shell
#   get_current_shell

#   # Set the trap for the current shell and enable error handling, if applicable
#   case "${_CURRENT_SHELL}" in
#     *ksh|*zsh|*bash)

#       # Collect all arguments passed to the script into an array without modifying or running them
#       # shellcheck disable=SC2124
#       declare -a _FULL_SCRIPT_ARGS=$@

#       # Check if the script is being run in debug mode, if so, enable debug mode on the script output
#       if [[ ${_FULL_SCRIPT_ARGS[*]} =~ ^.*-d.*$ ]]; then
#           set -x
#       fi

#       # Set for the current shell error handling and some other options
#       if [[ "${_CURRENT_SHELL}" == "bash" ]]; then
#         set -o errexit
#         set -o pipefail
#         set -o errtrace
#         set -o functrace
#         shopt -s inherit_errexit
#       fi

#       # Set the trap to clear the script cache on exit.
#       # It will handle the following situations: command line exit, hangup, interrupt, quit, abort, alarm, and termination.
#       trap 'clear_script_cache' EXIT HUP INT QUIT ABRT ALRM TERM
#       ;;
#   esac

#   return 0
# }
# set_trap "$@"


#################################################################################
# what_platform
#
# This function detects the current platform and architecture. It uses the
# uname command to get the operating system and architecture information.
# It is not meant to be run directly. It will be called when the script
# is run to determine the platform and architecture for building the binary.
# It will also set the _PLATFORM and _ARCH variables to the detected values.  
# It will also set the _PLATFORM_WITH_ARCH variable to the detected platform
# and architecture in the format of "os_arch". 
# 
#################################################################################
what_platform() {
  local _platform=""
  _platform="$(uname -o 2>/dev/null || echo "")"

  local _os=""
  _os="$(uname -s)"

  local _arch=""
  _arch="$(uname -m)"

  # Detect the platform and architecture
  case "${_os}" in
  *inux|*nix)
    _os="linux"
    case "${_arch}" in
    "x86_64")
      _arch="amd64"
      ;;
    "armv6")
      _arch="armv6l"
      ;;
    "armv8" | "aarch64")
      _arch="arm64"
      ;;
    .*386.*)
      _arch="386"
      ;;
    esac
    _platform="linux-${_arch}"
    ;;
  *arwin*)
    _os="darwin"
    case "${_arch}" in
    "x86_64")
      _arch="amd64"
      ;;
    "arm64")
      _arch="arm64"
      ;;
    esac
    _platform="darwin-${_arch}"
    ;;
  MINGW|MSYS|CYGWIN|Win*)
    _os="windows"
    case "${_arch}" in
    "x86_64")
      _arch="amd64"
      ;;
    "arm64")
      _arch="arm64"
      ;;
    esac
    _platform="windows-${_arch}"
    ;;
  *)
    _os=""
    _arch=""
    _platform=""
    ;;
  esac

  if [[ -z "${_platform}" ]]; then
    kbx_log "error" "Unsupported platform: ${_os} ${_arch}"
    kbx_log "error" "Please report this issue to the project maintainers."
    return 1
  fi

  # Normalize the platform string
  _PLATFORM_WITH_ARCH="${_platform//\-/\_}"
  _PLATFORM="${_os//\ /}"
  _ARCH="${_arch//\ /}"

  return 0
}

#################################################################################
# _get_os_arr_from_args
#
# This function takes a platform argument and returns an array of supported
# platforms. If the argument is "all", it returns all supported platforms.
# If the argument is not "all", it returns the specified platform.
# It is not meant to be run directly. It will be called when the script
# is run to determine the platform and architecture for building the binary.
# 
#################################################################################
_get_os_arr_from_args() {
  local _PLATFORM_ARG=$1
  local _PLATFORM_ARR=()

  if [[ "${_PLATFORM_ARG}" == "all" ]]; then
    _PLATFORM_ARR=( "${__PLATFORMS[@]}" )
  else
    _PLATFORM_ARR=( "${_PLATFORM_ARG}" )
  fi

  for _platform_pos in "${_PLATFORM_ARR[@]}"; do
    echo "${_platform_pos} "
  done

  return 0
}

#################################################################################
# _get_arch_arr_from_args
#
# This function takes an architecture argument and returns an array of supported
# architectures. If the argument is "all", it returns all supported architectures.
# If the argument is not "all", it returns the specified architecture.
# It is not meant to be run directly. It will be called when the script
# is run to determine the platform and architecture for building the binary.
#
#################################################################################
_get_arch_arr_from_args() {
  local _ARCH_ARG=$1
  local _ARCH_ARR=()

  if [[ "${_ARCH_ARG}" == "all" ]]; then
    _ARCH_ARR=( "${__ARCHs[@]}" )
  else
    _ARCH_ARR=( "${_ARCH_ARG}" )
  fi

  echo "${_ARCH_ARR[@]}"

  return 0
}

#################################################################################
# _get_os_from_args 
#
# This function takes a platform argument and returns the corresponding
# platform name. It is not meant to be run directly. It will be called when
# the script is run to determine the platform and architecture for building
# the binary. It will also set the _PLATFORM variable to the detected value.
#
#################################################################################
_get_os_from_args() {
  local _PLATFORM_ARG=$1
  case "${_PLATFORM_ARG}" in
    all|ALL|a|A|-a|-A)
      echo "all"
      ;;
    win|WIN|windows|WINDOWS|w|W|-w|-W)
      echo "windows"
      ;;
    linux|LINUX|l|L|-l|-L)
      echo "linux"
      ;;
    darwin|DARWIN|macOS|MACOS|m|M|-m|-M)
      echo "darwin"
      ;;
    *)
      kbx_log "error" "build_and_validate: Unsupported platform: '${_PLATFORM_ARG}'."
      kbx_log "error" "Please specify a valid platform (windows, linux, darwin, all)."
      exit 1
      ;;
  esac
  return 0
}

#################################################################################
# _get_arch_from_args
#
# This function takes an architecture argument and returns the corresponding
# architecture name. It is not meant to be run directly. It will be called when
# the script is run to determine the platform and architecture for building
# the binary. It will also set the _ARCH variable to the detected value.
#
#################################################################################
_get_arch_from_args() {
  local _ARCH_ARG=$1
  case "${_ARCH_ARG}" in
    all|ALL|a|A|-a|-A)
      echo "all"
      ;;
    amd64|AMD64|x86_64|X86_64|x64|X64)
      echo "amd64"
      ;;
    arm64|ARM64|aarch64|AARCH64)
      echo "arm64"
      ;;
    386|i386|I386)
      echo "386"
      ;;
    *)
      kbx_log "error" "build_and_validate: Unsupported architecture: '${_ARCH_ARG}'. Please specify a valid architecture (amd64, arm64, 386)."
      exit 1
      ;;
  esac
  return 0
}


#################################################################################
# detect_shell_rc
#
# Detect the shell configuration file based on the current shell
# It will check for common shell configuration files like .bashrc, .zshrc, etc.
# This function will check if application binary folder is in the PATH
# and if not, it will add it to the PATH in the appropriate shell configuration file.
# It is not meant to be run directly. It will be called when the script
# 
# Arguments:
#   $1 - target path to add to PATH
#
#################################################################################
add_to_path() {
    target_path="$1"
    shell_rc_file=$(detect_shell_rc)
    if [ -z "$shell_rc_file" ]; then
        kbx_log "error" "Could not determine shell configuration file."
        return 1
    fi

    if grep -q "export PATH=.*$target_path" "$shell_rc_file" 2>/dev/null; then
        kbx_log "success" "$target_path is already in $shell_rc_file."
        return 0
    fi

    echo "export PATH=$target_path:\$PATH" >> "$shell_rc_file"
    kbx_log "success" "Added $target_path to PATH in $shell_rc_file."
    kbx_log "success" "Run 'source $shell_rc_file' to apply changes."
}

#################################################################################
# install_binary
#
# Install the binary to the appropriate directory. 
# It will check if the user is root or not and install the binary
# in the appropriate directory. It will also check if the binary is already
# installed and if so, it will remove the old binary before installing the new one.
# It will also check if the binary is in the PATH and if not, it will add it to the PATH.
# It is not meant to be run directly. 
# 
#################################################################################
install_binary() {
    local _SUFFIX="${_PLATFORM_WITH_ARCH}"
    local _BINARY_TO_INSTALL="${_BINARY}${_SUFFIX:+_${_SUFFIX}}"
    kbx_log "info" "Installing binary: '$_BINARY_TO_INSTALL' like '$_APP_NAME'"

    if [ "$(id -u)" -ne 0 ]; then
        kbx_log "info" "You are not root. Installing in $_LOCAL_BIN..."
        mkdir -p "$_LOCAL_BIN"
        cp "$_BINARY_TO_INSTALL" "$_LOCAL_BIN/$_APP_NAME" || exit 1
        add_to_path "$_LOCAL_BIN"
    else
        kbx_log "info" "Root detected. Installing in $_GLOBAL_BIN..."
        cp "$_BINARY_TO_INSTALL" "$_GLOBAL_BIN/$_APP_NAME" || exit 1
        add_to_path "$_GLOBAL_BIN"
    fi
    clean
}

#################################################################################
# install_upx
#
# This function checks if UPX is installed and installs it if not.
# It will check if the user is root or not and install UPX in the appropriate
# directory. It will also check if UPX is already installed and if so, it will
# override the old UPX before installing the new one. 
# 
#################################################################################
install_upx() {
    if ! command -v upx > /dev/null; then
        kbx_log "info" "Installing UPX..."
        if [ "$(uname)" = "Darwin" ]; then
            brew install upx
        elif command -v apt-get > /dev/null; then
            sudo apt-get install -y upx
        else
            kbx_log "error" 'Install UPX manually from https://upx.github.io/'
            exit 1
        fi
    else
        kbx_log "success" ' UPX is already installed.'
    fi
}

#################################################################################
# check_dependencies
#
# Check if the required dependencies are installed, if not, install them.
#
# Arguments:
#   $@ - list of dependencies to check
#
#################################################################################
check_dependencies() {
    # shellcheck disable=SC2317
    for dep in "$@"; do
        if ! command -v "$dep" > /dev/null; then
            kbx_log "error" "$dep is not installed."
            exit 1
        else
            kbx_log "success" "$dep is installed."
        fi
    done
}

#################################################################################
# ensure_folders
#
# Ensure that all required folders exist. This function will create the
# required folders if they do not exist. It will also check if the folders
# are writable and if not, it will print an error message and exit.
# 
#################################################################################
ensure_folders(){
    # Create the build directory if it doesn't exist
    local _build_path="$(dirname "$_BINARY")"
    if [ ! -d "$_build_path" ]; then
      kbx_log "info" "Creating build directory: $_build_path"
      mkdir -p $(dirname "$_BINARY") || return 1
    fi
    kbx_log "success" "Build directory created: $_BUILD_PATH"
}

#################################################################################
# build_binary
#
# Build the binary for the specified platform and architecture.
# 
###############################################################################
# shellcheck disable=SC2207,SC2116,SC2091,SC2155,SC2005
build_binary() {

  declare -a __platform_arr="$(echo $(_get_os_arr_from_args "$1"))"
  declare -a _platform_arr=()
  eval _platform_arr="( $(echo "${__platform_arr[@]}") )"
  kbx_log "info" "Qty OS's: ${#_platform_arr[@]}"

  declare -a __arch_arr="$(echo $(_get_arch_arr_from_args "$2"))"
  declare -a _arch_arr=()
  eval _arch_arr="( $(echo "${__arch_arr[@]}") )"
  kbx_log "info" "Qty Arch's: ${#_arch_arr[@]}"

  for _platform_pos in "${_platform_arr[@]}"; do
    if test -z "${_platform_pos}"; then
      continue
    fi
    for _arch_pos in "${_arch_arr[@]}"; do
      if test -z "${_arch_pos}"; then
        continue
      fi
      if [[ "${_platform_pos}" != "darwin" && "${_arch_pos}" == "arm64" ]]; then
        continue
      fi
      if [[ "${_platform_pos}" != "windows" && "${_arch_pos}" == "386" ]]; then
        continue
      fi
      local _OUTPUT_NAME="$(printf '%s_%s_%s' "${_BINARY}" "${_platform_pos}" "${_arch_pos}")"
      if [[ "${_platform_pos}" == "windows" ]]; then
        _OUTPUT_NAME="$(printf '%s.exe' "${_OUTPUT_NAME}")"
      fi

      local _build_env=(
        "GOOS=${_platform_pos}"
        "GOARCH=${_arch_pos}"
      )
      local _build_args=(
        "-ldflags '-s -w -X main.version=$(git describe --tags) -X main.commit=$(git rev-parse HEAD) -X main.date=$(date +%Y-%m-%d)' "
        "-trimpath -o \"${_OUTPUT_NAME}\" \"${_CMD_PATH}\""
      )

      local _build_cmd=( "${_build_env[@]}" "go build " "${_build_args[*]}" )
      local _build_cmd_str=$(echo $(printf "%s" "${_build_cmd[*]//\ / }"))
      _build_cmd_str="$(printf '%s\n' "${_build_cmd_str//\ _/_}")"
      kbx_log "info" "$(printf '%s %s/%s' "Building the binary for" "${_platform_pos}" "${_arch_pos}")"
      kbx_log "info" "Command: ${_build_cmd_str}"

      local _cmdExec=$(bash -c "${_build_cmd_str}" 2>&1 && echo "true" || echo "false")

      # Build the binary using the environment variables and arguments
      if [[ "${_cmdExec}" == "false" ]]; then
        kbx_log "error" "Failed to build the binary for ${_platform_pos} ${_arch_pos}"
        kbx_log "error" "Command: ${_build_cmd_str}"
        return 1
      else
        # If the build was successful, check if UPX is installed and compress the binary (if not Windows)
        if [[ "${_platform_pos}" != "windows" ]]; then
            install_upx
            kbx_log "info" "Packing/compressing the binary with UPX..."
            upx "${_OUTPUT_NAME}" --force-overwrite --lzma --no-progress --no-color -qqq || true
            kbx_log "success" "Binary packed/compressed successfully: ${_OUTPUT_NAME}"
        fi
        # Check if the binary was created successfully (if not Windows)
        if [[ ! -f "${_OUTPUT_NAME}" ]]; then
          kbx_log "error" "Binary not found after build: ${_OUTPUT_NAME}"
          kbx_log "error" "Command: ${_build_cmd_str}"
          return 1
        else
          local compress_vars=( "${_platform_pos}" "${_arch_pos}" )
          compress_binary "${compress_vars[@]}" || return 1
          kbx_log "success" "Binary created successfully: ${_OUTPUT_NAME}"
        fi
      fi
    done
  done

  echo ""
  kbx_log "success" "All builds completed successfully!"
  echo ""

  return 0
}

#################################################################################
# compress_binary
#
# Compress the binary into a single tar.gz/zip file. 
# It will check user system and architecture and compress the binary
# in the appropriate format. It will also check if the binary is already
# compressed and if so, it will remove the old compressed file before 
# compressing the new one.
#
#################################################################################
# shellcheck disable=SC2207,SC2116,SC2091,SC2155,SC2005
compress_binary() {
  declare -a __platform_arr="$(echo $(_get_os_arr_from_args "$1"))"
  declare -a _platform_arr=()
  eval _platform_arr="( $(echo "${__platform_arr[@]}") )"
  kbx_log "info" "Qty OS's: ${#_platform_arr[@]}"

  declare -a __arch_arr="$(echo $(_get_arch_arr_from_args "$2"))"
  declare -a _arch_arr=()
  eval _arch_arr="( $(echo "${__arch_arr[@]}") )"
  kbx_log "info" "Qty Arch's: ${#_arch_arr[@]}"

  for _platform_pos in "${_platform_arr[@]}"; do
    if [[ -z "${_platform_pos}" ]]; then
      continue
    fi
    for _arch_pos in "${_arch_arr[@]}"; do
      if [[ -z "${_arch_pos}" ]]; then
        continue
      fi
      if [[ "${_platform_pos}" != "darwin" && "${_arch_pos}" == "arm64" ]]; then
        continue
      fi
      if [[ "${_platform_pos}" == "linux" && "${_arch_pos}" == "386" ]]; then
        continue
      fi

      local _BINARY_NAME="$(printf '%s_%s_%s' "${_BINARY}" "${_platform_pos}" "${_arch_pos}")"
      if [[ "${_platform_pos}" == "windows" ]]; then
        _BINARY_NAME="$(printf '%s.exe' "${_BINARY_NAME}")"
      fi

      local _OUTPUT_NAME="${_BINARY_NAME//\.exe/}"
      local _compress_cmd_exec=""
      if [[ "${_platform_pos}" != "windows" ]]; then
        _OUTPUT_NAME="${_OUTPUT_NAME}.tar.gz"
        kbx_log "info" "Compressing the binary for ${_platform_pos} ${_arch_pos} into ${_OUTPUT_NAME}..."
        _compress_cmd_exec=$(tar -czf "${_OUTPUT_NAME}" "${_BINARY_NAME}" 2>&1 && echo "true" || echo "false")
      else
        _OUTPUT_NAME="${_OUTPUT_NAME}.zip"
        kbx_log "info" "Compressing the binary for ${_platform_pos} ${_arch_pos} into ${_OUTPUT_NAME}..."
        _compress_cmd_exec=$(zip -r -9 "${_OUTPUT_NAME}" "${_BINARY_NAME}" 2>&1 && echo "true" || echo "false")
      fi
      if [[ "${_compress_cmd_exec}" == "false" ]]; then
        kbx_log "error" "Failed to compress the binary for ${_platform_pos} ${_arch_pos}"
        kbx_log "error" "Command: ${_compress_cmd_exec}"
        return 1
      else
        kbx_log "success" "Binary compressed successfully: ${_OUTPUT_NAME}"
      fi
    done
  done

  kbx_log "success" "All binaries compressed successfully!"

  return 0
}

#################################################################################
# validate_versions
#
# Validate the Go version wich is required to build the binary and check if
# the Go modules are tidy. It will check if the Go version is same or greater
# than the go.mod required version. 
# 
#################################################################################
validate_versions() {
    REQUIRED_GO_VERSION="${_VERSION_GO:-1.20.0}"
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$(printf '%s\n' "$REQUIRED_GO_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_GO_VERSION" ]]; then
        kbx_log "error" "Go version must be >= $REQUIRED_GO_VERSION. Detected: $GO_VERSION"
        exit 1
    fi
    kbx_log "success" "Go version is valid: $GO_VERSION"
    go mod tidy || return 1
}

#################################################################################
# sumary
#
# Print a summary of the installation process.
# 
#################################################################################
summary() {
    install_dir="$_BINARY"
    kbx_log "success" "Build and installation complete!"
    kbx_log "success" "Binary: $_BINARY"
    kbx_log "success" "Installed in: $install_dir"
    check_path "$install_dir"
}

#################################################################################
# build_and_validate
#
# Is the function tha precede the build process. It will check if the Go version
# is valid and if the Go modules are tidy. It will also check if the user
# provided a platform and architecture. If not, it will use the default values.
# 
##################################################################################
build_and_validate() {
    # Check if the Go version is valid
    validate_versions

    local _PLATFORM_ARG="$1"
    # _PLATFORM_ARG="$(_get_os_from_args "${1:-${_PLATFORM}}")"
    local _ARCH_ARG="$2"
    # _ARCH_ARG="$(_get_arch_from_args "${2:-${_ARCH}}")"

    kbx_log "info" "Building for platform: ${_PLATFORM_ARG}, architecture: ${_ARCH_ARG}" true
    local _WHICH_COMPILE_ARG=( "${_PLATFORM_ARG}" "${_ARCH_ARG}" )

    # Call the build function with the platform and architecture arguments
    build_binary "${_WHICH_COMPILE_ARG[@]}" || exit 1

    return 0
}

#################################################################################
# check_path
#
# Check if the installation directory is in the PATH
# 
# Arguments:
#   $1 - installation directory
# 
#################################################################################
check_path() {
    kbx_log "info" "Checking if the installation directory is in the PATH..."
    if ! echo "$PATH" | grep -q "$1"; then
        kbx_log "warn" "$1 is not in the PATH."
        kbx_log "warn" "Add the following to your ~/.bashrc, ~/.zshrc, or equivalent file:"
        kbx_log "warn" "export PATH=$1:\$PATH"
    else
        kbx_log "success" "$1 is already in the PATH."
    fi
}

#################################################################################
# download_binary
#
# Download the binary from the release URL. 
#
#################################################################################
download_binary() {
    # Obtem o sistema operacional e a arquitetura
    if ! what_platform > /dev/null; then
        kbx_log "error" "Failed to detect platform."
        return 1
    fi

    # Validação: Verificar se o sistema operacional ou a arquitetura são suportados
    if [[ -z "${_PLATFORM}" ]]; then
        kbx_log "error" "Unsupported platform: ${_PLATFORM}"
        return 1
    fi

    # Obter a versão mais recente de forma robusta (fallback para "latest")
    version=$(curl -s "https://api.github.com/repos/${_OWNER}/${_PROJECT_NAME}/releases/latest" | \
        grep "tag_name" | cut -d '"' -f 4 || echo "latest")

    if [ -z "$version" ]; then
        kbx_log "error" "Failed to determine the latest version."
        return 1
    fi

    # Construir a URL de download usando a função customizável
    release_url=$(get_release_url)
    kbx_log "info" "Downloading ${_APP_NAME} binary for OS=$os, ARCH=$arch, Version=$version..."
    kbx_log "info" "Release URL: ${release_url}"

    archive_path="${_TEMP_DIR}/${_APP_NAME}.tar.gz"

    # Realizar o download e validar sucesso
    if ! curl -L -o "${archive_path}" "${release_url}"; then
        kbx_log "error" "Failed to download the binary from: ${release_url}"
        return 1
    fi
    kbx_log "success" "Binary downloaded successfully."

    # Extração do arquivo para o diretório binário
    kbx_log "info" "Extracting binary to: $(dirname "${_BINARY}")"
    if ! tar -xzf "${archive_path}" -C "$(dirname "${_BINARY}")"; then
        kbx_log "error" "Failed to extract the binary from: ${archive_path}"
        rm -rf "${_TEMP_DIR}"
        exit 1
    fi

    # Limpar artefatos temporários
    rm -rf "${_TEMP_DIR}"
    kbx_log "success" "Binary extracted successfully."

    # Verificar se o binário foi extraído com sucesso
    if [ ! -f "$_BINARY" ]; then
        kbx_log "error" "Binary not found after extraction: $_BINARY"
        exit 1
    fi

    kbx_log "success" "Download and extraction of ${_APP_NAME} completed!"
}

#################################################################################
# install_from_release
#
# Install the downloaded binary from the release URL.
# It will check if the user is root or not and install the binary on the
# appropriate directory. It will also check if the binary is already
# installed and if so, it will remove the old binary before installing the new one.
# It will also check if the binary is in the PATH and if not, it will add it to the PATH.
# It is not meant to be run directly.
#
#################################################################################
install_from_release() {
    download_binary
    install_binary
}

#################################################################################
# show_about
#
# Print the ABOUT message
#
#################################################################################
show_about() {
    # Print the ABOUT message
    printf '%s\n\n' "${_ABOUT:-}"
}

#################################################################################
# show_banner
#
# Print the BANNER message
#
#################################################################################
show_banner() {
    # Print the ABOUT message
    printf '\n%s\n\n' "${_BANNER:-}"
}

#################################################################################
# show_headers
#
# Print the BANNER and ABOUT messages
# 
#################################################################################
show_headers() {
    # Print the BANNER message
    show_banner || return 1
    # Print the ABOUT message
    show_about || return 1
}

#################################################################################
# main
#
# Main function that handles the command line arguments and calls the
# appropriate functions based on the command provided. It will also
# check if the user has provided a command and if not, it will print
# an error message and exit. It will also check if the user has provided
# a platform and architecture and if not, it will use the default values.
#
#################################################################################
# shellcheck disable=SC2155
main() {
  # Detect the platform if not provided, will be used in the build command
  what_platform || exit 1

  # Show the banner information
  if [[ "${_DEBUG:-}" != true ]]; then
    show_headers
  else
    kbx_log "info" "Debug mode enabled. Skipping banner..."
    if [[ -z "${_HIDE_ABOUT:-}" ]]; then
      show_about
    fi
  fi

  _ARGS=( "$@" )
  local _default_label='Auto detect'
  local _arrArgs=( "${_ARGS[@]:0:$#}" )
  local _PLATFORM_ARG=$(_get_os_from_args "${_arrArgs[1]:-${_PLATFORM}}")
  local _ARCH_ARG=$(_get_arch_from_args "${_arrArgs[2]:-${_ARCH}}")

  # Check if the user has provided a command
  kbx_log "info" "Command: ${_arrArgs[0]:-}" true
  kbx_log "info" "Platform: ${_PLATFORM_ARG:-$_default_label}" true
  kbx_log "info" "Architecture: ${_ARCH_ARG:-$_default_label}" true

  case "${_arrArgs[0]:-}" in
    build|BUILD|-b|-B)
      # Call the build function with the detected platform
      build_and_validate "$_PLATFORM_ARG" "$_ARCH_ARG" || exit 1
      ;;
    install|INSTALL|-i|-I)
      kbx_log "info" "Executing install command..."
      read -r -p "Do you want to download the precompiled binary? [y/N] (No will build locally): " c </dev/tty
      kbx_log "info" "User choice: ${c}"

      if [ "$c" = "y" ] || [ "$c" = "Y" ]; then
          kbx_log "info" "Downloading precompiled binary..." true
          install_from_release "$_PLATFORM_ARG" "$_ARCH_ARG" || exit 1
      else
          kbx_log "info" "Building locally..." true
          build_and_validate "$_PLATFORM_ARG" "$_ARCH_ARG" || exit 1
          install_binary "$_PLATFORM_ARG" "$_ARCH_ARG" || exit 1
      fi

      summary
      ;;
    clear|clean|CLEAN|-c|-C)
      kbx_log "info" "Executing clean command..."
      clean || exit 1
      kbx_log "success" "Clean command executed successfully."
      ;;
    *)
      kbx_log "error" "Invalid command: $1"
      echo "Usage: $0 {build|install|clean}"
      ;;
  esac
}
# Execute the main function with all script arguments
## echo "MAKE ARGS: ${ARGS[*]:-}"
main "$@"
