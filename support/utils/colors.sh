#!/usr/bin/env bash
# shellcheck disable=SC2155

########################################################################
# kbx_detect_background_color
# 
# Detects the background color of the terminal.
#
###########################################################################
kbx_detect_background_color() {
  local _query_response=""
  local _test="$(printf "\e]11;?\a")" >/dev/null || return 1
  _query_response="$(echo -e "${_test}")" >/dev/null || return 1
  local _detected_background_color="${_query_response//[^0-9;:]/}" >/dev/null || return 1
  _detected_background_color="${_detected_background_color//\;/}" >/dev/null || return 1
  if test -z "$_detected_background_color"; then
    echo "rgb:00"
    return 0
  fi
  _detected_background_color="rgb:${_detected_background_color}"
  echo "${_detected_background_color:-}"
  return 0
}

#########################################################################
# kbx_get_high_contrast_color
#
# Determines the high contrast color based on the background color.
# The function takes a background color as an argument and returns
# either "dark" or "white" based on the brightness of the color.
#
##########################################################################
kbx_get_high_contrast_color() {
  local _bg_color="${1:-}"
  if [[ "${_bg_color}" =~ ^(rgb:ff|rgb:ee|rgb:dd|rgb:cc|rgb:bb|rgb:aa|rgb:99|rgb:88|rgb:77|rgb:66|rgb:55|rgb:44|rgb:33|rgb:22|rgb:11|rgb:00) ]]; then
    echo "dark"
  else
    echo "white"
  fi
  return 0
}

#########################################################################
# kbx_get_colors
#
# Sets up color variables for terminal output.
# The function checks if colors are supported in the terminal and
# sets up color variables accordingly.
# If colors are not supported, it sets the variables to empty strings.
# The function also checks for the background color and sets the
# foreground colors based on the contrast with the background.
#
# The function exports the following variables:
#   - _nocolor: No color
#   - _bold: Bold text
#   - _nobold: Normal text
#   - _underline: Underlined text    
#   - _nounderline: Normal text
#   - _red: Red text
#   - _green: Green text
#   - _yellow: Yellow text
#   - _blue: Blue text
#   - _magenta: Magenta text
#   - _cyan: Cyan text
#
############################################################################
kbx_get_colors(){
  # kbx_log levels
  if test -n "${NO_COLOR:-}" || test -n "${ANSI_COLORS_DISABLED:-}"; then
    export _nocolor=""
    export _bold=""
    export _nobold=""
    export _underline=""
    export _nounderline=""
    export _red=""
    export _green=""
    export _yellow=""
    export _blue=""
    export _magenta=""
    export _cyan=""
    return 0
  fi

  local _light_red="\033[38;5;196m"
  local _dark_red="\033[38;5;160m"

  local _light_green="\033[38;5;40m"
  local _dark_green="\033[38;5;34m"

  local _light_yellow="\033[38;5;220m"
  local _dark_yellow="\033[38;5;214m"

  local _light_blue="\033[38;5;39m"
  local _dark_blue="\033[38;5;21m"

  local _light_magenta="\033[38;5;207m"
  local _dark_magenta="\033[38;5;165m"

  local _light_cyan="\033[38;5;116m"
  local _dark_cyan="\033[38;5;45m"

  local _kbx_detected_background_color=$(kbx_detect_background_color) >/dev/null || return 1  
  local _kbx_contrast=$(kbx_get_high_contrast_color "$_kbx_detected_background_color") >/dev/null || return 1

  export _nocolor="\033[0m"
  export _bold="\033[1m"
  export _nobold="\033[22m"
  export _underline="\033[4m"
  export _nounderline="\033[24m"

  if test "$_kbx_contrast" == "dark"; then
    export _red="$_light_red"
    export _green="$_light_green"
    export _yellow="$_light_yellow"
    export _blue="$_light_blue"
    export _magenta="$_light_magenta"
    export _cyan="$_light_cyan"
  else
    export _red="$_dark_red"
    export _green="$_dark_green"
    export _yellow="$_dark_yellow"
    export _blue="$_dark_blue"
    export _magenta="$_dark_magenta"
    export _cyan="$_dark_cyan"
  fi
}

# Check if the script is being sourced, if so, load the functions
# if not, exit with an error
if test "${kbx_get_colors_loaded:-0}" != 1; then
  if test "${kbx_detect_bg_color_loaded:-0}" != 1; then
    export kbx_detect_bg_color_loaded=1
    export kbx_detect_background_color
  fi
  if test "${kbx_get_high_contrast_color_loaded:-0}" != 1; then
    export kbx_get_high_contrast_color_loaded=1
    export kbx_detect_background_color
  fi

  export kbx_get_colors_loaded=1

  kbx_get_colors
fi
