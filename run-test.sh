#!/bin/bash
set -e
    
for (( i=0; i<100; i++ ))
do
    go1.22.0 run cmd/integration_test/main.go cli  custom_commands/suggestions_preset
done
