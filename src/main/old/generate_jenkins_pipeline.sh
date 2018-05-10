#/usr/bin/env bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
main () {
   if [ "$1" == "--help" ]; then
      echo "usage: generate_jenkins_pipeline.sh"
      return
   fi
   cd $SCRIPT_DIR/../memoryFS
   go test -run MemoryFS_TestHelpGenerateJenkinsPipeline &> /tmp/xclip-in
   if [ $? != 0 ]; then
      echo "GENERATION FAILED:"
      cat /tmp/xclip-in
   fi
   xclip -sel clip < /tmp/xclip-in
   echo "Jenkins pipeline should be in paste buffer"
}
main $*
