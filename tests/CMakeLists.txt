cmake_minimum_required(VERSION 3.16)
project(
  vsyncer-tests
  LANGUAGES C
  DESCRIPTION "Integration tests of vsyncer")

enable_testing()

# if in GHA pipeline, use installed vsyncer, other use build
option(VSYNCER_LOCAL "Uses local vsyncer, instead of installed one" ON)
# if("$ENV{GITHUB_ACTIONS}" STREQUAL "true")
if(NOT "${VSYNCER_LOCAL}")
  set(VSYNCER "vsyncer")
else()
  # if running locally, use docker container

  set(VSYNCER_DIR "${CMAKE_SOURCE_DIR}/..")
  set(VSYNCER "${VSYNCER_DIR}/build/vsyncer")

  # we have to mount in the docker container the root of the vsyncer project
  set(ENVVARS "VSYNCER_DOCKER_VOLUMES=${VSYNCER_DIR}")
endif()

# ##############################################################################
# Integration tests
# ##############################################################################

set(LOG ${CMAKE_CURRENT_BINARY_DIR}/log.csv)
# fetch all c files in the current folder in CLIENTS
file(GLOB CLIENTS *.c)
# vsyncer commands
set(CMDS compile check optimize)
# check extra arguments
set(ARGS_check --csv-log ${LOG})

# define tests expectation
set(EXPECT_check_hang "CheckNotLive")
set(EXPECT_check_zero_exec "CheckRejected")
set(EXPECT_check_sigsev "CheckNotSafe")

# list tests that causes an error/bug
set(TRIGGERS_MC_BUG #
    check_genmc9_bug_genmc9 #
    check_error #
)

# list of checkers to test
set(CHECKERS dartagnan genmc9 genmc10)

# mappings to --checker option
set(CHECKER_dartagnan dartagnan)
set(CHECKER_genmc9 genmc)
set(CHECKER_genmc10 genmc)

set(GENMC_genmc9 /usr/share/genmc9/bin/genmc)
set(GENMC_genmc10 /usr/share/genmc10/bin/genmc)

# for each command client and model checker add a test
foreach(CMD ${CMDS})
  foreach(CLIENT ${CLIENTS})
    foreach(MC ${CHECKERS})
      get_filename_component(CASE_NAME ${CLIENT} NAME_WLE)

      set(CASE_NAME ${CMD}_${CASE_NAME})
      set(TEST_NAME ${CASE_NAME}_${MC})

      add_test(
        NAME ${TEST_NAME}
        COMMAND env ${ENVVARS} GENMC_PATH=${GENMC_${MC}} ${VSYNCER} --checker
                ${CHECKER_${MC}} -d ${CMD} ${ARGS_${CMD}} ${CLIENT})

      # if it triggers model checker bug expect to return non-zero value
      if(${TEST_NAME} IN_LIST TRIGGERS_MC_BUG OR ${CASE_NAME} IN_LIST
                                                 TRIGGERS_MC_BUG)
        set_property(TEST ${TEST_NAME} PROPERTY WILL_FAIL TRUE)
        # if the expecation exists for all model checkers
      elseif(EXPECT_${CASE_NAME})
        set_property(TEST ${TEST_NAME} PROPERTY PASS_REGULAR_EXPRESSION
                                                ${EXPECT_${CASE_NAME}})
        # if the expecation exists for a particular model checker
      elseif(EXPECT_${TEST_NAME})
        set_property(TEST ${TEST_NAME} PROPERTY PASS_REGULAR_EXPRESSION
                                                ${EXPECT_${TEST_NAME}})
      endif()
    endforeach()
  endforeach()
endforeach()
