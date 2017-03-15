
for i in imagerepo nodeinfo virtd virtlogd vmangel vmshim
do
    cd $i
    ./build.sh
    cd ..
done
