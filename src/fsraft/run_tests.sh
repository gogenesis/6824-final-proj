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

generate_tests () {
   cd $SCRIPT_DIR/../main 
   if [ ! -z $GOBIN ]; then
      $GOBIN generate
      return $?
   else
      echo "Skipping generation because GOBIN not set to point to a go binary"
   fi
}

main () {
   pushd `pwd` > /dev/null
   generate_tests
   if [ $? != 0 ]; then
      echo "Test generation failed."
      return 1
   fi
   cd $SCRIPT_DIR
   # As tests begin passing, to keep them included in future test runs,
   # they should be added here.
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestBasicOpenClose"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestDeleteNotFound"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestCloseClosed"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenOpened"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenNotFound"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenAlreadyExists"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenROClose"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenROClose4"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenROClose64"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenRWClose"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenRWClose4"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenRWClose64"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenCloseLeastFD"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenCloseDeleteMaxFD"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenOffsetEqualsZero"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenTruncate"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenAppend"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenCloseDeleteRoot"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenCloseDeleteRootMax"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestSeekErrorBadFD"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestSeekErrorBadOffsetOperation"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestSeekOffEOF"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWriteClosedFile"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWriteReadBasic"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWriteReadBasic4"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestCannotReadFromWriteOnly"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestCannotWriteToReadOnly"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWriteSomeButNotAll"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite1Byte"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite8Bytes"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite1KBytes"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite1MBytes"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite10MBytes"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite100MBytes"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestReadClosedFile"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead1ByteSimple"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead8BytesSimple"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead8BytesIter8"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead8BytesIter64"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead64BytesSimple"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead64BytesIter64K"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead64KBIter1MB"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead64KBIter10MB"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead1MBIter100MB"
	run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteReadVerfiyHoleExpansion"

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

