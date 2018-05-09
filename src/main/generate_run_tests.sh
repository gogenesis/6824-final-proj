#/usr/bin/env bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
main () {
   if [ "$1" == "" ]; then
      echo "usage: generate_run_tests.sh <module ex: MemoryFS>"
   fi
   cd $SCRIPT_DIR/../filesystem/
   # confirm we can build and format tests
   go build && go fmt $SCRIPT_DIR/../filesystem
   grep "@test" $SCRIPT_DIR/../filesystem/filesystem_tests.go -A 50 \
      |grep -e "Test.*,$" |sed -e "s/Test/run_test\ \"Test$1_Test/" \
      |sed -e s/,/\"/
   echo "# generated `date` by $ main/generate_run_tests.sh $1"
   echo ""
   echo "WARNING! That was wicked hacky; revist soon!"
}
main $*
