#!/bin/bash

PATH=/opt/homebrew/bin

mode=$($PATH/yabai -m query --spaces --space | $PATH/jq '.type')

if [ "$mode" == "\"bsp\"" ]
then
    $PATH/yabai -m space --layout float &> /dev/null
else
    $PATH/yabai -m space --layout bsp &> /dev/null
fi
