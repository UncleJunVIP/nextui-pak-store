#!/bin/zsh
adb push ./pak-store "/mnt/SDCARD/Tools/tg5040/Pak Store.pak"

adb shell rm "/mnt/SDCARD/Tools/tg5040/Pak\ Store.pak/pak-store.log" || true

printf "Pak Store has been pushed to device!"

printf "\a"
