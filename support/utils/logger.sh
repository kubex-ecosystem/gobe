#!/usr/bin/env bash
# shellcheck disable=SC2155,SC2163,SC2207


# shellcheck disable=SC2065,SC1091
test -z "$(declare -f log)" >/dev/null && source "${_kbx_path_helpers:-"$(dirname "${0}")"}/colors.sh"


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
__log_list_functions() {
  local _str_functions=$(declare -F | awk '{print $3}' | grep -v "^__") >/dev/null || return 61
  declare -a _functions=( $(echo "$_str_functions") ) > /dev/null || return 61
  echo "${_functions[@]}"
  return 0
}
__log_main_functions() {
  local _exported_functions=( $(__log_list_functions) ) >/dev/null || return 61
  for _exported_function in "${_exported_functions[@]}"; do
    export -f "${_exported_function}" || return 61
  done
  return 0
}

##########################################################################
# kbx_create_temp_dir
#
# Creates a temporary directory for logs.
# The function checks if the _TEMP_DIR variable is set and if it points to a
# valid directory. If not, it creates a new temporary directory and sets
# the _TEMP_DIR variable to point to it. The function also creates a kbx_log file
# in the temporary directory and sets the _kbx_log_output variable to point
# to the kbx_log file. The function exports the _TEMP_DIR and _kbx_log_output
# variables for use in other functions.
# The function returns 0 on success and 1 on failure.
# 
############################################################################
_kbx_create_temp_dir() {
  # Create a temporary directory for logs
  local _temp_check="$(test "$(test -z "${_TEMP_DIR:-}" || ! test -d "${_TEMP_DIR:-}")" && echo "true" || echo "false")" || return 1
  if [[ ! -d "${_TEMP_DIR:-}" ]]; then
    _TEMP_DIR="$(mktemp -d)" || return 1
    export _TEMP_DIR || return 1
    touch "${_TEMP_DIR}/log.txt" || return 1
    chmod 777 "${_TEMP_DIR}/log.txt" || return 1
  fi
  if test -f "${_TEMP_DIR}/log.txt"; then
    _kbx_log_output="${_TEMP_DIR}/log.txt" || return 1
  else
    _kbx_log_output="$(mktemp -p "${_TEMP_DIR}" log.txt)" || return 1
  fi
  export _kbx_log_output || return 1
}


##########################################################################
# kbx_tail_log
#
# Tails the kbx_log file and waits for it to finish.
# The function runs the tail command in the background and waits for it
# to finish. The function also sets the kbx_tail_pid variable to the 
# process ID of the tail command. The function exports the kbx_tail_pid
# variable for use in other functions. The function returns 0 on success
# and 1 on failure.
#
############################################################################
kbx_tail_log(){
  local tail_pid
  tail -f "${_kbx_log_output}" &
  tail_pid=$!
  export kbx_tail_pid="${tail_pid}"
  wait "${tail_pid}"
}

##########################################################################
# kbx_set_trap
#
# Sets up traps for various signals and errors.
# The function sets up traps for various signals (EXIT, HUP, INT, QUIT,
# ABRT, ALRM, TERM) and errors. The function also sets up a trap for
# the kbx_exit_handler function to handle errors and cleanup. The
# function exports the kbx_last_exit variable to store the last exit
# code. The function also sets up a trap for the kbx_end function to
# handle cleanup on exit. The function returns 0 on success and 1 on
# failure.
#
############################################################################
kbx_set_trap(){
  _curr_shell="$(cat /proc/$$/comm)"
  case "${0##*/}" in
    ${_curr_shell}*)
      shebang="$(head -1 "${0}")"
      _curr_shell="${shebang##*/}"
      ;;
  esac
  case "${_curr_shell}" in
    *bash|*ksh|*zsh)
      if test "${_curr_shell}" = "bash"; then
        set -o errtrace
        set -o pipefail
      fi
      trap 'kbx_exit_handler $? ${LINENO:-}' ERR
      ;;
  esac
  trap 'kbx_exit_handler $? ${LINENO:-} ' EXIT HUP INT QUIT ABRT ALRM TERM
}

