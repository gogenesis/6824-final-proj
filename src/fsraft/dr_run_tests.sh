#!/usr/bin/env bash 

export DFS_RAFT_LOG_LEVEL=0
export DFS_FSRAFT_LOG_LEVEL=2
export DFS_MEMORYFS_LOG_LEVEL=3

cd ../main
go generate
cd -
mkdir -p ./outfiles 
rm outfiles/*
touch outfiles/log 

iter() {
        testname="$1" 
        index="$2"
        outfile="outfiles/${testname}.${index}.out"
        touch "$outfile"
        #echo "running ${testname}.${index}"

        go test -run "$testname" > "$outfile" 2>&1

        exit_code="$?"
        if [ "$exit_code" = "0" ]; then
                echo "${testname} success ${index}" | tee -a outfiles/log
                rm "$outfile"
        else
                echo "${testname} FAIL ${index}!" | tee -a outfiles/log
        fi
}

run_test() {
        testname="$1" 
        quantity="$2"
        let n=0 # 
        #echo "running $testname $quantity times"
        while [ $n -lt ${quantity} ]; do
                let n++
                iter $testname $n
        done
}

run_test "TestClerk_OneClerkThreeServersNoErrors_TestBasicOpenClose" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestCannotDeleteRootDir" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestCannotReadFromWriteOnly" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestCannotWriteToReadOnly" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestCloseClosed" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestDeleteNotFound" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestHelpGenerateJenkinsPipeline" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestMkdir" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestMkdirAlreadyExists" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestMkdirNotFound" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestMkdirTree" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenAlreadyExists" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenAppend" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenCloseDeleteAcrossDirectories" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenCloseDeleteMaxFD" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenCloseDeleteRoot" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenCloseDeleteRootMax" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenCloseLeastFD" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenNotFound" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenOffsetEqualsZero" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenOpened" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenROClose" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenROClose4" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenROClose64" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenRWClose" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenRWClose4" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenRWClose64" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestOpenTruncate" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestReadClosedFile" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead1ByteSimple" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead64BytesIter64K" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead64BytesSimple" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead64KBIter1MB" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead8BytesIter64" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead8BytesIter8" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteRead8BytesSimple" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestRndWriteReadVerfiyHoleExpansion" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestSeekErrorBadFD" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestSeekErrorBadOffsetOperation" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestSeekOffEOF" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite1Byte" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite1KBytes" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite1MBytes" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestWrite8Bytes" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestWriteClosedFile" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestWriteReadBasic" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestWriteReadBasic4" 0
run_test "TestClerk_OneClerkThreeServersNoErrors_TestWriteSomeButNotAll" 1

run_test "TestClerk_OneClerkFiveServersErrors_TestBasicOpenClose" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestCannotDeleteRootDir" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestCannotReadFromWriteOnly" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestCannotWriteToReadOnly" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestCloseClosed" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestDeleteNotFound" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestHelpGenerateJenkinsPipeline" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestMkdir" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestMkdirAlreadyExists" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestMkdirNotFound" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestMkdirTree" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenAlreadyExists" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenAppend" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenCloseDeleteAcrossDirectories" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenCloseDeleteMaxFD" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenCloseDeleteRoot" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenCloseDeleteRootMax" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenCloseLeastFD" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenNotFound" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenOffsetEqualsZero" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenOpened" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenROClose" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenROClose4" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenROClose64" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenRWClose" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenRWClose4" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenRWClose64" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestOpenTruncate" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestReadClosedFile" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestRndWriteRead1ByteSimple" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestRndWriteRead64BytesIter64K" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestRndWriteRead64BytesSimple" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestRndWriteRead64KBIter1MB" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestRndWriteRead8BytesIter64" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestRndWriteRead8BytesIter8" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestRndWriteRead8BytesSimple" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestRndWriteReadVerfiyHoleExpansion" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestSeekErrorBadFD" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestSeekErrorBadOffsetOperation" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestSeekOffEOF" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestWrite1Byte" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestWrite1KBytes" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestWrite1MBytes" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestWrite8Bytes" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestWriteClosedFile" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestWriteReadBasic" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestWriteReadBasic4" 0
run_test "TestClerk_OneClerkFiveServersErrors_TestWriteSomeButNotAll" 0
