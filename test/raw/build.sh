export GZIP=-n9

for i in {1..3} ; do
	cd v$i && tar -czvf ../../data/v$i.tgz . && cd ..
done