##########################################################################
# kbx_end
#
#
# Cleans up and exits the script.
# The function resets the exit trap and kills the tail process if it
# is running. The function also sets the kbx_last_exit variable to the
# last exit code. The function returns 0 on success and 1 on failure.
# The function also prints the kbx_log file location and size.
#
############################################################################
kbx_end() {
  ## Reset exit trap.
  trap - EXIT HUP INT QUIT ABRT ALRM TERM
  if test -n "${kbx_tail_pid:-}"; then
    sleep 0.3
    kill -9 ${kbx_tail_pid}
  fi
  exit "${kbx_last_exit:-1}"
}

##########################################################################
# kbx_exit_handler
#
#
# Handles exit signals and errors.
# The function handles exit signals and errors by logging the error
# message and the line number where the error occurred. The function
# also logs the source script and the exit code. The function prints
# the kbx_log file location and size. The function also prints the elapsed
# time and the exit code. The function returns 0 on success and 1 on
# failure.
#
############################################################################
kbx_exit_handler() {
  true "BEGIN kbx_exit_handler() with args: $*"
  export kbx_last_exit="${kbx_last_exit:-${1}}"
  local line_number="${2:-0}"
  local err_source_script="${3:-}"

  ## Exit without errors.
  test "${kbx_last_exit}" = "0" && kbx_end || true

  ## Exit with errors.
  # shellcheck disable=SC3028
  if test -n "${err_source_script:-}"; then
    kbx_log notice "Executed script: '${0}'"
    kbx_log notice "What was being run at the time of the error was: ${err_source_script}"
  fi

  ## some shells have a bug that displays line 1 as LINENO
  if test "${line_number}" -gt 2; then
    kbx_log error "${kubex_pkg_name:-Kubex} aborted due to an error."
    kbx_log error "No need to panic. Nothing is broken. Just some rare condition has been hit."
    kbx_log error "Please report this bug if it has not been already."
    kbx_br || true
    kbx_log error "An error occurred at line: '${line_number}'"

    ## Easy version for wasting resources and better readability.
    line_above="$(pr -tn "${0}" | tail -n+$((line_number - 4)) | head -n4)"
    line_error="$(pr -tn "${0}" | tail -n+$((line_number)) | head -n1)"
    line_below="$(pr -tn "${0}" | tail -n+$((line_number + 1)) | head -n4)"
    printf '%s\n*%s\n%s\n' "${line_above}" "${line_error}" "${line_below}"
    ## Too complex.
    # awk 'NR>L-4 && NR<L+4 { printf " %-5d%3s %s\n",NR,(NR==L?">>>":""),$0 }' L="${line_number}" "${0}" >&2

    kbx_br || true
    kbx_log error "Please include the user kbx_log and the debug kbx_log in your bug report." || true
    kbx_log error "(For file locations where to find these logs, see above.)" || true
    kbx_br || true
  else
    if [ "${kbx_last_exit}" -gt 128 ] && [ "${kbx_last_exit}" -lt 193 ]; then
      signal_code=$((kbx_last_exit-128))
      signal_caught="$(kill -l "${signal_code}")"
      kbx_log error "Signal received: '${signal_caught}'" || true
    fi
  fi

  ## Print exit code.
  if test "${kbx_last_exit}" -gt 0; then
    kbx_log error "Exit code: '${kbx_last_exit}'" || true
  fi
  kbx_log notice "Time elapsed: $(kbx_get_elapsed)s." || true
  kbx_log notice "Exiting with code: '${kbx_last_exit}'" || true
  
  ## Print the kbx_log file location.
  if test -n "${_kbx_log_output:-}"; then
    kbx_log notice "Log file: '${_kbx_log_output}'" || true
  fi
  kbx_log notice "Log file size: $(du -h "${_kbx_log_output}" | awk '{print $1}')" || true
  kbx_log notice "Log file location: $(dirname "${_kbx_log_output}")" || true
  
  kbx_end || exit "${kbx_last_exit}" || return 1
}
## </editor-fold>

