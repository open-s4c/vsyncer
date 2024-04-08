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
set(LOG ${CMAKE_CURRENT_BINARY_DIR}/log.csv)

include(CTest)
file(GLOB CLIENTS *.c)
set(CMDS compile check optimize)
set(ARGS_check --csv-log ${LOG})

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
    set(CHECKERS genmc)
    set(PATH_genmc genmc)
else()
    set(VSYNCER vsyncer)
    set(CHECKERS genmc9 genmc10)
    set(PATH_genmc9 /usr/share/genmc10/genmc9)
    set(PATH_genmc10 /usr/share/genmc10/genmc10)
endif()

foreach(CMD ${CMDS})
    foreach(CLIENT ${CLIENTS})
        foreach(MC ${CHECKERS})
            get_filename_component(TEST_NAME ${CLIENT} NAME_WLE)
            set(TEST_NAME ${CMD}_${TEST_NAME})
            add_test(NAME ${TEST_NAME}_${MC} COMMAND env GENMC_CMD=${PATH_${MC}} ${VSYNCER} ${CMD} ${ARGS_${CMD}} ${CLIENT})
            if(EXPECT_${TEST_NAME})
                set_property(TEST ${TEST_NAME}_${MC} PROPERTY PASS_REGULAR_EXPRESSION ${EXPECT_${TEST_NAME}})
            endif()
        endforeach()
    endforeach()
endforeach()
