#!/bin/sh
echo "launching front end, back end and dependencies locally!"
ttab make db_boot
ttab make run_be
ttab make run_fe 
open localhost:7802/gateway