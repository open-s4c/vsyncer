cmake_minimum_required(VERSION 3.16)
project(
    vsyncer-tests
    LANGUAGES C
    DESCRIPTION
        "Integration tests of Vsyncer"
)

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
option(VSYNCER_LOCAL "Uses local vsyncer, instead of install one" ON)

include(CTest)
file(GLOB CLIENTS *.c)
set(CMDS compile check optimize)

set(EXPECT_check_hang "CheckNotLive")
set(EXPECT_check_zero_exec "CheckRejected")
set(EXPECT_check_sigsev "CheckNotSafe")
# this test should result in sth else this
# expectation is inaccurate
set(EXPECT_check_error "ERROR")

if(VSYNCER_LOCAL)
    set(VSYNCER_DIR ${CMAKE_SOURCE_DIR}/..)
    execute_process(COMMAND make build  WORKING_DIRECTORY ${VSYNCER_DIR})
    set(VSYNCER ${VSYNCER_DIR}/build/vsyncer)
else()
    set(VSYNCER vsyncer)
endif()


foreach(CMD ${CMDS})
    foreach(CLIENT ${CLIENTS})
        get_filename_component(TEST_NAME ${CLIENT} NAME_WLE)
        set(TEST_NAME ${CMD}_${TEST_NAME})
        add_test(NAME ${TEST_NAME} COMMAND ${VSYNCER} ${CMD} ${CLIENT})
        if(EXPECT_${TEST_NAME})
            set_property(TEST ${TEST_NAME} PROPERTY PASS_REGULAR_EXPRESSION ${EXPECT_${TEST_NAME}})
        endif()
    endforeach()
endforeach()
