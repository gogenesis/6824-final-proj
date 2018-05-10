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

generate_tests () {
   go generate
}

main () {
   pushd `pwd` > /dev/null
   cd $SCRIPT_DIR/../main 
   go generate
   if [ $? != 0 ]; then
      echo "FAIL: test generation... continuing to run all tests"
   else
      echo "Tests generated successfully!"
   fi
   cd $SCRIPT_DIR
   # As tests begin passing, to keep them included in future test runs,
   # they should be added here.
	  run_test "TestMemoryFS_TestBasicOpenClose"
	  run_test "TestMemoryFS_TestDeleteNotFound"
	  run_test "TestMemoryFS_TestCloseClosed"
	  run_test "TestMemoryFS_TestOpenOpened"
	  run_test "TestMemoryFS_TestOpenNotFound"
	  run_test "TestMemoryFS_TestOpenAlreadyExists"
	  run_test "TestMemoryFS_TestOpenROClose"
	  run_test "TestMemoryFS_TestOpenROClose"
	  run_test "TestMemoryFS_TestOpenROClose4"
	  run_test "TestMemoryFS_TestOpenROClose64"
	  run_test "TestMemoryFS_TestOpenRWClose"
	  run_test "TestMemoryFS_TestOpenRWClose4"
	  run_test "TestMemoryFS_TestOpenRWClose64"
	  run_test "TestMemoryFS_TestOpenCloseLeastFD"
	  run_test "TestMemoryFS_TestOpenCloseDeleteMaxFD"
	  run_test "TestMemoryFS_TestOpenCloseDeleteRoot"
	  run_test "TestMemoryFS_TestOpenCloseDeleteRootMax"
	  run_test "TestMemoryFS_TestSeekErrorBadFD"
	  run_test "TestMemoryFS_TestSeekErrorBadOffsetOperation"
	  run_test "TestMemoryFS_TestSeekOffEOF"
	  run_test "TestMemoryFS_TestWriteClosedFile"
	  run_test "TestMemoryFS_TestWriteReadBasic"
	  run_test "TestMemoryFS_TestWriteReadBasic4"
	  run_test "TestMemoryFS_TestWrite1Byte"
	  run_test "TestMemoryFS_TestWrite8Bytes"
	  run_test "TestMemoryFS_TestWrite1KBytes"
	  run_test "TestMemoryFS_TestWrite1MBytes"
	  run_test "TestMemoryFS_TestWrite10MBytes"
	  run_test "TestMemoryFS_TestWrite100MBytes"
	  run_test "TestMemoryFS_TestReadClosedFile"
	  run_test "TestMemoryFS_TestWriteRead1ByteSimple"
	  run_test "TestMemoryFS_TestWriteRead8BytesSimple"
	  run_test "TestMemoryFS_TestWriteRead8BytesIter8"
	  run_test "TestMemoryFS_TestWriteRead8BytesIter64"
	  run_test "TestMemoryFS_TestWriteRead64BytesIter64K"
	  run_test "TestMemoryFS_TestWriteRead64KBIter1MB"
	  run_test "TestMemoryFS_TestWriteRead64KBIter10MB"
	  run_test "TestMemoryFS_TestWriteRead1MBIter100MB"
	  run_test "TestMemoryFS_TestRndWriteReadVerfiyHoleExpansion"

	  # ======= the line in the sand ======

	  #@dir
	  #run_test "TestMemoryFS_TestMkdir"
	  #run_test "TestMemoryFS_TestMkdirTree"
	  #run_test "TestMemoryFS_TestOpenCloseDeleteAcrossDirectories"
# generated Tue May  8 20:40:51 EDT 2018 by $ main/generate_run_tests.sh MemoryFS
   echo ""
   popd > /dev/null
}

main $*

