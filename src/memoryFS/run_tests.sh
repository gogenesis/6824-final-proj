#!/usr/bin/env bash 

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" 

run_test () {
   local name="$1"
   echo "Start $name" 
   go test -run $name
   exit_code=$?
   if [ $exit_code != 0 ]; then
      exit $exit_code 
   fi
}

main () {
   pushd `pwd` > /dev/null
   cd $SCRIPT_DIR
   # As tests begin passing, to keep them included in future test runs,
   # they should be added here.
   run_test "MemoryFS_TestBasicOpenClose"
   run_test "MemoryFS_TestOpenROClose"
   run_test "MemoryFS_TestOpenRWClose"
   run_test "MemoryFS_TestOpenROClose4"
   run_test "MemoryFS_TestOpenROClose64"
   run_test "MemoryFS_TestOpenRWClose4"
   run_test "MemoryFS_TestOpenRWClose64"
   run_test "MemoryFS_TestOpenCloseLeastFD"
   run_test "MemoryFS_TestOpenCloseDeleteFD128"
   popd > /dev/null
}

main $*

