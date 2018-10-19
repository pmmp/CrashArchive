#!/bin/bash

find ./volumes/reports -name "*.log" -print0 | xargs -0 sudo rm -f
