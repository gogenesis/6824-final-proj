#!/usr/bin/env bash 

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" 

run_test () {
   local name="$1"
   echo "Start $name" 
   if [ ! -z $JENKINS ]; then
      /usr/local/go/bin/go test -run $name
   else
      go test -run $name
   fi
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
   run_test "MemoryFS_TestDeleteNotFound"
   run_test "MemoryFS_TestCloseClosed"
   run_test "MemoryFS_TestOpenOpened"
   run_test "MemoryFS_TestOpenNotFound"
   run_test "MemoryFS_TestOpenAlreadyExists"
   run_test "MemoryFS_TestOpenROClose"
   run_test "MemoryFS_TestOpenROClose"
   run_test "MemoryFS_TestOpenROClose4"
   run_test "MemoryFS_TestOpenROClose64"
   run_test "MemoryFS_TestOpenRWClose"
   run_test "MemoryFS_TestOpenRWClose4"
   run_test "MemoryFS_TestOpenRWClose64"
   run_test "MemoryFS_TestOpenCloseLeastFD"
   run_test "MemoryFS_TestOpenCloseDeleteRoot"
   run_test "MemoryFS_TestOpenCloseDeleteMaxFD"
   run_test "MemoryFS_TestOpenCloseDeleteRootMax"
   run_test "MemoryFS_TestSeekErrorBadFD"
   run_test "MemoryFS_TestSeekErrorBadOffsetOperation"
   run_test "MemoryFS_TestWriteClosedFile"
   run_test "MemoryFS_TestWrite1Byte"
   run_test "MemoryFS_TestWrite1KByte"
   run_test "MemoryFS_TestWrite1MByte"
   run_test "MemoryFS_TestWrite10MByte"
   run_test "MemoryFS_TestWrite100MByte"
   run_test "MemoryFS_TestReadClosedFile"
   echo ""
   popd > /dev/null
}

main $*

