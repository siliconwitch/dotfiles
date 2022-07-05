#!/bin/bash

mode=$(yabai -m query --spaces --space | jq '.type')

if [ "$mode" == "\"bsp\"" ]
then
    yabai -m space --layout float &> /dev/null
else
    yabai -m space --layout bsp &> /dev/null
fi