##############################################################################
# kbx_clear_keeping_buffer
#
# Clears the terminal screen and keeps the buffer.
# The function uses ANSI escape codes to clear the screen and reset the cursor
# position. The function also uses the tput command to clear the screen and
# reset the cursor position. The function returns 0 on success and 1 on failure.
# The function also prints a message to the terminal.
#
##############################################################################
kbx_clear_keeping_buffer(){
  >&2 printf "\033[H\033[2J" >/dev/tty || >&2 tput -x clear >/dev/tty || >&2 clear >/dev/tty
  return 0
}

##############################################################################
# kbx_br
#
# Prints a blank line to the terminal.
# The function uses ANSI escape codes to print a blank line to the terminal.
# The function returns 0 on success and 1 on failure.
# The function also prints a message to the terminal.
#
##############################################################################
kbx_br () {
  >&2 printf "\n" >/dev/tty || >&2 echo "" >/dev/tty
  return 0
}

##############################################################################
# kbx_yes_no_question
#
# Asks a yes/no question and returns the answer.
# The function uses ANSI escape codes to print a yes/no question to the
# terminal. The function also uses the read command to read the answer
# from the user. The function returns 0 on success and 1 on failure.
# The function also prints a message to the terminal.
#
# The function uses a timeout to wait for the user to answer the question.
# The function also uses a default answer if the user does not answer
# Arguments:
#   $1 - question to ask
#   $2 - default answer (y/n)
#   $3 - timeout in seconds
#
##############################################################################
kbx_yes_no_question() {
  local _question="${1}"
  local _default_answer="${2:-}"
  local _timeout="${3:-5}"
  local _answer=""
  local _counter=0
  while [[ ! "$_answer" =~ ^[yYnN]$ ]]; do
    if [ "$_counter" -gt 3 ]; then
      kbx_log error "Maximum number of attempts reached."
      kbx_br || true
      return 1
    else
      _counter=$((_counter + 1))
    fi
    read -rt "${_timeout}" -n 1 -rp "${_question} " _answer || _answer="${_default_answer:-}"
    if [ -z "${_answer}" ]; then
      if [ -n "${_default_answer}" ]; then
        _answer="${_default_answer}"
      fi
    fi
  done
  echo "${_answer}"
  kbx_br || true
  return 0
}

##############################################################################
# kbx_read_secret_with_callback
#
# Reads a secret from the user and calls a callback function with the secret.
# The function uses ANSI escape codes to read a secret from the user.
# The function also uses the read command to read the secret from the user without
# echoing it to the terminal. The function returns 0 on success and 1 on failure.
# 
# The function returns 0 on success and 1 on failure.
# 
############################################################################
kbx_read_secret_with_callback() {
  local _prompt="${1}"
  local _callback="${2:-}"
  local _secret=""
  local _counter=0
  while [ -z "${_secret}" ]; do
    if [ "$_counter" -gt 3 ]; then
      kbx_log error "Maximum number of attempts reached."
      return 1
    else
      _counter=$((_counter + 1))
    fi
    read -rsp "${_prompt}" _secret >/dev/tty
    kbx_br || true
    if [ -n "${_callback}" ]; then
      "${_callback}" "${_secret}"
    fi
  done
  return 0
}

