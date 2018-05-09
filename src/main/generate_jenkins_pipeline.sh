#/usr/bin/env bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
main () {
   if [ "$1" == "" ]; then
      echo "usage: generate_jenkins_pipeline.sh MemoryFS"
   fi
   cd $SCRIPT_DIR/../memoryFS
   go test -run MemoryFS_TestHelpGenerateJenkinsPipeline &> /tmp/xclip-in
   xclip -sel clip < /tmp/xclip-in
   echo "it should be in paste buffer"
}
main $*
