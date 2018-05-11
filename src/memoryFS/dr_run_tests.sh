#!/usr/bin/env bash 

mkdir -p ./outfiles 
rm outfiles/*
touch outfiles/log 

iter() {
        testname="$1" 
        index="$2"
        outfile="outfiles/${testname}.${index}.out"
        touch "$outfile"
        #echo "running ${testname}.${index}"

        go test -run "$testname" > "$outfile" 2>&0

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

run_test "TestMemoryFS_TestBasicOpenClose" 0
run_test "TestMemoryFS_TestCannotReadFromWriteOnly" 0
run_test "TestMemoryFS_TestCannotWriteToReadOnly" 0
run_test "TestMemoryFS_TestCloseClosed" 0
run_test "TestMemoryFS_TestDeleteNotFound" 0
run_test "TestMemoryFS_TestMkdir" 0
run_test "TestMemoryFS_TestMkdirAlreadyExists" 0
run_test "TestMemoryFS_TestMkdirNotFound" 0
run_test "TestMemoryFS_TestMkdirTree" 0
run_test "TestMemoryFS_TestOpenAlreadyExists" 0
run_test "TestMemoryFS_TestOpenAppend" 0
run_test "TestMemoryFS_TestOpenBlockNoContention" 10
run_test "TestMemoryFS_TestOpenBlockOneWaiting" 10
run_test "TestMemoryFS_TestOpenBlockMultipleWaiting" 10
run_test "TestMemoryFS_TestOpenBlockOnlyOne" 10
run_test "TestMemoryFS_TestOpenCloseDeleteAcrossDirectories" 0
run_test "TestMemoryFS_TestOpenCloseDeleteMaxFD" 0
run_test "TestMemoryFS_TestOpenCloseDeleteRoot" 0
run_test "TestMemoryFS_TestOpenCloseDeleteRootMax" 0
run_test "TestMemoryFS_TestOpenCloseLeastFD" 0
run_test "TestMemoryFS_TestOpenNotFound" 0
run_test "TestMemoryFS_TestOpenOffsetEqualsZero" 0
run_test "TestMemoryFS_TestOpenOpened" 0
run_test "TestMemoryFS_TestOpenROClose" 0
run_test "TestMemoryFS_TestOpenROClose4" 0
run_test "TestMemoryFS_TestOpenROClose64" 0
run_test "TestMemoryFS_TestOpenRWClose" 0
run_test "TestMemoryFS_TestOpenRWClose4" 0
run_test "TestMemoryFS_TestOpenRWClose64" 0
run_test "TestMemoryFS_TestOpenTruncate" 0
run_test "TestMemoryFS_TestReadClosedFile" 0
run_test "TestMemoryFS_TestSeekErrorBadFD" 0
run_test "TestMemoryFS_TestSeekErrorBadOffsetOperation" 0
run_test "TestMemoryFS_TestSeekOffEOF" 0
run_test "TestMemoryFS_TestWrite10MBytes" 0
run_test "TestMemoryFS_TestWrite1Byte" 0
run_test "TestMemoryFS_TestWrite1KBytes" 0
run_test "TestMemoryFS_TestWrite1MBytes" 0
run_test "TestMemoryFS_TestWrite8Bytes" 0
run_test "TestMemoryFS_TestWriteClosedFile" 0
run_test "TestMemoryFS_TestWriteRead1ByteSimple" 0
run_test "TestMemoryFS_TestWriteRead64BytesIter64K" 0
run_test "TestMemoryFS_TestWriteRead64BytesSimple" 0
run_test "TestMemoryFS_TestWriteRead64KBIter10MB" 0
run_test "TestMemoryFS_TestWriteRead64KBIter1MB" 0
run_test "TestMemoryFS_TestWriteRead8BytesIter64" 0
run_test "TestMemoryFS_TestWriteRead8BytesIter8" 0
run_test "TestMemoryFS_TestWriteRead8BytesSimple" 0
run_test "TestMemoryFS_TestWriteReadBasic" 0
run_test "TestMemoryFS_TestWriteReadBasic4" 0
