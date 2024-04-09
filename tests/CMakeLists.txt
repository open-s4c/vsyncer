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
include(CTest)
# Include vsyncer.cmake to have add_vsyncer_check available.
include(../cmake/vsyncer.cmake)

option(VSYNCER_LOCAL "Uses local vsyncer, instead of install one" ON)
set(LOG ${CMAKE_CURRENT_BINARY_DIR}/log.csv)

file(GLOB CLIENTS *.c)
set(CMDS compile check)
set(ARGS_check --csv-log ${LOG})

# tests with Status other than CheckOk
set(EXPECT_check_hang "CheckNotLive")
set(EXPECT_check_zero_exec "CheckRejected")
set(EXPECT_check_sigsev "CheckNotSafe")

# tests that will crash
set(TRIGGERS_MC_BUG #
    check_genmc9_bug_genmc9 #
    check_error #
)

if(VSYNCER_LOCAL)
    set(VSYNCER_DIR ${CMAKE_SOURCE_DIR}/..)
    execute_process(COMMAND make build  WORKING_DIRECTORY ${VSYNCER_DIR})
    set(VSYNCER ${VSYNCER_DIR}/build/vsyncer)
    set(CHECKERS genmc)
    set(PATH_genmc genmc)
    # Since we do not build the source files, we have to reference the source files
    # at the source directory
    set(SOURCES ok.c t1.c cmpxchg.c)
    list(TRANSFORM SOURCES PREPEND ${CMAKE_CURRENT_SOURCE_DIR}/)
    foreach(SRC ${SOURCES})
      # add_vsyncer_check works as a normal add_test, but the command runs inside
      # the vsyncer Docker container (unless VSYNCER_DOCKER=off).
      add_vsyncer_check(NAME ${SRC}-check COMMAND vsyncer check ${SRC})
    endforeach()
else()
    set(VSYNCER vsyncer)
    # add dartagnan to CHECKERS list once the expectatio is clear
    set(CHECKERS genmc9 genmc10)
    set(PATH_genmc9 /usr/share/genmc9/bin/genmc)
    set(PATH_genmc10 /usr/share/genmc10/bin/genmc)
endif()

foreach(CMD ${CMDS})
    foreach(CLIENT ${CLIENTS})
        foreach(MC ${CHECKERS})
            get_filename_component(CASE_NAME ${CLIENT} NAME_WLE)

            set(CASE_NAME ${CMD}_${CASE_NAME})
            set(TEST_NAME ${CASE_NAME}_${MC})

            if("${MC}" STREQUAL "dartagnan")
                add_test(NAME ${TEST_NAME} COMMAND ${VSYNCER} --checker ${MC} ${CMD} ${ARGS_${CMD}} ${CLIENT})
            else()
                add_test(NAME ${TEST_NAME} COMMAND env GENMC_CMD=${PATH_${MC}} ${VSYNCER} ${CMD} ${ARGS_${CMD}} ${CLIENT})
            endif()

            # if it triggers model checker bug expects to return non-zero
            if(${TEST_NAME} IN_LIST TRIGGERS_MC_BUG OR ${CASE_NAME} IN_LIST TRIGGERS_MC_BUG)
                set_property(TEST ${TEST_NAME} PROPERTY WILL_FAIL TRUE)
            # if the expecation exists and the same for all model checkers
            elseif(EXPECT_${CASE_NAME})
                set_property(TEST ${TEST_NAME} PROPERTY PASS_REGULAR_EXPRESSION ${EXPECT_${CASE_NAME}})
            # if the expecation exists for a particular model checker
            elseif(EXPECT_${TEST_NAME})
                set_property(TEST ${TEST_NAME} PROPERTY PASS_REGULAR_EXPRESSION ${EXPECT_${TEST_NAME}})
            endif()
        endforeach()
    endforeach()
endforeach()
