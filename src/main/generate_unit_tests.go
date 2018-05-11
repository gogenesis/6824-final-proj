// This program generates combination_tests.go. It can be invoked by running "go generate".
//
// Also, because this program is executable, it must be in package main.
//

package main

import (
	"filesystem"
	"fmt"
	"fsraft"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type genFileParameters struct {
	fileName                string
	pkg                     string // "package" is a reserved word
	imports                 []string
	headerComment           string
	fileSystemName          string
	testNamesToMethodBodies map[string]string // "Test${fileSystemName}_" is automatically prepended to test names
}

func main() {
	fmt.Println("Generating test files.")
	//First, generate a file with a unit test for memoryFS for every functionality test.
	memoryFSTestGenParams := genFileParameters{
		fileName:                "memoryFS_test.go",
		pkg:                     "memoryFS",
		imports:                 []string{"filesystem", "testing"},
		headerComment:           "// This file contains a unit test for every functionality test (found in filesystem_tests.go).",
		fileSystemName:          "MemoryFS",
		testNamesToMethodBodies: make(map[string]string, 0),
	}

	for _, functionality := range filesystem.FunctionalityTests {
		functionalityName := GetFunctionName(functionality)
		testName := fmt.Sprintf("%v", functionalityName)
		methodBody := fmt.Sprintf("mfs := CreateEmptyMemoryFS()\n        filesystem.%v(t, &mfs)", functionalityName)
		memoryFSTestGenParams.testNamesToMethodBodies[testName] = methodBody
	}

	genTestFile(memoryFSTestGenParams)

	// Then, the file to test the Clerk, which has all combinations of functionality tests and difficulties.
	combinationTestGenParams := genFileParameters{
		fileName: "combination_test.go",
		pkg:      "fsraft",
		imports:  []string{"filesystem", "testing"},
		headerComment: `// This file contains a unit test for every combination of functionality test
// (found in filesystem_tests.go) and difficulty (found in difficulties.go).`,
		fileSystemName:          "Clerk",
		testNamesToMethodBodies: make(map[string]string, 0),
	}

	for _, difficulty := range fsraft.Difficulties {
		difficultyName := GetFunctionName(difficulty)
		for _, functionality := range filesystem.FunctionalityTests {
			functionalityName := GetFunctionName(functionality)

			testName := fmt.Sprintf("%v_%v", difficultyName, functionalityName)
			methodBody := fmt.Sprintf("runFunctionalityTestWithDifficulty(t, filesystem.%v, %v)", functionalityName, difficultyName)
			combinationTestGenParams.testNamesToMethodBodies[testName] = methodBody
		}
	}

	genTestFile(combinationTestGenParams)

	// hacky, we are just using the core test names from memoryFSTestGenParams
	genPrecheckinScript(memoryFSTestGenParams)
}

func genTestFile(params genFileParameters) {
	fmt.Printf("Generating %v\n", params.fileName)
	genFile, err := os.Create(params.fileName)
	assertNoError(err)
	defer genFile.Close()

	header := fmt.Sprintf(`// Code generated by generate_unit_tests.go. DO NOT EDIT.
%v
// Generated at %v.

package %v

`, params.headerComment, time.Now().Format("Mon Jan 2 3:04:05 PM"), params.pkg)
	genFile.Write([]byte(header))

	for _, importedPackage := range params.imports {
		genFile.Write([]byte(fmt.Sprintf("import \"%v\"\n", importedPackage)))
	}
	if len(params.imports) > 0 {
		genFile.Write([]byte("\n"))
	}

	for testName, methodBody := range params.testNamesToMethodBodies {
		genFile.Write([]byte(fmt.Sprintf(
			`func Test%v_%v(t *testing.T) {
	%v
}

`, params.fileSystemName, testName, methodBody)))
	}

	os.Rename(params.fileName, "../"+params.pkg+"/"+params.fileName)
}

func assertNoError(e error) {
	if e != nil {
		panic(e.Error())
	}
}

func genPrecheckinScript(params genFileParameters) {
	filePath := "../test/run_precheckin_tests.sh"
	fmt.Printf("Generating %v\n", filePath)
	genFile, err := os.Create(filePath)
	assertNoError(err)
	defer genFile.Close()

	genFile.Write([]byte("#!/usr/bin/env bash\n"))
	genFile.Write([]byte("\n"))
	genFile.Write([]byte("export JENKINS_FAIL=0\n"))
	genFile.Write([]byte("SCRIPT_DIR=\"$( cd \"$( dirname \"${BASH_SOURCE[0]}\" )\" && pwd )\"\n"))
	genFile.Write([]byte("mkdir -p $SCRIPT_DIR/outfiles\n"))
	genFile.Write([]byte("if [ -z $JENKINS ]; then\n"))
	genFile.Write([]byte("  rm $SCRIPT_DIR/outfiles/*\n"))
	genFile.Write([]byte("fi\n"))
	genFile.Write([]byte("touch $SCRIPT_DIR/outfiles/log\n"))
	genFile.Write([]byte("\n"))
	genFile.Write([]byte("iter() {\n"))
	genFile.Write([]byte("        testname=\"$1\" \n"))
	genFile.Write([]byte("        index=\"$2\"\n"))
	genFile.Write([]byte("        outfile=\"$SCRIPT_DIR/outfiles/${testname}.${index}.out\"\n"))
	genFile.Write([]byte("        touch \"$outfile\"\n"))
	genFile.Write([]byte("        #echo \"running ${testname}.${index}\"\n"))
	genFile.Write([]byte(""))
	genFile.Write([]byte("        if [ -z $JENKINS ]; then\n"))
	genFile.Write([]byte("           go test -run \"$testname\" > \"$outfile\" 2>&1\n"))
	genFile.Write([]byte("           exit_code=$?\n"))
	genFile.Write([]byte("        else\n"))
	genFile.Write([]byte("           $GOBIN test -run \"$testname\" > \"$outfile\" 2>&1\n"))
	genFile.Write([]byte("           exit_code=$?\n"))
	genFile.Write([]byte("        fi\n"))
	genFile.Write([]byte(""))
	genFile.Write([]byte("        if [ $exit_code == 0 ]; then\n"))
	genFile.Write([]byte("                echo \"${testname} success ${index}\" | tee -a $SCRIPT_DIR/outfiles/log\n"))
	genFile.Write([]byte("                rm \"$outfile\"\n"))
	genFile.Write([]byte("        else\n"))
	genFile.Write([]byte("                echo \"${testname} FAIL ${index}!\" | tee -a $SCRIPT_DIR/outfiles/log\n"))
	genFile.Write([]byte("                export JENKINS_FAIL=1\n"))
	genFile.Write([]byte("        fi\n"))
	genFile.Write([]byte("}\n"))
	genFile.Write([]byte("\n"))
	genFile.Write([]byte("run_test() {\n"))
	genFile.Write([]byte("        testname=\"$1\"\n"))
	genFile.Write([]byte("        quantity=\"$2\"\n"))
	genFile.Write([]byte("        let n=0 # \n"))
	genFile.Write([]byte("        #echo \"running $testname $quantity times\"\n"))
	genFile.Write([]byte("        while [ $n -lt ${quantity} ]; do\n"))
	genFile.Write([]byte("                let n++\n"))
	genFile.Write([]byte("                time iter $testname $n\n"))
	genFile.Write([]byte("        done\n"))
	genFile.Write([]byte("}\n"))
	genFile.Write([]byte("\n"))
	//@dedup when too painful to update to with new difficulties
	genFile.Write([]byte("cd $SCRIPT_DIR/../memoryFS\n"))
	genFile.Write([]byte("echo Begin Core MemoryFS Tests\n"))
	for testName, _ := range params.testNamesToMethodBodies {
		genFile.Write([]byte(fmt.Sprintf(" run_test \"Test%s_%s\" 1\n", "MemoryFS", testName)))
	}
	genFile.Write([]byte("cd $SCRIPT_DIR/../fsraft\n"))
	genFile.Write([]byte("echo Begin Raft Difficulty 1 Tests - Reliable Network - Clerk_OneClerkThreeServersNoErrors Tests\n"))
	for testName, _ := range params.testNamesToMethodBodies {
		genFile.Write([]byte(fmt.Sprintf(" run_test \"Test%s_%s\" 1\n", "Clerk_OneClerkThreeServersNoErrors", testName)))
	}
	genFile.Write([]byte("echo Begin Raft Difficulty 2 Tests - Lossy Network - Clerk_OneClerkFiveServersUnreliableNet Tests\n"))
	for testName, _ := range params.testNamesToMethodBodies {
		genFile.Write([]byte(fmt.Sprintf(" run_test \"Test%s_%s\" 1\n", "Clerk_OneClerkFiveServersUnreliableNet", testName)))
	}
	genFile.Write([]byte("echo Begin Raft Difficulty 3 Tests - Snapshots - Clerk_OneClerkFiveServersUnreliableNet Tests\n"))
	for testName, _ := range params.testNamesToMethodBodies {
		genFile.Write([]byte(fmt.Sprintf(" run_test \"Test%s_%s\" 1\n", "OneClerkThreeServersSnapshots", testName)))
	}

   // Singleton tests
	genFile.Write([]byte(" echo \"Begin Singleton Tests\"\n"))
   genFile.Write([]byte(" run_test \"TestOneClerkFiveServersPartition\" 1\n"))
   genFile.Write([]byte(" run_test \"TestKVBasic\" 1\n"))

   // must come last
	genFile.Write([]byte("if [ ! -z $JENKINS ]; then\n"))
	genFile.Write([]byte("  exit $JENKINS_FAIL\n")) //surface any fails to jenkins
	genFile.Write([]byte("fi\n"))
	fmt.Printf("Generated full precheckin script at %v\n", filePath)
}

// Get the name of a function, not including its package.
func GetFunctionName(i interface{}) string {
	nameWithPackage := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	periodIndex := strings.Index(nameWithPackage, ".")
	return nameWithPackage[periodIndex+1:]
}
