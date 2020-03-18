#!/usr/bin/env bash
# Applies generated Swagger Server .zip into src/
ZIP_BASE="python-flask-server/"
echo "Unzipping $1 to $ZIP_BASE"
unzip -q $1
echo "Backing up controllers"
pushd src/swagger_server/controllers/ > /dev/null
for controller in *.py
do
  mv "$controller" "${controller%.py}.old.py"
done
popd > /dev/null
echo "Syncing new server code"
rsync -a --exclude-from=exclude_list.txt $ZIP_BASE src/
echo "Removing $ZIP_BASE"
rm -rf $ZIP_BASE
read -p "Remove $1? (y/n): " DELETE_CHOICE
case "$DELETE_CHOICE" in 
  y|Y ) rm $1;;
  * ) echo "Not deleted.";;
esac
echo "Done! :)"
