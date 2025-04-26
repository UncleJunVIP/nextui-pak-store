#!/bin/zsh
printf "Shit broke! Killing Pak Store."
sshpass -p 'tina' ssh root@192.168.1.16 "kill  \$(pidof dlv)" > /dev/null 2>&1
sshpass -p 'tina' ssh root@192.168.1.16 "kill  \$(pidof pak-store)" > /dev/null 2>&1