##############################################################################
# kbx_wait_with_escape
#
# Waits for a specified time and allows the user to escape the wait.
# The function uses ANSI escape codes to wait for a specified time and
# allows the user to escape the wait. The function also uses the read
# command to read the escape key from the user, while waiting for the
# specified time it will print endless dots or the specified character.
# The function returns 0 on success and 1 on failure.
#
# Arguments:
#   $1 - waiting time in seconds
#   $2 - message to display
#   $3 - character to display while waiting
#
############################################################################
kbx_wait_with_escape() {
  local _default_sleep_time=0500e-3
  local _BEAUTY_SLEEP_TIME="${_BEAUTY_SLEEP_TIME:-$_default_sleep_time}"
  local _waiting_char="."
  local _message=""
  local _waiting_time=0
  local _waiting_time_limit=5

  if [[ -n "${1:-}" && "${1:-}" != "" && "$1" =~ ^[0-9]+$ ]]; then
    _waiting_time_limit=$((1 * "${1}"))
  fi
  if [ -n "${2:-}" ] && [ "${2:-}" != "" ]; then
    _message="${2:-}"
  else
    _message=""
  fi
  if [ -n "${3:-}" ] && [ "${3:-}" != "" ]; then
    _waiting_char="${3:-.}"
  fi

  if [ -n "${_message}" ] && [ "${_message}" != "" ]; then
    if [ $_waiting_time_limit -gt 4 ]; then
      _message="${_message}"
    fi
    kbx_log notice "${_message}"
  fi
  
  local _half_second_marker=""
  local _escape_wait_dummy=""
  local _waiting_char_acum="s) "
  local _remaining_time=$((_waiting_time_limit - _waiting_time))
  local _remaining_message="${_remaining_time}"
  while [ "${_waiting_time}" -lt "${_waiting_time_limit}" ]; do
    printf "\r%s%d%s" "${_message} (" "${_remaining_time}" "${_waiting_char_acum}"
    if read -rt 0.5 -n 1 -s; then
      break
    else
      sleep "${_BEAUTY_SLEEP_TIME}"
      if [[ -z "${_half_second_marker}" ]]; then
        _half_second_marker="1"
        _waiting_char_acum="${_waiting_char_acum:-}${_waiting_char:-}"
      else
        _half_second_marker=""
        _waiting_time=$((1 + _waiting_time))
        _remaining_time=$(( _waiting_time_limit - _waiting_time ))
        _waiting_char_acum+="${_waiting_char:-}"
      fi
    fi
  done

  kbx_br || true

  return 0
}


##############################################################################
# kbx_std_log_selector
#
# Selects the standard kbx_log file based on the kbx_log level and type.
# The function uses ANSI escape codes to select the standard kbx_log file
# based on the kbx_log level and type. The function also uses the kbx_log_level
# and kbx_log_file variables to determine the kbx_log level and file. The
# function returns the kbx_log file path based on the kbx_log level and type.
# The function returns 0 on success and 1 on failure.
#
# Arguments:
#   $1 - kbx_log level (info, error, verbose, debug)
#   $2 - kbx_log type (bug, error, fatal, critical, emergency, warn, alert, info,
#       answer, notice, success, action, debug, trace, verbose)
#
############################################################################
kbx_std_log_selector() {
  local _log_level="${1:-${kbx_log_level:-${log_level:-info}}}"
  local _log_file="${kbx_log_file:-${log_file:-/dev/null}}"
  local _log_type="${2:-notice}"
  #####################
  ## kbx_log LEVELS: INFO ERROR VERBOSE DEBUG
  ## kbx_log TYPES: bug error fatal critical emergency warn alert info answer notice success action debug trace verbose
  ######################
  ## CLI LEVEL VISIBLE TYPES: bug fatal critical emergency
  ## INFO LEVEL VISIBLE TYPES: bug error fatal critical emergency warn alert info answer
  ## ERROR LEVEL VISIBLE TYPES: bug error fatal critical emergency answer
  ## VERBOSE LEVEL VISIBLE TYPES: bug error fatal critical emergency warn alert info answer notice success action verbose
  ## DEBUG LEVEL VISIBLE TYPES: bug error fatal critical emergency warn alert info answer notice success action debug trace verbose
  ######################
  case "${_log_level}" in
    debug)
      case "${_log_type}" in
        bug|error|fatal|critical|emergency|warn|alert|info|answer|notice|success|action|debug|trace|verbose)
          echo "/dev/tty"
          return 0
          ;;
        *)
          echo "/dev/null"
          return 0
          ;;
      esac
      ;;
    verbose)
      case "${_log_type}" in
        bug|error|fatal|critical|emergency|warn|alert|info|answer|notice|success|action|verbose)
          echo "/dev/tty"
          return 0
          ;;
        *)
          echo "/dev/null"
          return 0
          ;;
      esac
      ;;
    info)
      case "${_log_type}" in
        bug|error|fatal|critical|emergency|warn|alert|info|answer)
          echo "/dev/tty"
          return 0
          ;;
        *)
          echo "/dev/null"
          return 0
          ;;
      esac
      ;;
    error)
      case "${_log_type}" in
        bug|error|fatal|critical|emergency|answer)
          echo "/dev/tty"
          return 0
          ;;
        *)
          echo "/dev/null"
          return 0
          ;;
      esac
      ;;
    cli|null)
      case "${_log_type}" in
        bug|error|fatal|critical|emergency)
          echo "/dev/tty"
          return 0
          ;;
        *)
          echo "/dev/null"
          return 0
          ;;
      esac
      ;;
    *)
      export log_level="verbose"
      printf '\n%b\n' "[${_red}ERROR${_nocolor}]: Unsupported kbx_log level specified: ${_yellow}'${_log_level}'${_nocolor}" >/dev/tty
      printf '%b\n\n' "[${_green}VALID${_green}]: ${_blue}debug${_nocolor}, ${_blue}verbose${_nocolor}, ${_blue}info${_nocolor}, ${_blue}error${_nocolor}, ${_blue}cli${_nocolor}" >/dev/tty
      # mata o script enviando um exit 1 para toda árvore de processos do kubex usando o exec ou o kill -9
      exit 1 || kill -9 $$
      ;;
  esac
}

