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

run_test "TestMemoryFS_TestBasicOpenClose" 1
run_test "TestMemoryFS_TestCannotReadFromWriteOnly" 1
run_test "TestMemoryFS_TestCannotWriteToReadOnly" 1
run_test "TestMemoryFS_TestCloseClosed" 1
run_test "TestMemoryFS_TestDeleteNotFound" 1
run_test "TestMemoryFS_TestMkdir" 1
run_test "TestMemoryFS_TestMkdirAlreadyExists" 1
run_test "TestMemoryFS_TestMkdirNotFound" 1
run_test "TestMemoryFS_TestMkdirTree" 1
run_test "TestMemoryFS_TestOpenAlreadyExists" 1
run_test "TestMemoryFS_TestOpenAppend" 1
run_test "TestMemoryFS_TestOpenCloseDeleteAcrossDirectories" 1
run_test "TestMemoryFS_TestOpenCloseDeleteMaxFD" 1
run_test "TestMemoryFS_TestOpenCloseDeleteRoot" 1
run_test "TestMemoryFS_TestOpenCloseDeleteRootMax" 1
run_test "TestMemoryFS_TestOpenCloseLeastFD" 1
run_test "TestMemoryFS_TestOpenNotFound" 1
run_test "TestMemoryFS_TestOpenOffsetEqualsZero" 1
run_test "TestMemoryFS_TestOpenOpened" 1
run_test "TestMemoryFS_TestOpenROClose" 1
run_test "TestMemoryFS_TestOpenROClose4" 1
run_test "TestMemoryFS_TestOpenROClose64" 1
run_test "TestMemoryFS_TestOpenRWClose" 1
run_test "TestMemoryFS_TestOpenRWClose4" 1
run_test "TestMemoryFS_TestOpenRWClose64" 1
run_test "TestMemoryFS_TestOpenTruncate" 1
run_test "TestMemoryFS_TestReadClosedFile" 1
run_test "TestMemoryFS_TestSeekErrorBadFD" 1
run_test "TestMemoryFS_TestSeekErrorBadOffsetOperation" 1
run_test "TestMemoryFS_TestSeekOffEOF" 1
run_test "TestMemoryFS_TestWrite10MBytes" 1
run_test "TestMemoryFS_TestWrite1Byte" 1
run_test "TestMemoryFS_TestWrite1KBytes" 1
run_test "TestMemoryFS_TestWrite1MBytes" 1
run_test "TestMemoryFS_TestWrite8Bytes" 1
run_test "TestMemoryFS_TestWriteClosedFile" 1
run_test "TestMemoryFS_TestWriteRead1ByteSimple" 1
run_test "TestMemoryFS_TestWriteRead64BytesIter64K" 1
run_test "TestMemoryFS_TestWriteRead64BytesSimple" 1
run_test "TestMemoryFS_TestWriteRead64KBIter10MB" 1
run_test "TestMemoryFS_TestWriteRead64KBIter1MB" 1
run_test "TestMemoryFS_TestWriteRead8BytesIter64" 1
run_test "TestMemoryFS_TestWriteRead8BytesIter8" 1
run_test "TestMemoryFS_TestWriteRead8BytesSimple" 1
run_test "TestMemoryFS_TestWriteReadBasic" 1
run_test "TestMemoryFS_TestWriteReadBasic4" 1
