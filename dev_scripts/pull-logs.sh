#!/bin/zsh
rm ./mortar.log || true

adb pull "/mnt/SDCARD/Tools/tg5040/Mortar.pak/mortar.log" ./mortar.log

printf "All done!"

printf "\a"
