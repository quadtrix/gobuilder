#!/bin/bash
archs=""

for osarch in $(./gobuild -s);
do
  archs="${archs} -e ${osarch}"
done
./gobuild -V github.com/quadtrix/configmanager@gobuild-adapt -n ${archs}