##########################################################################
# kbx_log
#
# Logs messages with different levels (info, warn, error, success).
# The function takes a kbx_log level and a message as arguments and logs the
# message to a kbx_log file. The function also supports different output formats
# based on the kbx_log level. The function creates a temporary directory for logs
# if it does not exist and sets the kbx_log file path in the _kbx_log_output
# variable. The function uses colors for different kbx_log levels and formats
# the output accordingly. The function also supports a debug mode that
# logs messages to both the kbx_log file and the console.
# 
# kbx_log messages with different levels
# Arguments:
#   $1 - kbx_log level (info, warn, error, success)
#   $2 and above - message to log, can be multiple arguments. It will be
#                 concatenated into a single string.
# 
############################################################################
kbx_log(){
  ## Avoid clogging output if kbx_log() is working alright.
  if test "${xtrace:-}" = "1"; then
    set +o xtrace
  else
    case "${-}" in
      *x*)
        xtrace=1
        set +o xtrace
        ;;
    esac
  fi

  # Check if the output folder exists, if not, create a new temporary directory
  # and set the output file into the variable _TEMP_DIR for trap cleanup
  if [ ! -d "${_TEMP_DIR:-}" ]; then
    _kbx_create_temp_dir || return 1
  fi
  # Check if the output file exists, if not, create it
  if [ ! -f "$_kbx_log_output" ]; then
    touch "$_kbx_log_output"
  else 
    _kbx_create_temp_dir || return 1
  fi

  local log_type="${1:-notice}"
  ## capitalize kbx_log level
  local log_type_up="$(printf '%s' "${log_type}" | tr "[:lower:]" "[:upper:]")"
  shift 1
  ## set formatting based on kbx_log level
  local log_color=""
  case "${log_type}" in
    bug)
      log_color="${_underline}${_magenta}"
      ;;
    error|fatal|critical|emergency)
      log_color="${_red}"
      ;;
    warn|alert)
      log_color="${_yellow}"
      ;;
    info|answer)
      log_color="${_cyan}"
      ;;
    success|action)
      log_color="${_green}"
      ;;
    notice|verbose)
      log_color="${_nocolor}"
      ;;
    debug|trace)
      log_color="${_blue}"
      ;;
    null)
      log_color=""
      true
      ;;
    *)
      kbx_log bug "Unsupported kbx_log type specified: '${log_type}'"
      kbx_die 1 "Please report this bug."
  esac
  ## uniform kbx_log format
  log_color="${_bold}${log_color}"
  if [ "${kbx_log_source_script:-0}" != 0 ]; then
    local log_source_script="${BASH_SOURCE-}"
  fi
  local log_level_colorized="[${log_color}${log_type_up}${_nocolor}]: "
  local log_content="${*}"
  local log_file="$(kbx_std_log_selector "${kbx_log_level:-info}" "${log_type}")"
  if [ -n "${log_file}" ]; then
    if [ -n "${log_source_script:-}" ]; then
      printf '%b %b %b\n' "$(date +'%Y-%m-%d %H:%M:%S')" "${log_level_colorized}" "${log_content}" >> "${log_file}"
    else
      printf '%b %b\n' "${log_level_colorized}" "${log_content}" >> "${log_file}"
    fi
  fi
  if test "${xtrace:-}" = "1"; then
    set -o xtrace
  fi
  return 0
}

