#!/usr/bin/env bash
# shellcheck disable=SC1091,SC2154

###############################################################
# __wrapper
# 
# This is a wrapper script to run the kubex modules installation, 
# building, uninstalling, cleaning and testing.It is not meant to 
# be run directly to avoid any issues and protect the environment 
# and the user. It is meant to be run by the Makefile or other scripts.
# 
#########################################################################
. "$(dirname "$(readlink -e "${0}")")/wrapper.sh"
__wrapper "$@" || exit $?
