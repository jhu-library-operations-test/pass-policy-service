#!/bin/sh

printf "\n**** Begin Environment Variable Dump ****\n\n"
printenv | sort
printf "\n**** End Environment Variable Dump ****\n\n"

./pass-policy-service serve ${POLICY_FILE}
