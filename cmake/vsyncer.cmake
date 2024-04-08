# ##############################################################################
# Copyright (C) 2024 Huawei Technologies Co., Ltd. All rights reserved.
# SPDX-License-Identifier: MIT
# ##############################################################################
#
# Include this file in your CMake project to have add_vsyncer_check() function
# available.
#

if(EXISTS "/.dockerenv")
  set(DEFAULT_DOCKER off)
else()
  set(DEFAULT_DOCKER on)
endif()

option(VSYNCER_DOCKER "Use Docker to run vsyncer checks" ${DEFAULT_DOCKER})
option(VSYNCER_CHECK "Whether vsyncer checks are enabled" ON)

set(VSYNCER_DOCKER_IMAGE
    "ghcr.io/open-s4c/vsyncer"
    CACHE STRING "Docker image with vsyncer and model checkers")

set(VSYNCER_DOCKER_TAG
    "main"
    CACHE STRING "Tag of docker image with vsyncer")

execute_process(
  COMMAND bash -c "paste -d: <(id -u) <(id -g)"
  OUTPUT_VARIABLE VSYNCER_DOCKER_UID_GID
  OUTPUT_STRIP_TRAILING_WHITESPACE)

# `add_vsyncer_check` adds a target running the the vsyncer Docker container.
function(add_vsyncer_check)
  set(opts)
  set(ones NAME COMMAND)
  set(mult)
  cmake_parse_arguments(ARG "${opts}" "${ones}" "${mult}" ${ARGN})

  set(DOCKER_CMD docker run --rm)
  list(APPEND DOCKER_CMD -u ${VSYNCER_DOCKER_UID_GID})
  list(APPEND DOCKER_CMD -v ${PROJECT_SOURCE_DIR}:${PROJECT_SOURCE_DIR})
  list(APPEND DOCKER_CMD -v ${PROJECT_BINARY_DIR}:${PROJECT_BINARY_DIR})
  list(APPEND DOCKER_CMD -w ${CMAKE_CURRENT_BINARY_DIR})
  list(APPEND DOCKER_CMD ${VSYNCER_DOCKER_IMAGE}:${VSYNCER_DOCKER_TAG})

  if(${VSYNCER_CHECK})
    add_test(NAME ${ARG_NAME} COMMAND ${DOCKER_CMD} ${ARG_COMMAND}
                                      ${ARG_UNPARSED_ARGUMENTS})
  endif()
endfunction()
