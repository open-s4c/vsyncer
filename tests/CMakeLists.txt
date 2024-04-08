cmake_minimum_required(VERSION 3.22)

project(vsyncer-tests LANGUAGES C)

# ##############################################################################
# vsyncer checks
# ##############################################################################

# add_vsyncer_check function depends on add_test, so call enable_testing.
enable_testing()

# Include vsyncer.cmake to have add_vsyncer_check available.
include(../cmake/vsyncer.cmake)

# Since we do not build the source files, we have to reference the source files
# at the source directory
set(SOURCES ok.c t1.c cmpxchg.c)
list(TRANSFORM SOURCES PREPEND ${CMAKE_CURRENT_SOURCE_DIR}/)

foreach(SRC ${SOURCES})
  # add_vsyncer_check works as a normal add_test, but the command runs inside
  # the vsyncer Docker container (unless VSYNCER_DOCKER=off).
  add_vsyncer_check(NAME ${SRC}-check COMMAND vsyncer check ${SRC})
endforeach()
