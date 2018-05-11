#/usr/bin/env bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
main () {
   local precheckin_script="$SCRIPT_DIR/../test/run_precheckin_tests.sh"
   local jenkins_script="$SCRIPT_DIR/../test/jenkins_test.sh"
   if [ -f $precheckin_script ]; then
      echo "Warning: removing old $precheckin_script"
   fi
   go generate
   if [ $? != 0 ]; then
      echo "go generate failed :( aborting"
      return 1
   fi
   chmod u+rwx $precheckin_script && echo "precheckin tests chmodded ok"
   cp $precheckin_script $jenkins_script && echo "jenkins script generated ok"
}
main $*