##########################################################################
# kbx_die
#
# Handles all exit codes and errors.
# The function takes an exit code and an error message as arguments
# and logs the error message. 
# 
###########################################################################
kbx_die(){
  export kbx_start_time="${kbx_start_time:-$(date +%s)}"
  kbx_log error "${2:-An error occurred.}"
  if test "${1:-1}" = "0"; then
    kbx_log notice "Exiting without errors."
    >&2 kbx_br >/dev/tty || true
    kbx_exit_handler 0 || exit 0 || return 0
  fi
  if test "${kbx_allow_errors:-0}" = "1"; then
    kbx_log warn "Skipping termination because of with code '${1}' due to 'kbx_allow_errors' setting."
    return 0
  fi
  case "${1:-1}" in
    0|333|666|999)
      true
      ;;
    *)
      kbx_log error "Kubex aborting..."
      ;;
  esac
  local _line_number=$((1 * "${3:-0}")) || true
  local _error_source="${4:-}"
  export kbx_last_exit=$((1 * "${1:-1}")) || true
  exit ${kbx_last_exit:-1} || kill -9 $$ || return 1
}

##########################################################################
# kbx_run
#
# Runs a command in the background or foreground, logging the output,
# handling errors and registering the time elapsed to run the command.
# The function will sanitize the command before running it, set the
# necessary environment variables and run provided command.
# 
##########################################################################
kbx_run(){
  export kbx_start_time="${kbx_start_time:-$(date +%s)}"
  local level="${1:-info}"
  shift

  # shellcheck disable=SC2116
  ## Extra spaces appearing when breaking kbx_log_run on multiple lines.
  local cmd_no_space="$(echo "${@}")"
  cmd_no_space="${cmd_no_space//  / }"
  cmd_no_space="${cmd_no_space/route_exec/kubex}"
  if test "${dry_run:-}" = "1"; then
    kbx_log "${level}" "Skipping command: $ ${cmd_no_space}"
    return 0
  fi
  if test "${kbx_run_background:-}" = "1"; then
    kbx_log "${level}" "Background command starting: "
    kbx_log "${level}" "$ ${cmd_no_space} &"
    "${@}" &
    local kbx_background_pid="$!"
    disown "$kbx_background_pid"
  else
    kbx_log "${level}" "Loading command: " &&\
    kbx_log "${level}" "$ ${cmd_no_space}" &&\
    "${@}"
    kbx_exit_handler $?
  fi
}

##########################################################################
# kbx_get_elapsed
#
# Gets the elapsed time since the start of the script without
# echoing it to the terminal and without resetting the cursor position.
# The function uses ANSI escape codes to get the elapsed time since
# the start of the script. The function also uses the date command
# to get the elapsed time since the start of the script. 
#
############################################################################
kbx_get_elapsed(){
  export kbx_start_time="${kbx_start_time:-$(date +%s)}"
  printf '%s\n' "$(($(date +%s) - kbx_start_time))" || return 1
  return 0
}

##########################################################################
# kbx_print_elapsed
#
# Prints the elapsed time since the start of the script to the terminal default stdout.
# The function uses ANSI escape codes to print the elapsed time since
# the start of the script. The function also uses the date command
# to get the elapsed time since the start of the script.
# The function returns 0 on success and 1 on failure.
#
############################################################################
kbx_print_elapsed(){
  export kbx_start_time="${kbx_start_time:-$(date +%s)}"
  kbx_log notice "Time elapsed: $(kbx_get_elapsed)s."
}
## </editor-fold>

# Script entry point
if test "${kbx_log_loaded:-0}" != "1"; then
  export kbx_log_loaded=1
  __log_main_functions
  if test "${BASH_SOURCE-}" = "${0}"; then
    kbx_set_trap
  fi
fi
## </editor-fold>

